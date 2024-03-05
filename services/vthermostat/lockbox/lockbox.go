package lockbox

import (
	"log"
	"sync"
)

type Key struct{}

type LockBox[T any] struct {
	data  T
	key   Key
	mutex sync.Mutex
}

func New[T any](data T) *LockBox[T] {
	return &LockBox[T]{
		data:  data,
		mutex: sync.Mutex{},
	}
}

func (l *LockBox[T]) Take() (T, Key) {
	l.mutex.Lock()
	l.key = Key{}
	return l.data, l.key
}

func (l *LockBox[T]) Put(data T, key Key) {
	if l.key != key {
		log.Fatalln("incorrect key")
	}
	l.data = data
	l.mutex.Unlock()
}

func (l *LockBox[T]) Release(key Key) {
	if l.key != key {
		log.Fatalln("incorrect key")
	}
	l.mutex.Unlock()
}
