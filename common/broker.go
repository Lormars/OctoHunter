package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
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
var DividerP = Producer{name: "divider_broker"}
var CrawlP = Producer{name: "crawl_broker"}

var (
	conn  *amqp.Connection
	ch    *amqp.Channel
	mutex sync.Mutex
)
var queueNames = []string{
	"dork_broker", "cname_broker", "redirect_broker",
	"method_broker", "hopper_broker", "divider_broker", "crawl_broker",
}

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

	err = ch.Qos(
		100,   // prefetch count
		0,     // prefetch size
		false, // global
	)

	failOnError(err, "Failed to set QoS")

	for _, name := range queueNames {
		DeclareQueue(name)
	}

}

func DeclareQueue(name string) {
	_, err := ch.QueuePurge(name, false)
	failOnError(err, "Failed to purge a queue")
	_, err = ch.QueueDeclare(
		name,
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

}

func (p Producer) PublishMessage(body interface{}) {
	var messageBody []byte
	var contentType string
	var err error

	for {
		mutex.Lock()
		queueInfo, err := ch.QueueDeclarePassive(p.name, false, false, false, false, nil)
		mutex.Unlock()
		failOnError(err, "Failed to inspect a queue")
		logger.Debugf("Queue %s has %d messages ready", p.name, queueInfo.Messages)
		if queueInfo.Messages < 100 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch v := body.(type) {
	case string:
		messageBody = []byte(v)
		contentType = "text/plain"
	case *ServerResult:
		messageBody, err = json.Marshal(v)
		if err != nil {
			failOnError(err, "Failed to marshal struct to JSON")
		}
		contentType = "application/json"
	default:
		failOnError(fmt.Errorf("unknown type %T", v), "Failed to publish a message")

	}
	mutex.Lock()
	err = ch.PublishWithContext(
		ctx,
		"",
		p.name,
		false,
		false,
		amqp.Publishing{
			ContentType: contentType,
			Body:        messageBody,
		})
	mutex.Unlock()
	failOnError(err, "Failed to publish a message")
	logger.Debugf(" [x] Sent to %s", p.name)
}

func (p Producer) ConsumeMessage(handlerFunc interface{}, opts *Opts) {
	msgs, err := ch.Consume(
		p.name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")

	var forever = make(chan struct{})
	go func() {
		for d := range msgs {
			switch handler := handlerFunc.(type) {
			case func(string):
				logger.Debugf("Consumer %s Received a message: %s\n", p.name, d.Body)
				handler(string(d.Body))
			case func(*ServerResult):
				var serverResult ServerResult
				err := json.Unmarshal(d.Body, &serverResult)
				if err != nil {
					logger.Debugf("Error Unmarshalling JSON %s", err)
					continue
				}
				logger.Debugf("Consumer %s Received a message on URL: %v\n", p.name, serverResult.Url)
				handler(&serverResult)
			case func(*Opts):
				localOpts := &Opts{
					Module:         opts.Module,
					Concurrency:    opts.Concurrency,
					Target:         string(d.Body),
					DorkFile:       opts.DorkFile,
					HopperFile:     opts.HopperFile,
					MethodFile:     opts.MethodFile,
					RedirectFile:   opts.RedirectFile,
					CnameFile:      opts.CnameFile,
					DispatcherFile: opts.DispatcherFile,
				}
				logger.Debugf("Consumer %s Received a message: %s\n", p.name, d.Body)
				handler(localOpts)

			}
			d.Ack(false)
		}
	}()
	logger.Infof(" [*] %s Waiting for messages. To exit press CTRL+C\n", p.name)
	<-forever
}

func NoMessagesWaiting() bool {
	for _, name := range queueNames {
		queueInfo, err := ch.QueueDeclarePassive(name, false, false, false, false, nil)
		failOnError(err, "Failed to inspect a queue")
		if queueInfo.Messages > 0 {
			logger.Debugf("Queue %s still has %d messages waiting", name, queueInfo.Messages)
			return false
		}
	}
	return true

}

func Close() {
	if ch != nil {
		ch.Close()
	}
	if conn != nil {
		conn.Close()
	}
}
