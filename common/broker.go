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
	name         string
	conn         *amqp.Connection
	pubChannel   *amqp.Channel
	consChannel  *amqp.Channel
	mutex        sync.Mutex
	semaphore    chan struct{}
	consumers    []func()
	ShutdownChan chan struct{}
}

var OutputP = NewProducer("dork_broker")
var CnameP = NewProducer("cname_broker")
var RedirectP = NewProducer("redirect_broker")
var MethodP = NewProducer("method_broker")
var HopP = NewProducer("hopper_broker")
var DividerP = NewProducer("divider_broker")
var CrawlP = NewProducer("crawl_broker")
var SalesforceP = NewProducer("salesforce_broker")
var SplittingP = NewProducer("splitting_broker")
var Cl0P = NewProducer("cl0_broker")
var QuirksP = NewProducer("quirks_broker")
var mu sync.Mutex

var (
	queueProducers []*Producer
	concurrency    int
	purge          bool
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func NewProducer(name string) *Producer {
	return &Producer{name: name, ShutdownChan: make(chan struct{})}
}

func Init(options *Opts, purgebroker bool) []*Producer {
	concurrency = options.Concurrency
	purge = purgebroker
	queueProducers = []*Producer{
		OutputP, CnameP, RedirectP, MethodP, HopP, DividerP, CrawlP,
		SalesforceP, SplittingP, Cl0P, QuirksP,
	}
	for _, p := range queueProducers {
		p.initConnection()
		p.semaphore = make(chan struct{}, concurrency*100)
	}

	return queueProducers
}

func (p *Producer) initConnection() {
	conn, pubCh, consCh, err := connectRabbitMQ()
	failOnError(err, "Failed to connect to RabbitMQ")

	p.conn = conn
	p.pubChannel = pubCh
	p.consChannel = consCh

	err = p.initQueue(pubCh, true)
	failOnError(err, "Failed to initialize queue")

	//go p.checkQueue()
	go p.registerConsumers()
}

func connectRabbitMQ() (*amqp.Connection, *amqp.Channel, *amqp.Channel, error) {
	user := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")
	connStr := fmt.Sprintf("amqp://%s:%s@localhost:5672/", user, password)
	conn, err := amqp.DialConfig(connStr,
		amqp.Config{
			Heartbeat: 10 * time.Second,
		})
	if err != nil {
		return nil, nil, nil, err
	}

	pubCh, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, nil, err
	}

	err = pubCh.Qos(
		100,   // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		pubCh.Close()
		conn.Close()
		return nil, nil, nil, err
	}

	consCh, err := conn.Channel()
	if err != nil {
		pubCh.Close()
		conn.Close()
		return nil, nil, nil, err
	}

	err = consCh.Qos(
		100,   // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		consCh.Close()
		pubCh.Close()
		conn.Close()
		return nil, nil, nil, err
	}

	return conn, pubCh, consCh, nil
}

func (p *Producer) initQueue(ch *amqp.Channel, check bool) error {
	var err error
	if purge && check {
		_, err = ch.QueuePurge(p.name, false)
	}
	logger.Debugf("Purging queue error: %v", err)
	_, err = ch.QueueDeclare(
		p.name,
		false,
		false,
		false,
		false,
		nil,
	)

	return err
}

func (p *Producer) reconnect() {
	for {

		if p.pubChannel != nil {
			p.pubChannel.Close()
		}
		if p.consChannel != nil {
			p.consChannel.Close()
		}
		if p.conn != nil {
			p.conn.Close()
		}

		var err error
		p.conn, p.pubChannel, p.consChannel, err = connectRabbitMQ()
		if err != nil {
			logger.Warnf("Failed to reconnect to RabbitMQ, retrying in 5 seconds: %s", err)
			time.Sleep(5 * time.Second)
			continue
		}

		err = p.initQueue(p.pubChannel, false)
		if err != nil {
			logger.Warnf("Failed to declare queue, retrying in 5 seconds: %s", err)
			p.pubChannel.Close()
			p.conn.Close()
			time.Sleep(5 * time.Second)
			continue
		}
		//p.checkQueue()
		p.registerConsumers()
		logger.Infof("Successfully reconnected to RabbitMQ for queue %s", p.name)
		break
	}
}

