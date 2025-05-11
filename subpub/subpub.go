package subpub

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type MessageHandler func(msg interface{})

type Subscription interface {
	Unsubscribe()
}

type SubPub interface {
	Subscribe(subject string, cb MessageHandler) (Subscription, error)
	Publish(subject string, msg interface{}) error
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

	if subjectSubs, ok := s.sp.subscribers[s.subject]; ok {
		delete(subjectSubs, s.sub)
		if len(subjectSubs) == 0 { // Если подписок на тему больше нет, удаляем тему из общей мапы
			delete(s.sp.subscribers, s.subject)
		}
	}
	s.sp.mu.Unlock()
	close(s.sub.msgChan)

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
		return nil, errors.New("subpub is already stopped for subscribe")
	}

	newSubcription := &subscription{
		sub: &subscriber{
			handler: cb,
			msgChan: make(chan interface{}, 32),
		},
		subject: subject,
		sp:      s,
	}

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
		return errors.New("subpub is already stopped for publish")
	}

	for sub := range s.subscribers[subject] {
		select {
		case sub.msgChan <- msg:
			// сообщение безприпятственно отправлено
		default:
			fmt.Print("failed sending message: subject ", subject)
			// если есть какие-то проблемы с каналом, сообщение пропустится и не отправится
		}

	}

	return nil
}

func (s *subPub) Close(ctx context.Context) error {
	s.mu.Lock()
	if s.close {
		s.mu.Unlock()
		return errors.New("subPub is already closed")
	}
	s.close = true

	chanToClose := []chan interface{}{}

	for _, subject := range s.subscribers {
		for sub := range subject {
			chanToClose = append(chanToClose, sub.msgChan)
		}
	}

	s.mu.Unlock()

	for _, ch := range chanToClose {
		close(ch)
	}

	go func() {
		s.wg.Wait()
		close(s.closeChan)
	}()

	select {
	case <-s.closeChan:
		return nil // все горутины корректно завершились
	case <-ctx.Done():
		return ctx.Err() // закрытие контекста
	}
}

func NewSubPub() SubPub {
	return &subPub{
		subscribers: make(map[string]map[*subscriber]struct{}),
		closeChan:   make(chan struct{}),
	}
}
