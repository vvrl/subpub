package subpub

import (
	"context"
	"errors"
	"sync"
)

// MessageHandler is a callback function that processes messages delivered to subscribers
type MessageHandler func(msg interface{})

type Subscription interface {
	// Unsubscribe will remove interest in the current subject subscription is for
	Unsubscribe()
}

type SubPub interface {
	// Subscribe creates an asynchronous queue subscriber on the given subject
	Subscribe(subject string, cb MessageHandler) (Subscription, error)

	// Publish publishes the msg argument to the given subject
	Publish(subject string, msg interface{}) error

	// Close will shutdown sub-pub system
	// May be blocked by data delivery until the context is canceled
	Close(ctx context.Context) error
}

type subscriber struct {
	handler MessageHandler
	msgChan chan interface{}
}

type subscription struct {
	sub     *subscriber
	subject string
	sp      *subPub
}

func (s *subscription) Unsubscribe() {
	if s == nil || s.sp == nil {
		return
	}

	s.sp.mu.Lock()
	defer s.sp.mu.Unlock()
	defer close(s.sub.msgChan)

	if subjectSubs, ok := s.sp.subscribers[s.subject]; ok {
		delete(subjectSubs, s.sub)
		if len(subjectSubs) == 0 { // Если подписок на тему больше нет, удаляем тему из общей мапы
			delete(s.sp.subscribers, s.subject)
		}
	}

}

type subPub struct {
	mu          sync.RWMutex
	subscribers map[string]map[*subscriber]struct{}
	close       bool
	wg          sync.WaitGroup
	closeChan   chan struct{}
}

func (s *subPub) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	s.mu.Lock()

	if s.close {
		s.mu.Unlock()
		return nil, errors.New("subpub is stopped")
	}

	newSubcription := &subscription{
		sub: &subscriber{
			handler: cb,
			msgChan: make(chan interface{}, 32),
		},
		subject: subject,
		sp:      s,
	}

	// newSub := &subscriber{
	// 	handler: cb,
	// 	msgChan: make(chan interface{}, 32),
	// }

	if _, ok := s.subscribers[subject]; !ok {
		s.subscribers[subject] = make(map[*subscriber]struct{})
	}

	s.subscribers[subject][newSubcription.sub] = struct{}{} // добавление подписчика в тему

	s.wg.Add(1)
	s.mu.Unlock()

	//запуск обработчика сообщений для каждого подписчика
	go func(s *subscription) {
		defer s.sp.wg.Done() // уменьшаем waitGroup при завершении работы горутины

		for message := range s.sub.msgChan {
			s.sub.handler(message)
		}

	}(newSubcription)

	return newSubcription, nil
}

func (s *subPub) Publish(subject string, msg interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.close {
		return errors.New("subpub is stopped")
	}

	for sub := range s.subscribers[subject] {
		defer func() { // обработка паники на всякий случай, например если подписчик отписался
			if r := recover(); r != nil {
			}
		}()
		sub.msgChan <- msg
	}

	return nil
}

func (s *subPub) Close(ctx context.Context) error {
	s.mu.Lock()
	if s.close {
		s.mu.Unlock()
		return nil // Система уже закрыта
	}
	s.close = true

	s.mu.Unlock() // Освобождаем блокировку перед ожиданием WaitGroup

	go func() {
		s.wg.Wait()        // Ожидаем завершения всех горутин, запущенных Publish
		close(s.closeChan) // Сигнализируем, что все завершилось
	}()

	select {
	case <-s.closeChan:
		return nil // Корректное завершение
	case <-ctx.Done():
		// Контекст был отменен (например, по таймауту)
		return ctx.Err()
	}
}

func NewSubPub() SubPub {
	return &subPub{
		subscribers: make(map[string]map[*subscriber]struct{}),
		closeChan:   make(chan struct{}),
	}
}
