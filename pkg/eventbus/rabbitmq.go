package eventbus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type RabbitMQEventBus struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	log  *zap.Logger
}

func NewRabbitMQEventBus(ctx context.Context, uri string, log *zap.Logger) (*RabbitMQEventBus, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		conn, err := amqp.Dial(uri)
		if err != nil {
			return nil, err
		}
		ch, err := conn.Channel()
		if err != nil {
			conn.Close()
			return nil, err
		}
		return &RabbitMQEventBus{
			conn: conn,
			ch:   ch,
			log:  log,
		}, nil
	}
}

// Publish publishes a message to a specific topic
func (bus *RabbitMQEventBus) Publish(ctx context.Context, topic string, message interface{}) error {
	if bus.ch == nil {
		return errors.New("channel is not initialized")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		body, err := json.Marshal(message)
		if err != nil {
			return err
		}

		return bus.ch.Publish(
			"",    // default exchange
			topic, // routing key used as queue name
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			},
		)
	}
}

// Subscribe subscribes to messages on a given topic and calls the provided handler
func (bus *RabbitMQEventBus) Subscribe(ctx context.Context, topic string, handler func(message []byte) error) error {
	if bus.ch == nil {
		return errors.New("channel is not initialized")
	}
	if handler == nil {
		return errors.New("handler function cannot be nil")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		q, err := bus.ch.QueueDeclare(
			topic, // topic queue
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return err
		}

		msgs, err := bus.ch.Consume(
			q.Name, // queue
			"",     // consumer
			true,   // autoAck
			false,  // exclusive
			false,  // no local
			false,  // no wait
			nil,
		)
		if err != nil {
			return err
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case d, ok := <-msgs:
					if !ok {
						return
					}
					if err := handler(d.Body); err != nil {
						bus.log.Error("error handling message", zap.Error(err))
					}
				}
			}
		}()
		return nil
	}
}

// Close closes the RabbitMQ connection and channel
func (bus *RabbitMQEventBus) Close() error {
	var err error
	if bus.ch != nil {
		if cerr := bus.ch.Close(); cerr != nil {
			err = fmt.Errorf("failed to close channel: %w", cerr)
		}
	}
	if bus.conn != nil {
		if cerr := bus.conn.Close(); cerr != nil {
			if err != nil {
				err = fmt.Errorf("%v, failed to close connection: %w", err, cerr)
			} else {
				err = fmt.Errorf("failed to close connection: %w", cerr)
			}
		}
	}
	return err
}
