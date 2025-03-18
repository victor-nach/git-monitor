package eventbus

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"go.uber.org/zap"
)

type InMemoryEventBus struct {
	subscribers sync.Map 
	log         *zap.Logger
	bufferSize  int 
}

func NewInMemoryEventBus(log *zap.Logger, bufferSize int) *InMemoryEventBus {
	return &InMemoryEventBus{
		subscribers: sync.Map{},
		log:         log,
		bufferSize:  bufferSize,
	}
}

// Publish publishes a message to a specific topic
func (bus *InMemoryEventBus) Publish(ctx context.Context, topic string, message interface{}) error {
	log := bus.log.With(zap.String("topic", topic))
	log.Info("publishing event..", zap.Any("topic", topic))
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		body, err := json.Marshal(message)
		if err != nil {
			return err
		}

		// Load the subscribers for the topic
		if chs, ok := bus.subscribers.Load(topic); ok {
			for _, ch := range chs.([]chan []byte) {
				select {
				case ch <- body:
				default:
					bus.log.Warn("subscriber channel is full, dropping message", zap.String("topic", topic))
				}
			}
		}

		return nil
	}
}

// Subscribe subscribes to messages on a given topic and calls the provided handler
func (bus *InMemoryEventBus) Subscribe(ctx context.Context, topic string, handler func(message []byte) error) error {
	if handler == nil {
		return errors.New("handler function cannot be nil")
	}

	ch := make(chan []byte, bus.bufferSize)

	chs, _ := bus.subscribers.LoadOrStore(topic, []chan []byte{})
	updatedChs := append(chs.([]chan []byte), ch)
	bus.subscribers.Store(topic, updatedChs)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				if err := handler(msg); err != nil {
					bus.log.Error("error handling message", zap.Error(err))
				}
			}
		}
	}()

	return nil
}

func (bus *InMemoryEventBus) Close() error {
	bus.subscribers.Range(func(topic, chs interface{}) bool {
		for _, ch := range chs.([]chan []byte) {
			close(ch)
		}
		bus.subscribers.Delete(topic)
		return true
	})

	return nil
}
