package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/dag"
	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/templateengine"
	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/store"
	"golang.org/x/sync/errgroup"
)

type Job struct {
	RelayID string
	EventID string
	Payload []byte
	MsgAck  func(bool)
}

type WorkerPool struct {
	JobQueue   chan Job
	MaxWorkers int
	Store      *store.Store
	Registry   *Registry
	Logger     *slog.Logger
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	stopOnce   sync.Once
}

// Constructor with dependency injxtn
func NewWorkerPool(maxWorkers int, db *store.Store, reg *Registry, logger *slog.Logger) *WorkerPool {
	return &WorkerPool{
		JobQueue:   make(chan Job, 100),
		MaxWorkers: maxWorkers,
		Store:      db,
		Registry:   reg,
		Logger:     logger,
	}
}

// Spawns all worker goroutines
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.ctx, wp.cancel = context.WithCancel(ctx)
	wp.Logger.Info("starting worker pool",
		slog.Int("max_workers", wp.MaxWorkers),
		slog.Int("queue_size", cap(wp.JobQueue)),
	)
	for i := 0; i < wp.MaxWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
	wp.Logger.Info("worker pool started",
		slog.Int("workers", wp.MaxWorkers))
}

// Each worker runs its own goroutine
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	workerLogger := wp.Logger.With(slog.Int("worker_id", id))
	workerLogger.Debug("worker started")

	for job := range wp.JobQueue {
		start := time.Now()
		workerLogger.Info("processing relay",
			slog.String("relay_id", job.RelayID),
			slog.String("event_id", job.EventID))

		err := wp.process(wp.ctx, job, workerLogger)
		duration := time.Since(start)
		if err != nil {
			workerLogger.Error("relay execution failed",
				slog.String("relay_id", job.RelayID),
				slog.String("event_id", job.EventID),
				slog.Duration("duration", duration),
				slog.String("error", err.Error()))
			job.MsgAck(false)
		} else {
			workerLogger.Info("relay execution succeeded",
				slog.String("relay_id", job.RelayID),
				slog.String("event_id", job.EventID),
				slog.Duration("duration", duration))
			job.MsgAck(true)
		}
	}
}

