package common

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lormars/octohunter/internal/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	name         string
	messageChan  chan interface{}
	ShutdownChan chan struct{}
	pubCh        *amqp.Channel
	closed       bool
	closeOnce    sync.Once
	mu           sync.Mutex
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
var RCP = NewProducer("rc_broker")
var CorsP = NewProducer("cors_broker")
var PathConfuseP = NewProducer("pathconfuse_broker")
var Fuzz4034P = NewProducer("fuzz4034_broker")
var PathTraversalP = NewProducer("pathtraversal_broker")
var FuzzAPIP = NewProducer("fuzzapi_broker")
var FuzzUnkeyedP = NewProducer("fuzzunkeyed_broker")
var XssP = NewProducer("xss_broker")
var SstiP = NewProducer("ssti_broker")
var WaybackP = NewProducer("wayback_broker")
var GraphqlP = NewProducer("graphql_broker")
var MimeP = NewProducer("mime_broker")

var mu sync.Mutex
var GlobalMu sync.Mutex

var (
	queueProducers []*Producer
	WaitingQueue   = make(map[string]int)
	// concurrency    int
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func NewProducer(name string) *Producer {
	return &Producer{name: name, messageChan: make(chan interface{}, 1000), ShutdownChan: make(chan struct{}), closed: false}
}

func Init(options *Opts, purgebroker bool) []*Producer {
	// concurrency = options.Concurrency
	queueProducers = []*Producer{
		OutputP, CnameP, RedirectP, MethodP, HopP, DividerP, CrawlP,
		SalesforceP, SplittingP, Cl0P, QuirksP, RCP, CorsP, PathConfuseP, Fuzz4034P,
		PathTraversalP, FuzzAPIP, FuzzUnkeyedP, XssP, SstiP, WaybackP, GraphqlP, MimeP,
	}

	rabbitMQSetup()
	go monitorChannels(queueProducers)

	return queueProducers
}

func rabbitMQSetup() {
	user := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")
	connStr := fmt.Sprintf("amqp://%s:%s@localhost:5672/", user, password)
	conn, err := amqp.DialConfig(connStr,
		amqp.Config{
			Heartbeat: 10 * time.Second,
		})
	failOnError(err, "Failed to connect to RabbitMQ")

	pubCh, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	err = pubCh.Qos(
		100,   // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")
	_, err = pubCh.QueueDeclare(
		"dork_broker", // name
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")
	OutputP.pubCh = pubCh

}

func (p *Producer) PublishMessage(body interface{}) {
	var messageBody []byte
	var err error

	if p.name != "dork_broker" {
		var waitCh chan bool
		switch v := body.(type) {
		case string:
			messageBody = []byte(v)
			hostname := GetHostname(v)
			mu.Lock()
			BrokerSliding.AddRequest(hostname)
			mu.Unlock()
			waitCh = AddToBrokerQueue(hostname)
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
		case *XssInput:
			messageBody, err = json.Marshal(v)
			hostname := GetHostname(v.Url)
			mu.Lock()
			BrokerSliding.AddRequest(hostname)
			mu.Unlock()
			waitCh = AddToBrokerQueue(hostname)
			if err != nil {
				failOnError(err, "Failed to marshal struct to JSON")
			}
		default:
			failOnError(fmt.Errorf("unknown type %T", v), "Failed to publish a message")
		}
		<-waitCh
		mu.Lock()
		if !p.closed {
			p.messageChan <- messageBody
		}
		mu.Unlock()
	} else {
		messageBody = []byte(body.(string))
		err = p.pubCh.Publish(
			"",            // exchange
			"dork_broker", // routing key
			false,         // mandatory
			false,         // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        messageBody,
			})
		if err != nil {
			logger.Warnf("Failed to publish a message: %s", err)
		}
	}
	// logger.Debugf(" [x] Sent to %s", p.name)
}

func (p *Producer) ConsumeMessage(handlerFunc interface{}, opts *Opts) chan struct{} {
	closeChan := make(chan struct{})
	go func() {
		for {
			select {
			case <-closeChan:
				return
			case <-p.ShutdownChan:
				return
			default:
				d, ok := <-p.messageChan
				if !ok {
					return
				}
				switch handler := handlerFunc.(type) {
				case func(string):
					logger.Debugf("Consumer %s Received a message: %s\n", p.name, d.([]byte))
					handler(string(d.([]byte)))
				case func(*ServerResult):
					var serverResult ServerResult
					err := json.Unmarshal(d.([]byte), &serverResult)
					if err != nil {
						logger.Warnf("Error Unmarshalling JSON %s", err)
						continue
					}
					logger.Debugf("Consumer %s Received a message on URL: %v\n", p.name, serverResult.Url)
					handler(&serverResult)
				case func(*XssInput):
					var xssInput XssInput
					err := json.Unmarshal(d.([]byte), &xssInput)
					if err != nil {
						logger.Warnf("Error Unmarshalling JSON %s", err)
						continue
					}
					logger.Debugf("Consumer %s Received a message on URL: %v\n", p.name, xssInput.Url)
					handler(&xssInput)
				default:
					failOnError(fmt.Errorf("unknown type %T", handler), "Failed to consume a message")
				}
			}
		}
	}()
	return closeChan
}

func monitorChannels(producers []*Producer) {

	lastWait := make(map[string]int)
	for {
		time.Sleep(1 * time.Second)
		for _, p := range producers {
			GlobalMu.Lock()
			if p.name != "dork_broker" {
				name := strings.Split(p.name, "_")[0]
				if len(p.messageChan) > 500 {
					logger.Infof("Queue %s has %d messages waiting", p.name, len(p.messageChan))
					WaitingQueue[name] = 10 //if longer than 500, then just add more
					lastWait[p.name] = len(p.messageChan)
				} else {
					WaitingQueue[name] = len(p.messageChan) - lastWait[p.name]
					lastWait[p.name] = len(p.messageChan)
				}
			}
			GlobalMu.Unlock()
		}
	}
}

func (p *Producer) Close() {
	p.closeOnce.Do(func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.closed = true
		close(p.messageChan)
	})
}

func Close() {
	for _, p := range queueProducers {
		p.Close()
	}
}
