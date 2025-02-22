package rabbitmq

import (
	"context"
	"fmt"
	"sync"

	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/rs/zerolog/log"

	"github.com/rabbitmq/amqp091-go"
)

type consumerHandler func(ctx context.Context, msg amqp091.Delivery)

type rabbitConsumer struct {
	Handler    consumerHandler
	RoutingKey string
}

type consumerRegistry struct {
	Consumers []rabbitConsumer
}

var (
	registry = consumerRegistry{Consumers: []rabbitConsumer{}}
)

type consummerMapping struct {
	Mu   sync.Mutex
	Data map[string]consumerHandler
}

var cm = &consummerMapping{Mu: sync.Mutex{}, Data: map[string]consumerHandler{}}

func RegisterConsumer(h consumerHandler, routingKey string) {
	cm.Mu.Lock()
	defer cm.Mu.Unlock()
	cm.Data[routingKey] = h
}

func initializeConsumers(ctx context.Context, channel *amqp091.Channel) error {
	q, err := channel.QueueDeclare(
		fmt.Sprintf("mediapire-manager-%s", app.GetApp().Config.Name), // name
		true,  // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	for routingKey := range cm.Data {
		log.Debug().Msgf("Setting up consumer for routing key %s", routingKey)

		err = channel.QueueBind(
			q.Name,     // queue name
			routingKey, // routing key
			// TODO: make a constant
			"mediapire-exch", // exchange
			false,
			nil)
		if err != nil {
			return err
		}

	}

	msgs, err := channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			msg.Ack(false)

			log.Debug().Msgf("Handling message for routing key %s", msg.RoutingKey)

			cm.Mu.Lock()

			if handler, ok := cm.Data[msg.RoutingKey]; !ok {
				log.Debug().Msgf("No handler registered for routing key %s. Message acknowledge but no action taken", msg.RoutingKey)
			} else {
				go handler(context.Background(), msg)
			}

			cm.Mu.Unlock()
		}
	}()

	return nil
}
