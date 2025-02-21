package rabbitmq

import (
	"context"

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

func RegisterConsumer(h consumerHandler, routingKey string) {
	registry.Consumers = append(registry.Consumers, rabbitConsumer{Handler: h, RoutingKey: routingKey})
}

func initializeConsumers(ctx context.Context, channel *amqp091.Channel) error {
	for _, consumer := range registry.Consumers {
		log.Debug().Msgf("Setting up consumer for routing key %s", consumer.RoutingKey)

		q, err := env.Channel.QueueDeclare(
			consumer.RoutingKey, // name
			false,               // durable
			false,               // delete when unused
			false,               // exclusive
			false,               // no-wait
			nil,                 // arguments
		)
		if err != nil {
			return err
		}
		err = env.Channel.QueueBind(
			q.Name,              // queue name
			consumer.RoutingKey, // routing key
			// TODO: make a constant
			"mediapire-exch", // exchange
			false,
			nil)
		if err != nil {
			return err
		}

		msgs, err := env.Channel.Consume(
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

		go func(c rabbitConsumer) {
			for d := range msgs {
				log.Debug().Msgf("Handling message for routing key %s", c.RoutingKey)

				c.Handler(context.Background(), d)
			}
		}(consumer)

	}

	return nil
}
