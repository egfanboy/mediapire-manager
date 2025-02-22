package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

const (
	connectionString = "amqp://%s:%s@%s:%d/"
	maxRetries       = 5
)

type connectionEnv struct {
	Connection *amqp091.Connection
	Channel    *amqp091.Channel
	ErrorsCh   chan *amqp091.Error
}

func observeChannelError(ctx context.Context, ce *connectionEnv) {
	err, ok := <-ce.ErrorsCh
	if !ok {
		return
	}

	log.Err(err).Msg("Channel was closed. Attempting to reconnect.")

	ce.ConnectToChan(ctx)
}

func (ce *connectionEnv) ConnectToChan(ctx context.Context) error {
	if ce.Connection == nil {
		rabbitCfg := app.GetApp().Config.Rabbit
		conn, err := amqp091.Dial(fmt.Sprintf(connectionString, rabbitCfg.Username, rabbitCfg.Password, rabbitCfg.Address, rabbitCfg.Port))
		if err != nil {
			log.Err(err).Msg("Failed to connect to rabbitmq instance")
			return err
		}

		ce.Connection = conn
	}

	ch, err := env.Connection.Channel()
	if err != nil {
		log.Err(err).Msg("Failed to connect to rabbitmq channel")
		return err
	}

	ce.Channel = ch

	ce.ErrorsCh = ce.Channel.NotifyClose(make(chan *amqp091.Error))

	go observeChannelError(ctx, ce)

	err = ce.Channel.ExchangeDeclare(
		// TODO: make it a constant
		"mediapire-exch", // name
		"topic",          // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		log.Err(err).Msg("Failed to declare exchange on rabbitmq channel")
		return err
	}

	log.Info().Msg("Successfully connected to rabbitmq channel")

	return initializeConsumers(ctx, ce.Channel)
}

var (
	env = &connectionEnv{}
)

func Setup(ctx context.Context) error {
	err := env.ConnectToChan(ctx)
	if err != nil {
		log.Err(err).Msg("Failed to connect to rabbitmq")
		return err
	}

	return nil
}

func PublishMessage(ctx context.Context, routingKey string, messageBody interface{}) error {
	body, err := json.Marshal(messageBody)
	if err != nil {
		return err
	}

	for i := 0; i < maxRetries; i++ {
		if env.Channel.IsClosed() {
			// last iteration
			if i+1 == maxRetries {
				log.Error().Msgf("channel was never opened, cannot send message for routing key %s", routingKey)
				return errors.New("channel was not opened to send message")
			}
			log.Debug().Msgf("Channel is closed, waiting for it to open. Retry %d out of %d", i+1, maxRetries)
			time.Sleep(time.Second * 1)

		} else {
			break
		}
	}

	return env.Channel.PublishWithContext(ctx, "mediapire-exch", routingKey, false, false, amqp091.Publishing{
		ContentType: "text/plain",
		Body:        body,
	})
}

func Cleanup() {
	if env.Channel != nil {
		env.Channel.Close()
	}

	if env.Connection != nil {
		env.Connection.Close()
	}
}
