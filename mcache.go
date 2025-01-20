package mcache

import (
	"sync"
	"time"
)

type controlCache struct {
	items             map[string]cacheData
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	// sizeBuffer        int
	// sizeMax           int
	sync.RWMutex
}

type cacheData struct {
	data       any
	expiration int64
	// count      int
	// uptime     int64
	tags []string
}

// получаем новый экземпляр для работы с кешем
func New(defaultExpiration, cleanupInterval time.Duration) *controlCache {
	items := make(map[string]cacheData)

	cache := controlCache{
		items:             items,
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
	}

	if cache.defaultExpiration <= 0 {
		cache.defaultExpiration = time.Hour * 24 * 7
	}

	if cache.cleanupInterval <= 0 {
		cache.cleanupInterval = time.Minute * 5
	}
	go cache.startGC()

	return &cache
}

// Устанавливаем кеш
func (c *controlCache) Set(key string, data any, duration time.Duration, tags ...string) {
	var expiration int64

	// Если продолжительность жизни равна 0 - используется значение по-умолчанию
	if duration == 0 {
		duration = c.defaultExpiration
	}

	// Устанавливаем время истечения кеша
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.Lock()
	c.items[key] = cacheData{
		data:       data,
		expiration: expiration,
		tags:       tags,
	}
	c.Unlock()
}

// Получаем данные из кеша
func (c *controlCache) Get(key string) (any, bool) {
	c.RLock()
	defer c.RUnlock()

	item, is := c.items[key]

	// ключ не найден
	if !is {
		return nil, false
	}

	// Проверка на установку времени истечения, в противном случае он бессрочный
	if item.expiration > 0 {
		// Если в момент запроса кеш устарел возвращаем nil
		if time.Now().UnixNano() > item.expiration {
			return nil, false
		}
	}

	return item.data, true
}

// получаем данные из кеша или из результата выполнения функции с последующим кешированием данных
//
//	type User struct {
//		Id int
//		Name string
//	}
//
//	cache := mcache.New(time.Hour, time.Minute*10)
//
//	user := cache.GetFunc("user_1", func() *User {
//		return &User{1, "Vlad"}
//	}, time.Second * 30)
//
//	fmt.Println("Hello", user.Name)
func (c *controlCache) GetFunc(key string, callback func() any, duration time.Duration, tags ...string) any {
	data, is := c.Get(key)
	if !is {
		data = callback()
		c.Set(key, data, duration, tags...)
	}
	return data
}

// Удаляем по ключу
func (c *controlCache) Delete(key string) bool {
	c.Lock()
	defer c.Unlock()

	if _, found := c.items[key]; !found {
		return false
	}

	delete(c.items, key)
	return true
}

// Удаляем все
func (c *controlCache) FlushAll() {
	c.Lock()
	c.items = make(map[string]cacheData)
	c.Unlock()
}

// Удаляем по тегу
func (c *controlCache) FlushTag(tag string) {
	if keys := c.tagKeys(tag); len(keys) != 0 {
		c.clearItems(keys)
	}
}
