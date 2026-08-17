package messaging

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/LucasLuis-Dev/Routeforge/backend-go/domain"
)

type OutboxProcessor struct {
	outboxRepo     domain.OutboxRepository
	eventPublisher EventPublisher
	interval       time.Duration
}

func NewOutboxProcessor(outboxRepo domain.OutboxRepository, eventPublisher EventPublisher, interval time.Duration) *OutboxProcessor {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	return &OutboxProcessor{
		outboxRepo:     outboxRepo,
		eventPublisher: eventPublisher,
		interval:       interval,
	}
}

func (p *OutboxProcessor) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	log.Printf("Transactional Outbox Worker iniciado (verificação a cada %v)", p.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Transactional Outbox Worker encerrado.")
			return
		case <-ticker.C:
			p.processPendingEvents(ctx)
		}
	}
}

func (p *OutboxProcessor) processPendingEvents(ctx context.Context) {
	events, err := p.outboxRepo.GetPendingEvents(ctx, 50)
	if err != nil {
		log.Printf("Erro ao buscar eventos pendentes na Outbox: %v", err)
		return
	}

	if len(events) == 0 {
		return
	}

	for _, event := range events {
		var payload interface{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			log.Printf("Erro ao deserializar payload do evento outbox %s: %v", event.ID, err)
			continue
		}

		if p.eventPublisher != nil {
			if err := p.eventPublisher.PublishEvent(ctx, event.RoutingKey, payload); err != nil {
				log.Printf("Falha ao publicar evento da Outbox %s no RabbitMQ (RoutingKey: %s): %v", event.ID, event.RoutingKey, err)
				continue
			}
		}

		if err := p.outboxRepo.MarkProcessed(ctx, event.ID); err != nil {
			log.Printf("Erro ao marcar evento %s como processado na Outbox: %v", event.ID, err)
		} else {
			log.Printf("Evento da Outbox %s (key: %s) processado e publicado com sucesso no RabbitMQ!", event.ID, event.RoutingKey)
		}
	}
}
