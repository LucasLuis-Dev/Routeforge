package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeRouteforgeEvents = "routeforge_events"
)

type EventPublisher interface {
	PublishEvent(ctx context.Context, routingKey string, payload interface{}) error
	Close() error
}

type RabbitMQPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.Mutex
}

func NewRabbitMQPublisher(amqpURL string) (EventPublisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar no RabbitMQ (%s): %w", amqpURL, err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("erro ao abrir canal do RabbitMQ: %w", err)
	}

	// Declara Topic Exchange durável para eventos do ecossistema
	err = ch.ExchangeDeclare(
		ExchangeRouteforgeEvents, // name
		"topic",                  // type
		true,                     // durable
		false,                    // auto-deleted
		false,                    // internal
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("erro ao declarar Exchange '%s': %w", ExchangeRouteforgeEvents, err)
	}

	log.Println("RabbitMQ Event Publisher iniciado e pronto para emitir eventos de domínio!")

	return &RabbitMQPublisher{
		conn:    conn,
		channel: ch,
	}, nil
}

func (p *RabbitMQPublisher) PublishEvent(ctx context.Context, routingKey string, payload interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar evento JSON: %w", err)
	}

	pubCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = p.channel.PublishWithContext(
		pubCtx,
		ExchangeRouteforgeEvents, // exchange
		routingKey,               // routing key (ex: ride.requested, ride.accepted)
		false,                    // mandatory
		false,                    // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Timestamp:    time.Now(),
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)

	if err != nil {
		return fmt.Errorf("erro ao publicar evento no RabbitMQ (key: %s): %w", routingKey, err)
	}

	log.Printf("Evento publicado no RabbitMQ: [%s] -> %s", routingKey, string(body))
	return nil
}

func (p *RabbitMQPublisher) Close() error {
	if p.channel != nil {
		_ = p.channel.Close()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
	return nil
}
