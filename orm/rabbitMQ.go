package orm

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	reconnectDelay   = 5 * time.Second // 连接断开后多久重连
	reconnectMaxTime = 0               // 发送消息或订阅时，等待重连次数 0 一直重复连接
)

var RabbitMQ Rabbit

type RabbitMQConf struct {
	Host      string `env:"RABBITMQ_HOST"`
	Port      int64  `env:"RABBITMQ_PORT"`
	User      string `env:"RABBITMQ_USER"`
	Password  string `env:"RABBITMQ_PASSWORD"`
	Publish   string `env:"RABBITMQ_PUBLISH_TOPIC"`
	Subscribe string `env:"RABBITMQ_SUBSCRIBE_TOPIC"`
}
type Rabbit struct {
	connection    *amqp.Connection
	channel       *amqp.Channel
	isConnected   bool             // 是否已连接
	done          chan bool        // 正常关闭
	notifyClose   chan *amqp.Error // 异常关闭
	notifyConnect chan bool        // 连接成功提醒
	url           string           // rabbitMQ地址
	SubScribeMap  sync.Map         // 订阅或消息
}
type SubscribeCallback func(string)
type SubScribe struct {
	Name     string   // name
	Kind     string   // type
	Keys     []string //keys
	Callback SubscribeCallback
}
type Publish struct {
	Name string // name
	Kind string // type
	Key  string // routing key
	Body string // body
}

// 初始化 rabbitMQ中的交换关系
type SubScribeExchange struct {
	Name     string              `json:"name"` // name
	Kind     string              `json:"kind"` // type
	Keys     []string            `json:"keys"` // key
	Children []SubScribeExchange `json:"children"`
}

