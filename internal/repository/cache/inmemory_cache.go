package cache

import (
	"sync"
	"time"

	"order-service0/internal/domain/entities"
)

type inMemoryCache struct {
	mu      sync.RWMutex
	orders  map[string]*cacheEntry
	maxSize int
	ttl     time.Duration
}

type cacheEntry struct {
	order      *entities.Order
	expiresAt  time.Time
	lastAccess time.Time
}

func NewInMemoryCache(maxSize int, ttl time.Duration) *inMemoryCache {
	if maxSize <= 0 {
		maxSize = 1000 // дефолтный размер
	}
	if ttl <= 0 {
		ttl = 15 * time.Minute // дефолтное время жизни
	}
	cache := &inMemoryCache{
		orders:  make(map[string]*cacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}
	// Запускаем горутину для периодической очистки
	go cache.cleanupWorker()
	return cache
}

func (c *inMemoryCache) Set(orderUID string, order *entities.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Если кэш заполнен, удаляем самый старый элемент
	if len(c.orders) >= c.maxSize {
		c.evictOldest()
	}
	c.orders[orderUID] = &cacheEntry{
		order:      order,
		expiresAt:  time.Now().Add(c.ttl),
		lastAccess: time.Now(),
	}
}

func (c *inMemoryCache) Get(orderUID string) (*entities.Order, bool) {
	c.mu.RLock()
	entry, exists := c.orders[orderUID]
	c.mu.RUnlock()
	if !exists {
		return nil, false
	}
	// Проверяем не истек ли TTL
	if time.Now().After(entry.expiresAt) {
		c.mu.Lock()
		delete(c.orders, orderUID)
		c.mu.Unlock()
		return nil, false
	}
	// Обновляем время последнего доступа
	c.mu.Lock()
	entry.lastAccess = time.Now()
	c.mu.Unlock()
	return entry.order, true
}

func (c *inMemoryCache) GetAll() map[string]*entities.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]*entities.Order)
	now := time.Now()
	for k, v := range c.orders {
		if now.Before(v.expiresAt) {
			result[k] = v.order
		}
	}
	return result
}

func (c *inMemoryCache) Restore(orders map[string]*entities.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders = make(map[string]*cacheEntry)
	for k, v := range orders {
		c.orders[k] = &cacheEntry{
			order:      v,
			expiresAt:  time.Now().Add(c.ttl),
			lastAccess: time.Now(),
		}
	}
}

func (c *inMemoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	for k, v := range c.orders {
		if oldestKey == "" || v.lastAccess.Before(oldestTime) {
			oldestKey = k
			oldestTime = v.lastAccess
		}
	}
	if oldestKey != "" {
		delete(c.orders, oldestKey)
	}
}

func (c *inMemoryCache) cleanupWorker() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		c.cleanupExpired()
	}
}

func (c *inMemoryCache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, v := range c.orders {
		if now.After(v.expiresAt) {
			delete(c.orders, k)
		}
	}
}