// Executes the actual workflow logic
func (wp *WorkerPool) process(ctx context.Context, job Job, logger *slog.Logger) (err error) {
	if job.EventID != "" {
		isNew, dedupeErr := wp.Store.RegisterEvent(ctx, job.RelayID, job.EventID)
		if dedupeErr != nil {
			return dedupeErr
		}
		if !isNew {
			logger.Info("duplicate event skipped",
				slog.String("relay_id", job.RelayID),
				slog.String("event_id", job.EventID))
			return nil
		}
	}

	executionID, createExecErr := wp.Store.CreateExecution(ctx, job.RelayID, job.EventID, job.Payload)
	if createExecErr != nil {
		return fmt.Errorf("create execution: %w", createExecErr)
	}

	defer func() {
		logCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		status := "success"
		details := ""
		if err != nil {
			status = "failed"
			details = err.Error()
		}

		if completeErr := wp.Store.CompleteExecution(logCtx, executionID, status, details); completeErr != nil {
			logger.Error("failed to complete execution", slog.String("error", completeErr.Error()))
		}
	}()

	userID, ownerErr := wp.Store.GetRelayOwner(ctx, job.RelayID)
	if ownerErr != nil {
		return ownerErr
	}

	actions, edges, fetchErr := wp.Store.GetRelayGraph(ctx, job.RelayID)
	if fetchErr != nil {
		return fetchErr
	}

	nodes := make([]dag.Node, len(actions))
	actionMap := make(map[string]store.RelayAction)
	for i, act := range actions {
		nodes[i] = dag.Node{ID: act.NodeID}
		actionMap[act.NodeID] = act
	}

	dagEdges := make([]dag.Edge, len(edges))
	for i, e := range edges {
		dagEdges[i] = dag.Edge{From: e.ParentNodeID, To: e.ChildNodeID, Condition: e.Condition}
	}

	graph, err := dag.New(nodes, dagEdges)
	if err != nil {
		return fmt.Errorf("invalid workflow graph: %w", err)
	}

	waves := graph.Waves()

	outputsMap := make(map[string]StepOutput)
	teOutputsMap := make(map[string]templateengine.StepOutput)
	var outputsMutex sync.Mutex

	for waveIdx, wave := range waves {
		logger.Debug("executing wave", slog.Int("wave_index", waveIdx), slog.Int("nodes", len(wave)))

		eg, waveCtx := errgroup.WithContext(ctx)

		for _, nodeID := range wave {
			nodeID := nodeID
			act := actionMap[nodeID]

			eg.Go(func() error {
				resolved, resolveErr := wp.resolveSecrets(waveCtx, userID, act.Config)
				if resolveErr != nil {
					return fmt.Errorf("action %s (%s) secret resolution failed: %w", act.ActionType, nodeID, resolveErr)
				}

				outputsMutex.Lock()
				teMapCopy := make(map[string]templateengine.StepOutput)
				for k, v := range teOutputsMap {
					teMapCopy[k] = v
				}
				outputsMutex.Unlock()

				resolved = templateengine.Resolve(resolved, job.Payload, teMapCopy)

				stepID, stepCreateErr := wp.Store.CreateExecutionStep(waveCtx, executionID, nodeID, act.ActionType, redactConfig(act.Config, resolved))
				if stepCreateErr != nil {
					return fmt.Errorf("create step for %s: %w", nodeID, stepCreateErr)
				}

				logger.Debug("executing action", slog.String("node_id", nodeID), slog.String("action_type", act.ActionType))

				executor, pluginErr := wp.Registry.Get(act.ActionType)
				if pluginErr != nil {
					_ = wp.Store.CompleteExecutionStep(context.Background(), stepID, "failed", pluginErr.Error(), nil)
					return pluginErr
				}

				output, execErr := executor.Execute(waveCtx, resolved, job.Payload, outputsMap)
				if execErr != nil {
					_ = wp.Store.CompleteExecutionStep(context.Background(), stepID, "failed", execErr.Error(), nil)
					return fmt.Errorf("action %s (%s) failed: %w", act.ActionType, nodeID, execErr)
				}

				if completeErr := wp.Store.CompleteExecutionStep(context.Background(), stepID, "success", "", output); completeErr != nil {
					logger.Error("failed to complete execution step", slog.String("step_id", stepID), slog.String("error", completeErr.Error()))
				}

				outputsMutex.Lock()
				outputsMap[nodeID] = StepOutput{
					ActionType: act.ActionType,
					NodeID:     nodeID,
					Output:     output,
				}
				teOutputsMap[nodeID] = templateengine.StepOutput{
					ActionType: act.ActionType,
					NodeID:     nodeID,
					Output:     json.RawMessage(output),
				}
				outputsMutex.Unlock()

				return nil
			})
		}

		if waveErr := eg.Wait(); waveErr != nil {
			return waveErr
		}
	}

	return nil
}

func (wp *WorkerPool) resolveSecrets(ctx context.Context, userID string, config map[string]any) (map[string]any, error) {
	resolved := make(map[string]any, len(config))
	for k, v := range config {
		if strings.HasSuffix(k, "_ref") {
			secretName, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("secret ref %q must be a string", k)
			}
			if strings.TrimSpace(secretName) == "" {
				continue
			}
			actualKey := strings.TrimSuffix(k, "_ref")
			value, err := wp.Store.ResolveSecret(ctx, userID, secretName)
			if err != nil {
				return nil, fmt.Errorf("resolve %q: %w", secretName, err)
			}
			resolved[actualKey] = value
		} else {
			resolved[k] = v
		}
	}
	return resolved, nil
}

func redactConfig(original, resolved map[string]any) map[string]any {
	alwaysSensitive := map[string]struct{}{
		"webhook_url": {},
		"api_key":     {},
		"token":       {},
		"password":    {},
		"secret":      {},
	}

	secretKeys := make(map[string]struct{})
	for k := range original {
		if before, ok := strings.CutSuffix(k, "_ref"); ok {
			secretKeys[before] = struct{}{}
		}
	}

	safe := make(map[string]any, len(resolved))
	for k, v := range resolved {
		_, isSecretRef := secretKeys[k]
		_, isAlways := alwaysSensitive[k]
		if isSecretRef || isAlways {
			safe[k] = "[redacted]"
		} else {
			safe[k] = v
		}
	}
	return safe
}

func (wp *WorkerPool) Shutdown() {
	wp.stopOnce.Do(func() {
		wp.Logger.Info("Initializing worker pool shutdown")
		close(wp.JobQueue)
		wp.wg.Wait()
		if wp.cancel != nil {
			wp.cancel()
		}
		wp.Logger.Info("Worker pool shutdown complete")
	})
}
