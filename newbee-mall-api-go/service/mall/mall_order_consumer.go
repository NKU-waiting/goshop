package mall

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	kf "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"main.go/constants"
	"main.go/global"
	"main.go/model/common"
	"main.go/model/common/enum"
	"main.go/model/mall"
	mallReq "main.go/model/mall/request"
	"main.go/model/manage"
	"main.go/pkg/kafka"
	"time"
)

// WorkerPool 管理 goroutine 池
type WorkerPool struct {
	workerCount int
	jobs        chan kf.Message
	quit        chan struct{}
}

// NewWorkerPool 创建新的 worker pool
func NewWorkerPool(workerCount int) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		jobs:        make(chan kf.Message, workerCount*2),
		quit:        make(chan struct{}),
	}
}

// Start 启动 worker pool
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		go wp.worker(i)
	}
}

// worker 处理任务
func (wp *WorkerPool) worker(id int) {
	defer func() {
		if r := recover(); r != nil {
			if global.GVA_LOG != nil {
				global.GVA_LOG.Error("Worker panic recovered", zap.Int("worker_id", id), zap.Any("panic", r))
			} else {
				fmt.Printf("Worker %d panic recovered: %v\n", id, r)
			}
		}
	}()

	for {
		select {
		case msg := <-wp.jobs:
			processOrderMessage(msg)
		case <-wp.quit:
			if global.GVA_LOG != nil {
				global.GVA_LOG.Info("Worker shutting down", zap.Int("worker_id", id))
			}
			return
		}
	}
}

// Submit 提交任务到 worker pool
func (wp *WorkerPool) Submit(msg kf.Message) {
	select {
	case wp.jobs <- msg:
	case <-wp.quit:
		if global.GVA_LOG != nil {
			global.GVA_LOG.Warn("Worker pool is shutting down, message dropped")
		}
	}
}

// Stop 停止 worker pool
func (wp *WorkerPool) Stop() {
	close(wp.quit)
	close(wp.jobs)
}

func RunOrderConsumer() {
	consumer := kafka.InitConsumer(global.GVA_CONFIG.Kafka.Addr, global.GVA_CONFIG.Kafka.Topic, global.GVA_CONFIG.Kafka.Group)
	defer consumer.Close()

	// 创建 worker pool，限制并发数为 10
	workerPool := NewWorkerPool(10)
	workerPool.Start()
	defer workerPool.Stop()

	fmt.Println("Starting Kafka order consumer with worker pool...")

	for {
		m, err := consumer.ReadMessage(context.Background())
		if err != nil {
			global.GVA_LOG.Error("Failed to read message from Kafka", zap.Error(err))
			continue
		}

		// 提交到 worker pool 而不是直接创建 goroutine
		workerPool.Submit(m)
	}
}

func processOrderMessage(m kf.Message) {
	startTime := time.Now()
	var orderEvent mallReq.OrderEvent

	// 记录开始处理
	global.GVA_LOG.Info("Start processing order message",
		zap.String("key", string(m.Key)),
		zap.Int("partition", m.Partition),
		zap.Int64("offset", m.Offset))

	err := json.Unmarshal(m.Value, &orderEvent)
	if err != nil {
		global.GVA_LOG.Error("Failed to unmarshal order event", zap.Error(err), zap.ByteString("message", m.Value))
		return
	}

	global.GVA_LOG.Info("Processing order", zap.String("orderNo", orderEvent.OrderNo))

	// 使用事务来确保数据一致性
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		// 0. 检查订单是否已存在（幂等性）
		var existingOrder manage.MallOrder
		checkErr := tx.Where("order_no = ?", orderEvent.OrderNo).First(&existingOrder).Error
		if checkErr == nil {
			global.GVA_LOG.Info("Order already exists, skipping",
				zap.String("orderNo", orderEvent.OrderNo))
			return nil
		}

		// 1. 检查库存
		for _, item := range orderEvent.Items {
			var goodsInfo manage.MallGoodsInfo
			if err := tx.Where("goods_id = ?", item.GoodsId).First(&goodsInfo).Error; err != nil {
				return fmt.Errorf("商品 %d 不存在", item.GoodsId)
			}
			if goodsInfo.StockNum < item.GoodsCount {
				return fmt.Errorf("商品 %s 库存不足", item.GoodsName)
			}
		}

		// 2. 扣减库存
		for _, item := range orderEvent.Items {
			if err := tx.Model(&manage.MallGoodsInfo{}).Where("goods_id = ?", item.GoodsId).Update("stock_num", gorm.Expr("stock_num - ?", item.GoodsCount)).Error; err != nil {
				return fmt.Errorf("扣减商品 %d 库存失败: %w", item.GoodsId, err)
			}
		}

		// 3. 计算总价
		priceTotal := 0
		for _, item := range orderEvent.Items {
			priceTotal += item.GoodsCount * item.SellingPrice
		}
		if priceTotal < constants.MinOrderPrice {
			return errors.New("订单价格异常")
		}

		// 4. 创建订单
		newBeeMallOrder := manage.MallOrder{
			OrderNo:    orderEvent.OrderNo,
			UserId:     orderEvent.UserId,
			TotalPrice: priceTotal,
			PayStatus:  0, // 待支付
			OrderStatus: enum.ORDER_PRE_PAY.Code(), // 待支付
			CreateTime: common.JSONTime{Time: time.Now()},
			UpdateTime: common.JSONTime{Time: time.Now()},
		}
		if err := tx.Save(&newBeeMallOrder).Error; err != nil {
			return fmt.Errorf("创建订单失败: %w", err)
		}

		// 5. 创建订单地址
		var newBeeMallOrderAddress mall.MallOrderAddress
		copier.Copy(&newBeeMallOrderAddress, &orderEvent.Address)
		newBeeMallOrderAddress.OrderId = newBeeMallOrder.OrderId
		if err := tx.Save(&newBeeMallOrderAddress).Error; err != nil {
			return fmt.Errorf("创建订单地址失败: %w", err)
		}

		// 6. 创建订单项
		var newBeeMallOrderItems []manage.MallOrderItem
		for _, item := range orderEvent.Items {
			var newBeeMallOrderItem manage.MallOrderItem
			copier.Copy(&newBeeMallOrderItem, &item)
			newBeeMallOrderItem.OrderId = newBeeMallOrder.OrderId
			newBeeMallOrderItem.CreateTime = common.JSONTime{Time: time.Now()}
			newBeeMallOrderItems = append(newBeeMallOrderItems, newBeeMallOrderItem)
		}
		if err := tx.Save(&newBeeMallOrderItems).Error; err != nil {
			return fmt.Errorf("创建订单项失败: %w", err)
		}

		// 7. 删除购物车项目
		if err := tx.Where("cart_item_id in ?", orderEvent.ShoppingCartItemIds).Updates(mall.MallShoppingCartItem{IsDeleted: 1}).Error; err != nil {
			return fmt.Errorf("删除购物车项目失败: %w", err)
		}

		return nil
	})

	// 计算处理时间
	duration := time.Since(startTime)

	if err != nil {
		global.GVA_LOG.Error("Failed to process order",
			zap.String("orderNo", orderEvent.OrderNo),
			zap.Error(err),
			zap.Duration("duration", duration))
		// TODO: 可以在这里添加逻辑，例如将失败的订单信息存入死信队列
	} else {
		global.GVA_LOG.Info("Order processed successfully",
			zap.String("orderNo", orderEvent.OrderNo),
			zap.Duration("duration", duration),
			zap.Int("items_count", len(orderEvent.Items)))
	}
}
