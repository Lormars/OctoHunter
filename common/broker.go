package common

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lormars/octohunter/internal/cacher"
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
	semaphore    chan struct{}
}

var OutputP = NewProducer("dork_broker", "gen")
var CnameP = NewProducer("cname_broker", "gen")
var RedirectP = NewProducer("redirect_broker", "gen")
var MethodP = NewProducer("method_broker", "gen")
var HopP = NewProducer("hopper_broker", "gen")
var DividerP = NewProducer("divider_broker", "gen")
var CrawlP = NewProducer("crawl_broker", "gen")
var SalesforceP = NewProducer("salesforce_broker", "gen")
var SplittingP = NewProducer("splitting_broker", "gen")
var Cl0P = NewProducer("cl0_broker", "gen")
var QuirksP = NewProducer("quirks_broker", "gen")
var RCP = NewProducer("rc_broker", "gen")
var CorsP = NewProducer("cors_broker", "gen")
var PathConfuseP = NewProducer("pathconfuse_broker", "gen")
var Fuzz4034P = NewProducer("fuzz4034_broker", "fuzz")
var PathTraversalP = NewProducer("pathtraversal_broker", "gen")
var FuzzAPIP = NewProducer("fuzzapi_broker", "fuzz")
var FuzzUnkeyedP = NewProducer("fuzzunkeyed_broker", "fuzz")
var FuzzPathP = NewProducer("fuzzpath_broker", "fuzz")
var XssP = NewProducer("xss_broker", "gen")
var SstiP = NewProducer("ssti_broker", "gen")
var WaybackP = NewProducer("wayback_broker", "wayback")
var GraphqlP = NewProducer("graphql_broker", "gen")
var MimeP = NewProducer("mime_broker", "gen")

