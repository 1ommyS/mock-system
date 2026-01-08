package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"dsl-runner-svc/internal/application"
	"dsl-runner-svc/internal/domain"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	Reader  *kafka.Reader
	Handler *application.JobHandler
	Logger  *slog.Logger
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		msg, err := c.Reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			c.Logger.Error("kafka fetch failed", "error", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if err := c.handleMessage(ctx, msg); err != nil {
			if application.IsTransient(err) {
				c.Logger.Warn("job transient error", "error", err)
				continue
			}
			c.Logger.Error("job failed", "error", err)
		}
		if err := c.Reader.CommitMessages(ctx, msg); err != nil {
			c.Logger.Error("commit failed", "error", err)
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, msg kafka.Message) error {
	var payload domain.JobMessage
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		return err
	}
	c.Logger.Info("job received", "job_id", payload.JobID, "job_key", payload.JobKey, "mode", payload.Mode)
	return c.Handler.Process(ctx, payload)
}
