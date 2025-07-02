package di

import (
	"fmt"
	"reflect"
	"sync"
)

// Container 依赖注入容器
type Container struct {
	mu        sync.RWMutex
	services  map[string]interface{}
	factories map[string]func() interface{}
}

// NewContainer 创建新的容器
func NewContainer() *Container {
	return &Container{
		services:  make(map[string]interface{}),
		factories: make(map[string]func() interface{}),
	}
}

// Register 注册服务
func (c *Container) Register(name string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = service
}

// RegisterFactory 注册工厂函数
func (c *Container) RegisterFactory(name string, factory func() interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.factories[name] = factory
}

// Get 获取服务
func (c *Container) Get(name string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 首先检查已注册的服务
	if service, exists := c.services[name]; exists {
		return service, nil
	}

	// 然后检查工厂函数
	if factory, exists := c.factories[name]; exists {
		service := factory()
		c.services[name] = service // 缓存服务实例
		return service, nil
	}

	return nil, fmt.Errorf("service %s not found", name)
}

// MustGet 获取服务，如果不存在则panic
func (c *Container) MustGet(name string) interface{} {
	service, err := c.Get(name)
	if err != nil {
		panic(err)
	}
	return service
}

// GetAs 获取服务并转换为指定类型
func (c *Container) GetAs(name string, target interface{}) error {
	service, err := c.Get(name)
	if err != nil {
		return err
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	serviceValue := reflect.ValueOf(service)
	targetType := targetValue.Elem().Type()

	if !serviceValue.Type().AssignableTo(targetType) {
		return fmt.Errorf("service %s is not assignable to %s", serviceValue.Type(), targetType)
	}

	targetValue.Elem().Set(serviceValue)
	return nil
}

// Has 检查服务是否存在
func (c *Container) Has(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, hasService := c.services[name]
	_, hasFactory := c.factories[name]
	return hasService || hasFactory
}

// Remove 移除服务
func (c *Container) Remove(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.services, name)
	delete(c.factories, name)
}

// Clear 清空容器
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services = make(map[string]interface{})
	c.factories = make(map[string]func() interface{})
}

// ListServices 列出所有服务名称
func (c *Container) ListServices() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var names []string
	for name := range c.services {
		names = append(names, name)
	}
	for name := range c.factories {
		if _, exists := c.services[name]; !exists {
			names = append(names, name)
		}
	}
	return names
}