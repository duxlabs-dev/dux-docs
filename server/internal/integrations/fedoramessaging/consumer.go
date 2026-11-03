package fedoramessaging

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/duxlabs-dev/dux-docs/server/internal/log"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	PublicBrokerURL = "amqps://fedora:@rabbitmq.fedoraproject.org/%2Fpublic_pubsub"
	exchange        = "amq.topic"
	prefetch        = 25
	minBackoff      = 5 * time.Second
	maxBackoff      = 5 * time.Minute
)

type Handler func(ctx context.Context, topic string, body []byte)

type Consumer struct {
	url     string
	tls     *tls.Config
	topics  []string
	handler Handler
}

func New(url, tlsDir string, topics []string, handler Handler) (*Consumer, error) {
	tlsCfg, err := loadTLS(tlsDir)
	if err != nil {
		return nil, err
	}
	if url == "" {
		url = PublicBrokerURL
	}
	return &Consumer{url: url, tls: tlsCfg, topics: topics, handler: handler}, nil
}

func (c *Consumer) Run(ctx context.Context) {
	backoff := minBackoff
	for {
		err := c.consume(ctx)
		if ctx.Err() != nil {
			return
		}
		log.Warn("Fedora messaging connection lost, reconnecting", "err", err, "in", backoff)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, maxBackoff)
	}
}

func (c *Consumer) consume(ctx context.Context) error {
	conn, err := amqp.DialConfig(c.url, amqp.Config{
		TLSClientConfig: c.tls,
		SASL:            []amqp.Authentication{&amqp.ExternalAuth{}},
		Heartbeat:       60 * time.Second,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.Qos(prefetch, 0, false); err != nil {
		return err
	}

	queue, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return err
	}
	for _, topic := range c.topics {
		if err := ch.QueueBind(queue.Name, topic, exchange, false, nil); err != nil {
			return err
		}
	}

	deliveries, err := ch.Consume(queue.Name, "", false, true, false, false, nil)
	if err != nil {
		return err
	}
	log.Info("Connected to Fedora messaging", "queue", queue.Name, "topics", c.topics)

	connClosed := conn.NotifyClose(make(chan *amqp.Error, 1))
	chClosed := ch.NotifyClose(make(chan *amqp.Error, 1))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-connClosed:
			return err
		case err := <-chClosed:
			return err
		case d, ok := <-deliveries:
			if !ok {
				return amqp.ErrClosed
			}
			c.handler(ctx, d.RoutingKey, d.Body)
			if err := d.Ack(false); err != nil {
				return err
			}
		}
	}
}