var FuzzSemaphore = make(chan struct{}, 200)
var GenSemaphore = make(chan struct{}, 900)
var WaybackSemaphore = make(chan struct{}, 1)
var GlobalMu sync.Mutex
var mu sync.Mutex
var (
	queueProducers []*Producer
	WaitingQueue   = make(map[string]int)
	// concurrency    int
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func NewProducer(name, ptype string) *Producer {
	if ptype == "fuzz" {
		return &Producer{name: name, messageChan: make(chan interface{}, 1000), ShutdownChan: make(chan struct{}), semaphore: FuzzSemaphore, closed: false}
	} else if ptype == "wayback" {
		return &Producer{name: name, messageChan: make(chan interface{}, 1000), ShutdownChan: make(chan struct{}), semaphore: WaybackSemaphore, closed: false}
	} else {
		return &Producer{name: name, messageChan: make(chan interface{}, 1000), ShutdownChan: make(chan struct{}), semaphore: GenSemaphore, closed: false}
	}
}

func Init(options *Opts, purgebroker bool) []*Producer {
	// concurrency = options.Concurrency
	queueProducers = []*Producer{
		OutputP, CnameP, RedirectP, MethodP, HopP, DividerP, CrawlP,
		SalesforceP, SplittingP, Cl0P, QuirksP, RCP, CorsP, PathConfuseP, Fuzz4034P,
		PathTraversalP, FuzzAPIP, FuzzUnkeyedP, XssP, SstiP, WaybackP, GraphqlP, MimeP, FuzzPathP,
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
	mu.Lock()
	OutputP.pubCh = pubCh
	mu.Unlock()
}

func (p *Producer) PublishMessage(body interface{}) {
	var messageBody []byte
	var err error

	if p.name != "dork_broker" {
		// var waitCh chan bool
		switch v := body.(type) {
		case string:
			messageBody = []byte(v)
			hashed := Hash(string(messageBody))
			if !cacher.CheckCache(hashed, p.name) {
				return
			}
		case *ServerResult:
			messageBody, err = json.Marshal(v)
			if err != nil {
				failOnError(err, "Failed to marshal struct to JSON")
			}
			hashed := Hash(v.Url + v.Body + strconv.Itoa(v.StatusCode))
			if !cacher.CheckCache(hashed, p.name) {
				return
			}
		case *XssInput:
			messageBody, err = json.Marshal(v)
			if err != nil {
				failOnError(err, "Failed to marshal struct to JSON")
			}
			hashed := Hash(string(messageBody))
			if !cacher.CheckCache(hashed, p.name) {
				return
			}
		default:
			failOnError(fmt.Errorf("unknown type %T", v), "Failed to publish a message")
		}
		p.mu.Lock()
		isClosed := p.closed
		p.mu.Unlock()
		if !isClosed {
			p.messageChan <- messageBody
		}
	} else {
		messageBody = []byte(body.(string))
		mu.Lock()
		err = p.pubCh.Publish(
			"",            // exchange
			"dork_broker", // routing key
			false,         // mandatory
			false,         // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        messageBody,
			})
		mu.Unlock()
		if err != nil {
			logger.Warnf("Failed to publish a message: %s", err)
		}
	}
	// logger.Debugf(" [x] Sent to %s", p.name)
}

func (p *Producer) ConsumeMessage(handlerFunc interface{}) {
	go func() {
		for {
			select {
			case <-p.ShutdownChan:
				return
			default:
				d, ok := <-p.messageChan
				if !ok {
					logger.Debugf("Channel %s is closed", p.name)
					return
				}
				p.semaphore <- struct{}{}
				switch handler := handlerFunc.(type) {
				case func(string):
					logger.Debugf("Consumer %s Received a message: %s\n", p.name, d.([]byte))
					GlobalMu.Lock()
					ProducerBenches = append(ProducerBenches, ProducerBench{
						Producer: p.name,
						Time:     time.Now(),
						Hostname: GetHostname(string(d.([]byte))),
					})
					GlobalMu.Unlock()
					go func() {
						handler(string(d.([]byte)))
						<-p.semaphore
					}()
				case func(*ServerResult):
					var serverResult ServerResult
					err := json.Unmarshal(d.([]byte), &serverResult)
					if err != nil {
						logger.Warnf("Error Unmarshalling JSON %s", err)
						continue
					}
					logger.Debugf("Consumer %s Received a message on URL: %v\n", p.name, serverResult.Url)
					GlobalMu.Lock()
					ProducerBenches = append(ProducerBenches, ProducerBench{
						Producer: p.name,
						Time:     time.Now(),
						Hostname: GetHostname(serverResult.Url),
					})
					GlobalMu.Unlock()
					go func() {
						handler(&serverResult)
						<-p.semaphore
					}()
				case func(*XssInput):
					var xssInput XssInput
					err := json.Unmarshal(d.([]byte), &xssInput)
					if err != nil {
						logger.Warnf("Error Unmarshalling JSON %s", err)
						continue
					}
					logger.Debugf("Consumer %s Received a message on URL: %v\n", p.name, xssInput.Url)
					GlobalMu.Lock()
					ProducerBenches = append(ProducerBenches, ProducerBench{
						Producer: p.name,
						Time:     time.Now(),
						Hostname: GetHostname(xssInput.Url),
					})
					GlobalMu.Unlock()
					go func() {
						handler(&xssInput)
						<-p.semaphore
					}()
				default:
					failOnError(fmt.Errorf("unknown type %T", handler), "Failed to consume a message")
				}
			}
		}
	}()
}

func monitorChannels(producers []*Producer) {

	lastWait := make(map[string]int)
	for {
		time.Sleep(1 * time.Second)
		for _, p := range producers {
			GlobalMu.Lock()
			if p.name != "dork_broker" {
				name := strings.Split(p.name, "_")[0]
				if len(p.messageChan) > 500 && p.name != "crawl_broker" {
					logger.Debugf("Queue %s has %d messages waiting", p.name, len(p.messageChan))
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

func ConsumerUsage() int {
	return len(GenSemaphore) + len(FuzzSemaphore) + len(WaybackSemaphore)
}
