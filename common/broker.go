package common

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lormars/octohunter/internal/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	name string
}

var OutputP = Producer{name: "dork_broker"}
var CnameP = Producer{name: "cname_broker"}
var RedirectP = Producer{name: "redirect_broker"}
var MethodP = Producer{name: "method_broker"}
var HopP = Producer{name: "hopper_broker"}
var (
	conn *amqp.Connection
	ch   *amqp.Channel
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func Init() {
	var err error
	user := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")
	connStr := fmt.Sprintf("amqp://%s:%s@localhost:5672/", user, password)
	conn, err = amqp.Dial(connStr)
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err = conn.Channel()
	failOnError(err, "Failed to open a channel")
	DeclareQueue("dork_broker")
	DeclareQueue("cname_broker")
	DeclareQueue("redirect_broker")
	DeclareQueue("method_broker")
	DeclareQueue("hopper_broker")

}

func DeclareQueue(name string) {
	_, err := ch.QueueDeclare(
		name,
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

}

func (p Producer) PublishMessage(body string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ch.PublishWithContext(
		ctx,
		"",
		p.name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")
	logger.Debugf(" [x] Sent %s to %s", body, p.name)
}

func (p Producer) ConsumeMessage(f Atomic, opts *Opts) {
	msgs, err := ch.Consume(
		p.name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")
	var forever = make(chan struct{})
	go func() {
		for d := range msgs {
			localOpts := &Opts{
				Module:         opts.Module,
				Target:         string(d.Body),
				DorkFile:       opts.DorkFile,
				HopperFile:     opts.HopperFile,
				MethodFile:     opts.MethodFile,
				RedirectFile:   opts.RedirectFile,
				CnameFile:      opts.CnameFile,
				DispatcherFile: opts.DispatcherFile,
			}
			logger.Debugf("Producer %s Received a message: %s\n", p.name, d.Body)
			f(localOpts)
		}

	}()
	logger.Infoln("Waiting for messages: ", p.name)
	<-forever

}

func Close() {
	if ch != nil {
		ch.Close()
	}
	if conn != nil {
		conn.Close()
	}
}
