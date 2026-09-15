package forwardauth

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/segmentio/kafka-go"
)

func NewRevocationConsumer(queueURL, destination, group string) (RevocationConsumer, error) {
	parsed, err := url.Parse(strings.TrimSpace(queueURL))
	if err != nil {
		return nil, fmt.Errorf("parse QUEUE_URL: %w", err)
	}
	switch strings.ToLower(parsed.Scheme) {
	case "amqp", "amqps":
		return newRabbitConsumer(queueURL, destination)
	case "kafka":
		brokers := strings.Split(parsed.Host, ",")
		if len(brokers) == 0 || strings.TrimSpace(brokers[0]) == "" {
			return nil, fmt.Errorf("QUEUE_URL contains no Kafka broker")
		}
		return &kafkaConsumer{reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers, Topic: destination, GroupID: group, MinBytes: 1, MaxBytes: 1 << 20,
		})}, nil
	default:
		return nil, fmt.Errorf("unsupported QUEUE_URL scheme %q", parsed.Scheme)
	}
}

type rabbitConsumer struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	queue      string
}

func newRabbitConsumer(endpoint, queue string) (*rabbitConsumer, error) {
	connection, err := amqp.Dial(endpoint)
	if err != nil {
		return nil, fmt.Errorf("connect RabbitMQ: %w", err)
	}
	channel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("open RabbitMQ channel: %w", err)
	}
	if _, err := channel.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		_ = channel.Close()
		_ = connection.Close()
		return nil, fmt.Errorf("declare RabbitMQ queue: %w", err)
	}
	if err := channel.Qos(64, 0, false); err != nil {
		_ = channel.Close()
		_ = connection.Close()
		return nil, fmt.Errorf("configure RabbitMQ QoS: %w", err)
	}
	return &rabbitConsumer{connection: connection, channel: channel, queue: queue}, nil
}

func (c *rabbitConsumer) Run(ctx context.Context, handle func(context.Context, []byte) error) error {
	deliveries, err := c.channel.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume RabbitMQ queue: %w", err)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, open := <-deliveries:
			if !open {
				return fmt.Errorf("RabbitMQ delivery channel closed")
			}
			if err := handle(ctx, delivery.Body); err != nil {
				_ = delivery.Nack(false, !IsPermanentEventError(err))
				continue
			}
			if err := delivery.Ack(false); err != nil {
				return fmt.Errorf("ack RabbitMQ event: %w", err)
			}
		}
	}
}

func (c *rabbitConsumer) Close() error {
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.connection != nil {
		return c.connection.Close()
	}
	return nil
}

type kafkaConsumer struct{ reader *kafka.Reader }

func (c *kafkaConsumer) Run(ctx context.Context, handle func(context.Context, []byte) error) error {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil && ctx.Err() != nil {
			return nil
		}
		if err != nil {
			return fmt.Errorf("fetch Kafka event: %w", err)
		}
		for {
			err = handle(ctx, message.Value)
			if err == nil || IsPermanentEventError(err) {
				break
			}
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(100 * time.Millisecond):
			}
		}
		if err := c.reader.CommitMessages(ctx, message); err != nil {
			return fmt.Errorf("commit Kafka event: %w", err)
		}
	}
}

func (c *kafkaConsumer) Close() error { return c.reader.Close() }
