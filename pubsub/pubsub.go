package pubsub

import "sync"

// Subscriber – канал для отправки сообщений подписчику
type Subscriber chan interface{}

// PubSub реализует базовый механизм pub/sub
type PubSub struct {
	mu          sync.RWMutex
	subscribers map[string][]Subscriber
}

func NewPubSub() *PubSub {
	return &PubSub{
		subscribers: make(map[string][]Subscriber),
	}
}

// Subscribe – подписка на определённую тему
func (ps *PubSub) Subscribe(topic string) Subscriber {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ch := make(Subscriber)
	ps.subscribers[topic] = append(ps.subscribers[topic], ch)
	return ch
}

// Unsubscribe – отменяет подписку
func (ps *PubSub) Unsubscribe(topic string, sub Subscriber) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	subs := ps.subscribers[topic]
	for i, s := range subs {
		if s == sub {
			ps.subscribers[topic] = append(subs[:i], subs[i+1:]...)
			close(s)
			break
		}
	}
}

// Publish – публикует сообщение в определённую тему
func (ps *PubSub) Publish(topic string, msg interface{}) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	for _, sub := range ps.subscribers[topic] {
		// Отправляем сообщение в горутине, чтобы не блокировать публикацию
		go func(s Subscriber) {
			s <- msg
		}(sub)
	}
}
