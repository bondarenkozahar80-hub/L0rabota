package kafka

import (
	"context"
	"log"
	"order-service0/internal/usecase"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

type OrderConsumer struct {
	reader       *kafka.Reader
	orderUseCase usecase.OrderUseCase
}

func NewOrderConsumer(brokers []string, topic, groupID string, orderUseCase usecase.OrderUseCase) *OrderConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &OrderConsumer{
		reader:       reader,
		orderUseCase: orderUseCase,
	}
}

func (c *OrderConsumer) Start(ctx context.Context) {
	log.Println("Starting Kafka consumer...")
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka consumer...")
			return
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				log.Printf("Error fetching message: %v", err)
				continue
			}

			if err := c.orderUseCase.ProcessOrderMessage(ctx, msg.Value); err != nil {
				log.Printf("Error processing order message: %v", err)
				continue
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Error committing message: %v", err)
			}
		}
	}
}

func (c *OrderConsumer) Close() error {
	return c.reader.Close()
}
