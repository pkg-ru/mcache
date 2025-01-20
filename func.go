package mcache

import (
	"slices"
	"time"
)

func (c *controlCache) startGC() {
	for {
		// ожидаем время установленное в cleanupInterval
		<-time.After(c.cleanupInterval)

		if c.items == nil {
			return
		}

		// Ищем элементы с истекшим временем жизни и удаляем из хранилища
		if keys := c.expiredKeys(); len(keys) != 0 {
			c.clearItems(keys)
		}

	}

}

// expiredKeys возвращает список "просроченных" ключей
func (c *controlCache) expiredKeys() (keys []string) {
	c.RLock()
	defer c.RUnlock()
	for k, i := range c.items {
		if i.expiration > 0 && time.Now().UnixNano() > i.expiration {
			keys = append(keys, k)
		}
	}
	return
}

// clearItems удаляет ключи из переданного списка
func (c *controlCache) clearItems(keys []string) {
	c.Lock()
	defer c.Unlock()

	for _, k := range keys {
		delete(c.items, k)
	}
}

// expiredKeys возвращает список ключей содержащих "тэг"
func (c *controlCache) tagKeys(tag string) (keys []string) {
	c.RLock()
	defer c.RUnlock()
	for k, i := range c.items {
		if slices.Contains(i.tags, tag) {
			keys = append(keys, k)
		}
	}
	return
}
