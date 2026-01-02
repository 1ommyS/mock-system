package application

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"
)

type OutboxWorker struct {
	Repo      OutboxRepository
	Auth      AuthClient
	Interval  time.Duration
	BatchSize int
}

func (w *OutboxWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *OutboxWorker) process(ctx context.Context) {
	events, err := w.Repo.ListDue(ctx, w.BatchSize)
	if err != nil {
		slog.Error("outbox list failed", "error", err)
		return
	}
	for _, event := range events {
		if err := w.handleEvent(ctx, event.EventType, event.Payload); err != nil {
			slog.Warn("outbox event failed", "event_id", event.ID, "event_type", event.EventType, "error", err)
			delay := 5 * (event.Attempts + 1)
			_ = w.Repo.MarkAttempt(ctx, event.ID, delay, err.Error())
			continue
		}
		_ = w.Repo.MarkDone(ctx, event.ID)
	}
}

func (w *OutboxWorker) handleEvent(ctx context.Context, eventType string, payload []byte) error {
	switch eventType {
	case "register_mock":
		var data struct {
			MockID      string `json:"mockId"`
			OwnerUserID string `json:"ownerUserId"`
		}
		if err := json.Unmarshal(payload, &data); err != nil {
			return err
		}
		if data.MockID == "" || data.OwnerUserID == "" {
			return ErrInvalidRequest
		}
		return w.Auth.RegisterResource(ctx, "mock", data.MockID, data.OwnerUserID)
	default:
		return ErrInvalidRequest
	}
}
