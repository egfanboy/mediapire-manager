package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/rabbitmq/amqp091-go"
)

type connectionEnv struct {
	Connection *amqp091.Connection
	Channel    *amqp091.Channel
}

const (
	connectionString = "amqp://%s:%s@%s:%d/"
)

var (
	env = connectionEnv{}
)

func Setup(ctx context.Context) error {
	rabbitCfg := app.GetApp().Config.Rabbit
	var err error
	env.Connection, err = amqp091.Dial(fmt.Sprintf(connectionString, rabbitCfg.Username, rabbitCfg.Password, rabbitCfg.Address, rabbitCfg.Port))
	if err != nil {
		return err
	}

	env.Channel, err = env.Connection.Channel()
	if err != nil {
		return err
	}

	err = env.Channel.ExchangeDeclare(
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
		return err
	}

	return initializeConsumers(ctx, env.Channel)
}

func PublishMessage(ctx context.Context, routingKey string, messageBody interface{}) error {
	body, err := json.Marshal(messageBody)
	if err != nil {
		return err
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
