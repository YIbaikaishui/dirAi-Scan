package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Message 消息接口
type Message interface {
	GetID() string
	GetType() string
	GetPayload() []byte
	GetTimestamp() time.Time
}

// Queue 队列接口
type Queue interface {
	Publish(ctx context.Context, message Message) error
	Subscribe(ctx context.Context, messageType string, handler MessageHandler) error
	Close() error
}

// MessageHandler 消息处理器
type MessageHandler func(ctx context.Context, message Message) error

// simpleMessage 简单消息实现
type simpleMessage struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Payload   []byte    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

func (m *simpleMessage) GetID() string        { return m.ID }
func (m *simpleMessage) GetType() string      { return m.Type }
func (m *simpleMessage) GetPayload() []byte   { return m.Payload }
func (m *simpleMessage) GetTimestamp() time.Time { return m.Timestamp }

// inMemoryQueue 内存队列实现
type inMemoryQueue struct {
	mu        sync.RWMutex
	handlers  map[string][]MessageHandler
	messages  chan Message
	closed    bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewInMemoryQueue 创建内存队列
func NewInMemoryQueue(bufferSize int) Queue {
	ctx, cancel := context.WithCancel(context.Background())
	q := &inMemoryQueue{
		handlers: make(map[string][]MessageHandler),
		messages: make(chan Message, bufferSize),
		ctx:      ctx,
		cancel:   cancel,
	}

	// 启动消息处理协程
	go q.processMessages()

	return q
}

// NewMessage 创建新消息
func NewMessage(id, msgType string, payload interface{}) (Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &simpleMessage{
		ID:        id,
		Type:      msgType,
		Payload:   data,
		Timestamp: time.Now(),
	}, nil
}

// Publish 发布消息
func (q *inMemoryQueue) Publish(ctx context.Context, message Message) error {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.closed {
		return fmt.Errorf("queue is closed")
	}

	select {
	case q.messages <- message:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-q.ctx.Done():
		return fmt.Errorf("queue is closed")
	}
}

// Subscribe 订阅消息
func (q *inMemoryQueue) Subscribe(ctx context.Context, messageType string, handler MessageHandler) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return fmt.Errorf("queue is closed")
	}

	q.handlers[messageType] = append(q.handlers[messageType], handler)
	return nil
}

// Close 关闭队列
func (q *inMemoryQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil
	}

	q.closed = true
	q.cancel()
	close(q.messages)
	return nil
}

// processMessages 处理消息
func (q *inMemoryQueue) processMessages() {
	for {
		select {
		case message, ok := <-q.messages:
			if !ok {
				return
			}
			q.handleMessage(message)
		case <-q.ctx.Done():
			return
		}
	}
}

// handleMessage 处理单个消息
func (q *inMemoryQueue) handleMessage(message Message) {
	q.mu.RLock()
	handlers := q.handlers[message.GetType()]
	q.mu.RUnlock()

	for _, handler := range handlers {
		go func(h MessageHandler) {
			if err := h(q.ctx, message); err != nil {
				// TODO: 添加错误处理逻辑
				fmt.Printf("Error handling message %s: %v\n", message.GetID(), err)
			}
		}(handler)
	}
}