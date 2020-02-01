package base

import (
	"errors"
	"log"

	"github.com/nsqio/go-nsq"
)

// QueueProducer 消息队列生成者接口
type QueueProducer interface {
	Stop()
	Publish(topicName string, messageBody []byte) error
}

// NSQQueueProducer nsq的消息队列接口
type NSQQueueProducer struct {
	producer *nsq.Producer
}

// NewQueueProducer 新的消息生成者
func NewQueueProducer(address string) (QueueProducer, error) {
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer(address, config)
	if err != nil {
		return nil, err
	}

	return &NSQQueueProducer{producer: producer}, nil
}

// Stop 结束消息生产
func (qm *NSQQueueProducer) Stop() {
	if qm.producer != nil {
		qm.producer.Stop()
	}
}

// Publish 发送消息
func (qm *NSQQueueProducer) Publish(topicName string, messageBody []byte) error {
	// Synchronously publish a single message to the specified topic.
	// Messages can also be sent asynchronously and/or in batches.
	return qm.producer.Publish(topicName, messageBody)
}

// QueueConsumer 消息队列的消费者
type QueueConsumer struct {
	consumer *nsq.Consumer
	StopChan chan int
}

// Stop 结束消息消费
func (qc *QueueConsumer) Stop() {
	if qc.consumer != nil {
		qc.consumer.Stop()
	}
}

// AddHandler 添加消息回调
func (qc *QueueConsumer) AddHandler(hander nsq.Handler) {
	if qc.consumer == nil {
		return
	}
	qc.consumer.AddHandler(hander) //&myMessageHandler{}
}

// ConnectToNSQLookupd 连接消费队列
func (qc *QueueConsumer) ConnectToNSQLookupd(address string) error {
	if qc.consumer == nil {
		return errors.New("consumer nil")
	}
	err := qc.consumer.ConnectToNSQLookupd(address)
	if err != nil {
		return err
	}
	return nil
}

// NewConsumer 新建消费者
func NewConsumer(topic string, channel string) (*QueueConsumer, error) {
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &QueueConsumer{
		consumer: consumer,
		StopChan: consumer.StopChan,
	}, nil
}