func (p *Producer) PublishMessage(body interface{}) {
	var messageBody []byte
	var contentType string
	var err error

	if p.name != "dork_broker" {
		p.semaphore <- struct{}{}
		defer func() {
			<-p.semaphore
		}()
		var waitCh chan bool
		switch v := body.(type) {
		case string:
			messageBody = []byte(v)
			hostname := GetHostname(v)
			mu.Lock()
			BrokerSliding.AddRequest(hostname)
			mu.Unlock()
			waitCh = AddToBrokerQueue(hostname)
			contentType = "text/plain"
		case *ServerResult:
			messageBody, err = json.Marshal(v)
			hostname := GetHostname(v.Url)
			mu.Lock()
			BrokerSliding.AddRequest(hostname)
			mu.Unlock()
			waitCh = AddToBrokerQueue(hostname)
			if err != nil {
				failOnError(err, "Failed to marshal struct to JSON")
			}
			contentType = "application/json"
		default:
			failOnError(fmt.Errorf("unknown type %T", v), "Failed to publish a message")
		}
		<-waitCh
	} else {
		messageBody = []byte(body.(string))
		contentType = "text/plain"
	}
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		p.mutex.Lock()
		if p.pubChannel == nil || p.pubChannel.IsClosed() {
			logger.Warnf("Failed to publish a message, attempting to reconnect: %s", p.name)
			p.reconnect()
			p.mutex.Unlock()
			cancel()
			continue
		}
		err = p.pubChannel.PublishWithContext(
			ctx,
			"",
			p.name,
			false,
			false,
			amqp.Publishing{
				ContentType: contentType,
				Body:        messageBody,
			})
		p.mutex.Unlock()
		if err != nil {
			logger.Warnf("Failed to publish a message, attempting to reconnect: %s", err)
			cancel()
			p.reconnect()
		} else {
			cancel()
			break
		}
	}
	failOnError(err, "Failed to publish a message")
	logger.Debugf(" [x] Sent to %s", p.name)
}

func (p *Producer) ConsumeMessage(handlerFunc interface{}, opts *Opts) {
	for {
		select {
		case <-p.ShutdownChan:
			logger.Debugf("Consumer %s Shutting down", p.name)
			return
		default:

			msgs, err := p.consChannel.Consume(
				p.name,
				"",
				false,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				logger.Warnf("Failed to register a consumer: %s", err)
				p.mutex.Lock()
				p.reconnect()
				p.mutex.Unlock()
				continue
			}

			p.consumers = append(p.consumers, func() { p.ConsumeMessage(handlerFunc, opts) })
			closeChan := p.consChannel.NotifyClose(make(chan *amqp.Error))
			var forever = make(chan struct{})
			go func() {

				for {
					select {
					case d := <-msgs:
						d.Ack(false)
						switch handler := handlerFunc.(type) {
						case func(string):
							logger.Debugf("Consumer %s Received a message: %s\n", p.name, d.Body)
							handler(string(d.Body))
						case func(*ServerResult):
							var serverResult ServerResult
							err := json.Unmarshal(d.Body, &serverResult)
							if err != nil {
								logger.Warnf("Error Unmarshalling JSON %s", err)
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
					case <-closeChan:
						logger.Warnf("Consumer %s Connection closed, attempting to reconnect", p.name)
						close(forever)
						return
					case <-p.ShutdownChan:
						return
					}
				}
			}()
			logger.Debugf(" [*] %s Waiting for messages. To exit press CTRL+C\n", p.name)
			<-forever
		}
	}
}

func (p *Producer) registerConsumers() {
	for _, consumer := range p.consumers {
		consumer()
	}
}

func NoMessagesWaiting() bool {
	for _, p := range queueProducers {
		queueInfo, err := p.pubChannel.QueueDeclarePassive(p.name, false, false, false, false, nil)
		failOnError(err, "Failed to inspect a queue")
		if queueInfo.Messages > 0 {
			logger.Debugf("Queue %s still has %d messages waiting", p.name, queueInfo.Messages)
			return false
		}
	}
	return true
}

func Close() {
	for _, p := range queueProducers {
		if p.pubChannel != nil {
			p.pubChannel.Close()
		}
		if p.consChannel != nil {
			p.consChannel.Close()
		}
		if p.conn != nil {
			p.conn.Close()
		}
	}
}
