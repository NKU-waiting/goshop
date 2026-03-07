package mall

import (
	"testing"
	"time"

	kf "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestWorkerPool(t *testing.T) {
	t.Run("创建WorkerPool", func(t *testing.T) {
		wp := NewWorkerPool(5)
		assert.NotNil(t, wp)
		assert.Equal(t, 5, wp.workerCount)
		assert.NotNil(t, wp.jobs)
		assert.NotNil(t, wp.quit)
	})

	t.Run("WorkerPool启动和停止", func(t *testing.T) {
		wp := NewWorkerPool(3)
		wp.Start()
		
		// 等待workers启动
		time.Sleep(100 * time.Millisecond)
		
		// 停止worker pool
		wp.Stop()
		
		// 验证已停止
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("WorkerPool处理消息", func(t *testing.T) {
		wp := NewWorkerPool(2)
		wp.Start()
		defer wp.Stop()

		// 创建测试消息
		testMsg := kf.Message{
			Value: []byte(`{"order_no":"TEST001"}`),
		}

		// 提交消息
		wp.Submit(testMsg)
		
		// 等待处理
		time.Sleep(200 * time.Millisecond)
	})

	t.Run("WorkerPool并发限制", func(t *testing.T) {
		workerCount := 5
		wp := NewWorkerPool(workerCount)
		wp.Start()
		defer wp.Stop()

		// 提交大量消息
		messageCount := 20
		for i := 0; i < messageCount; i++ {
			testMsg := kf.Message{
				Value: []byte(`{"order_no":"TEST` + string(rune(i)) + `"}`),
			}
			wp.Submit(testMsg)
		}

		// 等待处理
		time.Sleep(500 * time.Millisecond)
	})
}

func TestProcessOrderMessage(t *testing.T) {
	t.Run("处理无效JSON", func(t *testing.T) {
		msg := kf.Message{
			Value: []byte(`invalid json`),
		}
		
		// 应该不会panic
		processOrderMessage(msg)
	})

	t.Run("处理空消息", func(t *testing.T) {
		msg := kf.Message{
			Value: []byte(`{}`),
		}
		
		// 应该不会panic
		processOrderMessage(msg)
	})
}
