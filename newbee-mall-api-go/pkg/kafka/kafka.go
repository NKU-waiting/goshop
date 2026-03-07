package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	kf "github.com/segmentio/kafka-go"
)

// Producer 全局 Kafka 生产者
var Producer *kf.Writer

// InitProducer 初始化 Kafka 生产者
func InitProducer(brokerURL, topic string) {
	Producer = &kf.Writer{
		Addr:     kf.TCP(brokerURL), // Kafka broker 地址
		Topic:    topic,             // 默认主题
		Balancer: &kf.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond, // 批量发送超时
		MaxAttempts:  3,                     // 最大重试次数
		RequiredAcks: kf.RequireOne,         // 至少发送到一个副本
	}
	fmt.Println("Kafka producer initialized for topic:", topic)
}

// SendMessage 发送消息到 Kafka
func SendMessage(ctx context.Context, key, value []byte) error {
	if Producer == nil {
		return fmt.Errorf("Kafka producer not initialized")
	}
	msg := kf.Message{
		Key:   key,
		Value: value,
	}
	return Producer.WriteMessages(ctx, msg)
}

// CloseProducer 关闭 Kafka 生产者
func CloseProducer() error {
	if Producer != nil {
		fmt.Println("Closing Kafka producer...")
		err := Producer.Close()
		if err != nil {
			log.Printf("Error closing Kafka producer: %v\n", err)
			return err
		}
		fmt.Println("Kafka producer closed successfully")
	}
	return nil
}

// CloseConsumer 关闭 Kafka 消费者
func CloseConsumer(reader *kf.Reader) error {
	if reader != nil {
		fmt.Println("Closing Kafka consumer...")
		err := reader.Close()
		if err != nil {
			log.Printf("Error closing Kafka consumer: %v\n", err)
			return err
		}
		fmt.Println("Kafka consumer closed successfully")
	}
	return nil
}

// InitConsumer 初始化 Kafka 消费者
func InitConsumer(brokerURL, topic, groupID string) *kf.Reader {
	reader := kf.NewReader(kf.ReaderConfig{
		Brokers:  []string{brokerURL},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  1 * time.Second, // 最多等待 1 秒
		QueueCapacity: 100,        // 队列容量
		HeartbeatInterval: 3 * time.Second,
		CommitInterval:    1 * time.Second, // 自动提交 offset
		MaxAttempts:       3,
	})
	fmt.Printf("Kafka consumer initialized for topic: %s, groupID: %s\n", topic, groupID)
	return reader
}

// ConsumeMessages 消费消息的示例
func ConsumeMessages(ctx context.Context, reader *kf.Reader, handler func(msg kf.Message)) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer context cancelled, stopping message consumption.")
			return
		default:
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v\n", err)
				continue
			}
			handler(m)
		}
	}
}
