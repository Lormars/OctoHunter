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
var SalesforceP = Producer{name: "salesforce_broker"}
var SplittingP = Producer{name: "splitting_broker"}
var Cl0P = Producer{name: "cl0_broker"}

var (
	conn      *amqp.Connection
	ch        *amqp.Channel
	mutex     sync.Mutex
	semaphore map[string]int = make(map[string]int)
)
var queueNames = []string{
	"dork_broker", "cname_broker", "redirect_broker",
	"method_broker", "hopper_broker", "divider_broker", "crawl_broker",
	"salesforce_broker", "splitting_broker", "cl0_broker",
}

var concurrency int

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

var purge bool

func Init(options *Opts, purgebroker bool) {
	var err error
	concurrency = options.Concurrency
	purge = purgebroker
	conn, ch, err = connectRabbitMQ()
	failOnError(err, "Failed to connect to RabbitMQ")

	err = initQueues(ch)
	failOnError(err, "Failed to initialize queues")
	checkQueue()

}

func connectRabbitMQ() (*amqp.Connection, *amqp.Channel, error) {
	user := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")
	connStr := fmt.Sprintf("amqp://%s:%s@localhost:5672/", user, password)
	conn, err := amqp.Dial(connStr)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	err = ch.Qos(
		100,   // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, nil, err
	}

	return conn, ch, nil
}

func initQueues(ch *amqp.Channel) error {
	var err error
	for _, name := range queueNames {
		//first purge queue
		if purge {
			_, err = ch.QueuePurge(name, false)
		}
		logger.Debugf("Purging queue error: %v", err)
		_, err = ch.QueueDeclare(
			name,
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func reconnect() {
	for {
		var err error
		conn, ch, err = connectRabbitMQ()
		if err != nil {
			logger.Warnf("Failed to reconnect to RabbitMQ, retrying in 2 seconds: %s", err)
			time.Sleep(2 * time.Second)
			continue
		}

		err = initQueues(ch)
		if err != nil {
			logger.Warnf("Failed to declare queues, retrying in 2 seconds: %s", err)
			ch.Close()
			conn.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		logger.Infof("Successfully reconnected to RabbitMQ")
		break
	}
}

func checkQueue() {
	for _, name := range queueNames {
		if name == "dork_broker" { //no need for rate limit on dork_broker
			continue
		}
		queueInfo, err := ch.QueueDeclarePassive(name, false, false, false, false, nil)
		if err != nil {
			failOnError(err, "Failed to inspect queue"+name)
		}
		semaphore[name] = queueInfo.Messages
		logger.Debugf("Queue %s has %d messages ready", name, queueInfo.Messages)
	}

}

func (p Producer) PublishMessage(body interface{}) {
	var messageBody []byte
	var contentType string
	var err error
	if p.name != "dork_broker" { //no need for rate limit on dork_broker
		for {
			mutex.Lock()
			logger.Debugf("Semaphore %s: %d", p.name, semaphore[p.name])
			if semaphore[p.name] < concurrency*100 {
				logger.Debugf("Waiting for semaphore %s with queue: %d", p.name, semaphore[p.name])
				mutex.Unlock()
				break
			}
			mutex.Unlock()
			time.Sleep(2 * time.Second)
		}

		mutex.Lock()
		semaphore[p.name]++
		mutex.Unlock()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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
	for {
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
		if err != nil {
			logger.Warnf("Failed to publish a message, attempting to reconnect: %s", err)
			reconnect()
		} else {
			break
		}
	}
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
			mutex.Lock()
			semaphore[p.name]--
			mutex.Unlock()
		}
	}()
	logger.Debugf(" [*] %s Waiting for messages. To exit press CTRL+C\n", p.name)
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