// rabbitMQ 初始化
func (r *Rabbit) Init(mqConf RabbitMQConf) error {
	log.Println("connect rabbitMQ")

	r.notifyConnect = make(chan bool)
	r.done = make(chan bool)
	r.isConnected = false
	r.url = fmt.Sprintf("amqp://%s:%s@%s:%d/",
		mqConf.User,
		mqConf.Password,
		mqConf.Host,
		mqConf.Port,
	)

	if err := r.Conn(); err != nil {
		return errors.New("rabbitMQ connect fail")
	}

	//初始化exchange和channel
	subscribeExchangeArr := []SubScribeExchange{
		//{
		//	Name: "topic_one",
		//	Kind: "topic",
		//	Children: []SubScribeExchange{
		//		{
		//			Name: "topic_two",
		//			Kind: "topic",
		//			Keys: []string{
		//				"two",
		//			},
		//		},
		//		{
		//			Name: "channel_three",
		//			Kind: "channel",
		//			Keys: []string{
		//				"three",
		//			},
		//		},
		//	},
		//},
		//{
		//	Name: "channel_four",
		//	Kind: "channel",
		//	Keys: []string{
		//		"four",
		//	},
		//},
	}
	if len(subscribeExchangeArr) > 0 {
		for _, v := range subscribeExchangeArr {
			r.InitExchange(v)
		}
	}

	go r.Connect()
	return nil
}
func (r *Rabbit) Connect() {
	i := 1
	for {
		if !r.isConnected {
			log.Println("[rabbitMQ]", "connect rabbitMQ")
			if err := r.Conn(); err != nil {
				log.Println("[rabbitMQ]", i, err.Error(), "Failed to connect rabbitMQ. Retrying...")
			}
			i++
		}

		select {
		case <-r.done:
			return
		case <-r.notifyClose:
			if r.isConnected {
				r.channel.Close()
				r.connection.Close()
				r.isConnected = false
			}
		}
		time.Sleep(reconnectDelay)
	}
}
func (r *Rabbit) Conn() error {
	conn, err := amqp.Dial(r.url)
	if err != nil {
		return err
	}

	channel, err := conn.Channel()
	if err != nil {
		return err
	}

	r.connection = conn
	r.channel = channel

	r.isConnected = true

	r.notifyClose = make(chan *amqp.Error)
	r.channel.NotifyClose(r.notifyClose)

	//重新执行订阅
	r.SubscribeRun()

	//提醒重新连接成功
	if r.notifyConnect != nil {
		close(r.notifyConnect)
	}
	r.notifyConnect = make(chan bool)
	return nil
}
func (r *Rabbit) Close() {
	if r.done != nil {
		//close(r.done)
	}
	if r.isConnected {
		r.channel.Close()
		r.connection.Close()
		r.isConnected = false
	}
}
func (r *Rabbit) wait() error {
	if r.isConnected {
		return nil
	}

	idleDelay := time.NewTimer(reconnectDelay)
	defer idleDelay.Stop()

	i := 1
	for {
		idleDelay.Reset(reconnectDelay)
		select {
		case <-r.done:
			goto END
		case <-r.notifyConnect:
		case <-idleDelay.C:
		}

		if r.isConnected {
			return nil
		}
		if reconnectMaxTime > 0 && i >= reconnectMaxTime {
			goto END
		}
		i++
	}
END:
	return errors.New("connect rabbitMQ fail")
}
func (r *Rabbit) InitExchange(subScribeExchange SubScribeExchange) error {
	switch subScribeExchange.Kind {
	case "channel":
		_, err := r.channel.QueueDeclare(
			subScribeExchange.Name, // name
			false,                  // durable
			false,                  // delete when unused
			false,                  // exclusive
			false,                  // no-wait
			nil,                    // arguments
		)
		if err != nil {
			return err
		}
		return nil
	case "direct":
	case "fanout":
	case "headers":
	case "topic":
	default:
		return errors.New("kind is wrong")
	}
	err := r.channel.ExchangeDeclare(
		subScribeExchange.Name, // name
		subScribeExchange.Kind, // type
		true,                   // durable
		false,                  // auto-deleted
		false,                  // internal
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		return err
	}
	if subScribeExchange.Children == nil && len(subScribeExchange.Children) <= 0 {
		return nil
	}
	for _, vTmp := range subScribeExchange.Children {
		if err := r.InitExchange(vTmp); err == nil {
			if vTmp.Keys == nil || len(vTmp.Keys) <= 0 {
				vTmp.Keys = append(vTmp.Keys, "")
			}
			switch vTmp.Kind {
			case "channel":
				for _, k := range vTmp.Keys {
					err := r.channel.QueueBind(
						vTmp.Name,              // queue name
						k,                      // routing key
						subScribeExchange.Name, // exchange
						false,
						nil)
					if err != nil {
						return err
					}
				}
				return nil
			case "direct":
			case "fanout":
			case "headers":
			case "topic":
			default:
				return errors.New("kind is wrong")
			}
			for _, k := range vTmp.Keys {
				err := r.channel.ExchangeBind(
					vTmp.Name,              // name
					k,                      // routing key
					subScribeExchange.Name, // exchange
					false,
					nil)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// 执行订阅
func (r *Rabbit) Subscribe(name, kind string, callback SubscribeCallback, keys ...string) {
	if keys == nil || len(keys) <= 0 {
		keys = append(keys, "")
	}
	r.SubScribeMap.Store(strings.Join(append(keys, name), "_"), SubScribe{
		Name:     name,
		Kind:     kind,
		Keys:     keys,
		Callback: callback,
	})
}
func (r *Rabbit) SubscribeRun() error {
	if err := r.wait(); err != nil {
		return err
	}
	r.SubScribeMap.Range(func(k, v interface{}) bool {
		if v == nil || reflect.TypeOf(v).Name() != "SubScribe" {
			return false
		}
		subScribe := v.(SubScribe)
		if err := r.SubscribeAction(subScribe); err != nil {
			log.Println(subScribe.Name, err.Error())
		}
		return true
	})
	return nil
}
func (r *Rabbit) SubscribeAction(subScribe SubScribe) error {
	if subScribe.Name == "" {
		return errors.New("name is null")
	}

	switch subScribe.Kind {
	case "channel":
		if _, err := r.channel.QueueDeclare(
			subScribe.Name, // name
			false,          // durable
			false,          // delete when unused
			false,          // exclusive
			false,          // no-wait
			nil,            // arguments
		); err != nil {
			return err
		}
		msgs, err := r.channel.Consume(
			subScribe.Name, // name
			"",             // consumer
			true,           // auto-ack
			false,          // exclusive
			false,          // no-local
			false,          // no-wait
			nil,            // arg
		)
		if err != nil {
			return err
		}
		go func() {
			for d := range msgs {
				subScribe.Callback(fmt.Sprintf("%s", d.Body))
			}
		}()
		log.Println("[subscribe]", "rabbitMQ", subScribe.Kind, subScribe.Name, subScribe.Keys)
		return nil
	case "direct":
	case "fanout":
	case "headers":
	case "topic":
	default:
		return errors.New("kind `" + subScribe.Kind + "` is wrong")
	}

	if err := r.channel.ExchangeDeclare(
		subScribe.Name, // name
		subScribe.Kind, // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	); err != nil {
		return err
	}

	q, err := r.channel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	if len(subScribe.Keys) <= 0 {
		if err := r.channel.QueueBind(
			q.Name,         // queue name
			"",             // routing key
			subScribe.Name, // exchange
			false,
			nil,
		); err != nil {
			return err
		}
	} else {
		for _, s := range subScribe.Keys {
			if err := r.channel.QueueBind(
				q.Name,         // queue name
				s,              // routing key
				subScribe.Name, // exchange
				false,
				nil,
			); err != nil {
				return err
			}
		}
	}

	msgs, err := r.channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			subScribe.Callback(fmt.Sprintf("%s", d.Body))
		}
	}()
	log.Println("[subscribe]", "rabbitMQ", subScribe.Kind, subScribe.Name, subScribe.Keys)
	return nil
}

// 发送订阅消息
func (r *Rabbit) Publish(publish Publish) error {
	//等待连接， 如果一直无法连接则退出
	if err := r.wait(); err != nil {
		return err
	}
	if publish.Kind == "channel" {
		return r.PublishChannel(publish)
	}
	return r.PublishDefault(publish)
}
func (r *Rabbit) PublishDefault(publish Publish) error {
	if err := r.channel.ExchangeDeclare(
		publish.Name, // name
		publish.Kind, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	); err != nil {
		return err
	}
	return r.channel.Publish(
		publish.Name, // exchange
		publish.Key,  // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(publish.Body),
		})
}
func (r *Rabbit) PublishChannel(publish Publish) error {
	if _, err := r.channel.QueueDeclare(
		publish.Name, // name
		false,        // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	); err != nil {
		return err
	}
	return r.channel.Publish(
		"",           // exchange
		publish.Name, // name
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(publish.Body),
		})
}
