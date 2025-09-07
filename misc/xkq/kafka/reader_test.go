package kafka

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

func TestBatchReader(t *testing.T) {

	wg := sync.WaitGroup{}
	wg.Add(2)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Topic:   "test_kfk_topic",
		Brokers: []string{"127.0.0.1:9094"},
		Dialer: &kafka.Dialer{
			DualStack: true,
			SASLMechanism: plain.Mechanism{
				Username: os.Getenv("ENV_KFK_USERNAME"),
				Password: os.Getenv("ENV_KFK_PASSWORD"),
			},
		},
		GroupID: "batchreader-group",
	})

	batchReader := NewBatchReader(reader, BatchReaderConfig{
		BatchSize:    10,
		BatchTimeout: time.Second * 2,
	})

	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*59)
		defer cancel()
		for {
			fmt.Printf("before batch fetch: %v\n", time.Now().Format(time.DateTime))
			msgs, err := batchReader.BatchFetchMessages(ctx)
			if err != nil {
				break
			}

			fmt.Printf("after batch fetch: %v\n", time.Now().Format(time.DateTime))
			for _, msg := range msgs {
				fmt.Printf("msg [key=%s, value=%s]\n", msg.Key, msg.Value)
			}

			err = batchReader.CommitMessages(ctx, msgs...)
			if err != nil {
				fmt.Printf("can not commit messages %v\n", err)
			}
		}
	}()

	writer := kafka.Writer{
		Addr: kafka.TCP("127.0.0.1:9094"),
		Transport: &kafka.Transport{
			SASL: plain.Mechanism{
				Username: os.Getenv("ENV_KFK_USERNAME"),
				Password: os.Getenv("ENV_KFK_PASSWORD"),
			},
		},
		BatchSize: 1,
	}

	go func() {
		defer wg.Done()
		time.Sleep(time.Second)
		for i := range 30 {
			writer.WriteMessages(context.Background(), kafka.Message{
				Key:   []byte("test-batch-reader-msg-key"),
				Value: []byte(strconv.Itoa(i)),
				Topic: "test_kfk_topic",
			})
			time.Sleep(time.Millisecond * 10)
		}
	}()

	wg.Wait()
	batchReader.Close()
}
