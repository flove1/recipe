package database

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

type EventBus struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisEventBus initializes a new RedisEventBus
func NewRedisEventBus(redisClient *redis.Client) *EventBus {
	return &EventBus{
		client: redisClient,
		ctx:    context.Background(),
	}
}

func (bus *EventBus) Publish(topic string, message string) error {
	err := bus.client.Publish(bus.ctx, topic, message).Err()
	if err != nil {
		log.Printf("error publishing message: %v", err)
	}
	return err
}

func (bus *EventBus) Subscribe(topic string, handler func(message string)) {
	go func() {
		subscriber := bus.client.Subscribe(bus.ctx, topic)
		defer subscriber.Close()

		for msg := range subscriber.Channel() {
			handler(msg.Payload)
		}
	}()
}
