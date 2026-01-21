package engine

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/store"
)

type Job struct {
	RelayID string
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

// Spaws all worker goroutines
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.ctx, wp.cancel = context.WithCancel(ctx)
	wp.Logger.Info("starting worker pool",
		slog.Int("max_workers", wp.MaxWorkers),
		slog.Int("queue_size", cap(wp.JobQueue)),
	)
	for i := 0; i < wp.MaxWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
	wp.Logger.Info("worker pool started",
		slog.Int("workers", wp.MaxWorkers))
}

// Each worker runs its own goroutine
func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()
	workerLogger := wp.Logger.With(slog.Int("worker_id", id))
	workerLogger.Debug("worker started")
	for {
		select {
		case <-wp.ctx.Done():
			workerLogger.Info("worker shutting down")
			return
		case job := <-wp.JobQueue:
			start := time.Now()
			workerLogger.Info("processing relay", slog.String("relay_id", job.RelayID))
			err := wp.process(wp.ctx, job, workerLogger)
			duration := time.Since(start)
			if err != nil {
				workerLogger.Error("relay execution failed", slog.String("relay_id", job.RelayID),
					slog.Duration("duration", duration), slog.String("error", err.Error()))
				job.MsgAck(false)
			} else {
				workerLogger.Info("relay execution succeeded", slog.String("relay_id", job.RelayID), slog.Duration("duration", duration))
				job.MsgAck(true)
			}
		}
	}
}

// Executes the actual workflow logic
func (wp *WorkerPool) process(ctx context.Context, job Job, logger *slog.Logger) (err error) {
	status := "success"
	details := "Relay executed successfully"
	defer func() {
		if err != nil {
			status = "failed"
			details = err.Error()
		}
		logErr := wp.Store.LogExecution(context.Background(), job.RelayID, status, details)
		if logErr != nil {
			logger.Error("failed to save execution log", slog.String("error", logErr.Error()))
		}
	}()
	logger.Debug("fetching relay instructions", slog.String("relay_id", job.RelayID))

	instruction, fetchErr := wp.Store.GetRelayInstructions(ctx, job.RelayID)
	if fetchErr != nil {
		return fetchErr
	}

	logger.Debug("fetched relay instructions", slog.String("action_type", instruction.ActionType))

	executor, pluginErr := wp.Registry.Get(instruction.ActionType)
	if pluginErr != nil {
		return pluginErr
	}
	logger.Debug("execution instruction", slog.String("action_type", instruction.ActionType))
	return executor.Execute(ctx, instruction.Config, job.Payload)
}

func (wp *WorkerPool) Shutdown() {
	wp.Logger.Info("Initializing worker pool shutdown")

	if wp.cancel != nil {
		wp.cancel()
	}
	close(wp.JobQueue)
	wp.wg.Wait()
	wp.Logger.Info("Worker pool shutdown complete")
}
