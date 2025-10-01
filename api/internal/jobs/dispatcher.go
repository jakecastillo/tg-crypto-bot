package jobs

import (
    "context"
    "encoding/json"
    "time"

    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"
    "github.com/rs/zerolog"
)

// Dispatcher publishes intents to Redis streams for the exec service.
type Dispatcher struct {
    redis      *redis.Client
    streamName string
    logger     zerolog.Logger
    ttl        time.Duration
}

// NewDispatcher constructs a Dispatcher.
func NewDispatcher(redis *redis.Client, stream string, ttl time.Duration, logger zerolog.Logger) *Dispatcher {
    return &Dispatcher{redis: redis, streamName: stream, logger: logger, ttl: ttl}
}

// Intent payload for trade execution.
type Intent struct {
    ID        string          `json:"id"`
    Principal string          `json:"principal"`
    Payload   json.RawMessage `json:"payload"`
    CreatedAt time.Time       `json:"created_at"`
}

// Publish emits a new job to the stream.
func (d *Dispatcher) Publish(ctx context.Context, principal string, payload interface{}) (string, error) {
    raw, err := json.Marshal(payload)
    if err != nil {
        return "", err
    }
    intent := Intent{
        ID:        uuid.NewString(),
        Principal: principal,
        Payload:   raw,
        CreatedAt: time.Now().UTC(),
    }
    encoded, err := json.Marshal(intent)
    if err != nil {
        return "", err
    }
    args := &redis.XAddArgs{
        Stream: d.streamName,
        Values: map[string]interface{}{"intent": encoded},
    }
    if err := d.redis.XAdd(ctx, args).Err(); err != nil {
        return "", err
    }
    d.logger.Info().Str("intent_id", intent.ID).Msg("published trade intent")
    return intent.ID, nil
}
