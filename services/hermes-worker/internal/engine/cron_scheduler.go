package engine

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/eulerbutcooler/hermes/packages/hermes-common/pkg/cronutil"
	"github.com/eulerbutcooler/hermes/services/hermes-worker/internal/store"
	"github.com/google/uuid"
)

type CronStore interface {
	GetCronRelaysDue(ctx context.Context) ([]store.CronRelay, error)
	UpdateRelayNextRun(ctx context.Context, relayID string, nextRunAt *time.Time) error
}

type CronScheduler struct {
	store    CronStore
	jobQueue chan Job
	logger   *slog.Logger
	ticker   *time.Ticker
	done     chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

func NewCronScheduler(store CronStore, jobQueue chan Job, logger *slog.Logger) *CronScheduler {
	return &CronScheduler{
		store:    store,
		jobQueue: jobQueue,
		logger:   logger,
		done:     make(chan struct{}),
	}
}

func (cs *CronScheduler) Start(ctx context.Context) {
	cs.ticker = time.NewTicker(30 * time.Second)
	cs.logger.Info("cron scheduler started")

	cs.tick(ctx)

	cs.wg.Go(func() {
		for {
			select {
			case <-cs.done:
				cs.ticker.Stop()
				cs.logger.Info("cron scheduler stopped")
				return
			case <-ctx.Done():
				cs.ticker.Stop()
				cs.logger.Info("cron scheduler context cancelled")
				return
			case <-cs.ticker.C:
				cs.tick(ctx)
			}
		}
	})
}

func (cs *CronScheduler) Stop() {
	cs.stopOnce.Do(func() {
		close(cs.done)
	})
	cs.wg.Wait()
}

func (cs *CronScheduler) tick(ctx context.Context) {
	due, err := cs.store.GetCronRelaysDue(ctx)
	now := time.Now().UTC()
	if err != nil {
		cs.logger.Error("failed to fetch due cron relays", slog.String("error", err.Error()))
		return
	}

	for _, relay := range due {
		cs.logger.Info("firing cron relay", slog.String("relay_id", relay.ID))

		payload, err := json.Marshal(map[string]any{
			"trigger":  "cron",
			"fired_at": now.Format(time.RFC3339),
			"relay_id": relay.ID,
		})

		if err != nil {
			cs.logger.Error("failed to marshal cron payload",
				slog.String("relay_id", relay.ID),
				slog.String("error", err.Error()))
			continue
		}

		job := Job{
			RelayID: relay.ID,
			EventID: uuid.New().String(),
			Payload: payload,
			MsgAck:  func(bool) {},
		}

		// Safely attempt to send to the channel, recovering if it was closed during shutdown
		select {
		case <-cs.done:
			cs.logger.Info("cron scheduler stopping; skipping enqueue", slog.String("relay_id", relay.ID))
			return
		case cs.jobQueue <- job:
		default:
			cs.logger.Warn("job queue is full; cron relay skipped", slog.String("relay_id", relay.ID))
		}

		rawSchedule, ok := relay.TriggerConfig["schedule"]
		if !ok {
			cs.logger.Error("missing cron schedule in trigger config", slog.String("relay_id", relay.ID))
			continue
		}
		schedule, ok := rawSchedule.(string)
		if !ok || schedule == "" {
			cs.logger.Error("invalid cron schedule in trigger config",
				slog.String("relay_id", relay.ID))
			continue
		}

		nextRun, err := cronutil.ComputeNextRun(schedule, now)
		if err != nil {
			cs.logger.Error("failed to compute next run",
				slog.String("relay_id", relay.ID),
				slog.String("error", err.Error()))
			continue
		}

		if err := cs.store.UpdateRelayNextRun(ctx, relay.ID, &nextRun); err != nil {
			cs.logger.Error("failed to update next_run_at",
				slog.String("relay_id", relay.ID),
				slog.String("error", err.Error()))
		}
	}
}
