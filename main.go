package mcache

import (
	"time"
)

var instance *controlCache

func init() {
	if instance == nil {
		instance = New(time.Hour*2, time.Minute*10)
	}
}

// настройка: время жизни кеша по умолчанию, и интервал очистки кеша
func Setting(defaultExpiration, cleanupInterval time.Duration) {
	instance.defaultExpiration = defaultExpiration
	instance.cleanupInterval = cleanupInterval

	if instance.defaultExpiration <= 0 {
		instance.defaultExpiration = time.Hour * 24 * 7
	}

	if instance.cleanupInterval <= 0 {
		instance.cleanupInterval = time.Minute * 5
	}
}

// получаем данные
//
//	type User struct {
//		Id int
//		Name string
//	}
//
//	user, is := mcache.Get[User]("user_1")
//	if is {
//		fmt.Println("Hello", user.Name)
//	}
func Get[T any](key string) (T, bool) {
	res, is := instance.Get(key)
	if is {
		data, is := res.(T)
		if is {
			return data, true
		}
	}
	return *new(T), false
}

// сохраняем данные
func Set(key string, data any, duration time.Duration, tags ...string) {
	instance.Set(key, data, duration, tags...)
}

// получаем данные из кеша или из результата выполнения функции с последующим кешированием данных
//
//	type User struct {
//		Id int
//		Name string
//	}
//
//	user := mcache.GetFunc("user_1", func() *User {
//		return &User{1, "Vlad"}
//	}, time.Second * 30)
//
//	fmt.Println("Hello", user.Name)
func GetFunc[T any](key string, callback func() T, duration time.Duration, tags ...string) T {
	res := instance.GetFunc(key, func() any {
		return callback()
	}, duration, tags...)

	data, is := res.(T)
	if is {
		return data
	}

	return *new(T)
}

// Удаляем по ключу
func Delete(key string) bool {
	return instance.Delete(key)
}

// Удаляем все
func FlushAll() {
	instance.FlushAll()
}

// Удаляем по тегу
func FlushTag(tag string) {
	instance.FlushTag(tag)
}
