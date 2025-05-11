package subpub_test

import (
	"context"
	"errors"
	"fmt"
	"subpub-vk/subpub"
	"sync"
	"testing"
	"time"
)

type Handler struct {
	wg       *sync.WaitGroup
	mu       sync.Mutex
	messages []interface{}
	sign     string
}

func (h *Handler) GetMessages() ([]interface{}, int) {

	h.mu.Lock()
	defer h.mu.Unlock()

	messagesCopy := make([]interface{}, len(h.messages))
	numMessages := copy(messagesCopy, h.messages) // копирование чтобы не трогать истиный рабочий слайс + получение кол-ва элементов слайса

	return messagesCopy, numMessages
}

func CreateHandler(sign string, wg *sync.WaitGroup) *Handler {
	return &Handler{
		wg:       wg,
		messages: make([]interface{}, 0),
		sign:     sign,
	}
}

func (h *Handler) Handling(msg interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.messages = append(h.messages, msg)
	fmt.Printf("Handler '%s' received: %v (Total: %d)\n", h.sign, msg, len(h.messages))
	h.wg.Done()
}

func TestSubscribeAndPublish(t *testing.T) {
	sp := subpub.NewSubPub()

	wg := new(sync.WaitGroup)

	message := "hello"
	subject := "test.subject1"
	handler := CreateHandler("test.handler.1", wg)

	_, err := sp.Subscribe(subject, handler.Handling)
	if err != nil {
		t.Fatalf("Subscribe() error: %v", err)
	}

	wg.Add(1)
	err = sp.Publish(subject, message)
	if err != nil {
		t.Fatalf("Publish() error: %v", err)
	}
	wg.Wait()

	receivedMessages, messageNums := handler.GetMessages()

	if messageNums != 1 {
		t.Fatalf("Expected 1 message. got %v", messageNums)
	}
	if receivedMessages[0] != message {
		t.Errorf("Expected message '%s', got '%v'", message, receivedMessages[0])
	}

	sp.Close(context.Background())

}

func TestUnsubscribe_StopsMessages(t *testing.T) {
	sp := subpub.NewSubPub()
	wg := new(sync.WaitGroup)

	subject := "test.subject1"
	message1 := "msg 1"
	message2 := "msg 2"
	handler := CreateHandler("test.unsubscribe", wg)

	sub, err := sp.Subscribe(subject, handler.Handling)
	if err != nil {
		t.Fatalf("Subscribe() error: %v", err)
	}

	wg.Add(1)
	err = sp.Publish(subject, message1)
	if err != nil {
		t.Fatalf("Publish(message1) error: %v", err)
	}

	wg.Wait()

	sub.Unsubscribe()

	err = sp.Publish(subject, message2)
	if err != nil {
		t.Fatalf("Publish(message2) error: %v", err)
	}

	_, msgNum := handler.GetMessages()

	if msgNum != 1 {
		t.Errorf("Expected 1 message only, got %d", msgNum)
	}
}

func TestSubscribeAfterCloseSubPub(t *testing.T) {
	sp := subpub.NewSubPub()
	Handler := CreateHandler("test.subAfterClose", nil)
	expectedErr := errors.New("subpub is already stopped for subscribe")

	err := sp.Close(context.Background())
	if err != nil {
		t.Fatal("Close subpub error")
	}
	_, err = sp.Subscribe("closed sp", Handler.Handling)

	if err.Error() != expectedErr.Error() {
		t.Fatalf("Subscribe after close expected: %v, got: %v", expectedErr, err)
	}
}

func TestPublishAfterCloseSubPub(t *testing.T) {
	sp := subpub.NewSubPub()
	expectedErr := errors.New("subpub is already stopped for publish")

	err := sp.Close(context.Background())
	if err != nil {
		t.Fatal("Close subpub error")
	}

	err = sp.Publish("closed sp", "")

	if err.Error() != expectedErr.Error() {
		t.Fatalf("Publish after close expected: %v, got: %v", expectedErr, err)
	}
}

func TestCloseClosedSubPub(t *testing.T) {
	sp := subpub.NewSubPub()
	expectedErr := errors.New("subPub is already closed")

	err := sp.Close(context.Background())
	if err != nil {
		t.Fatal("Close subpub error")
	}

	err = sp.Close(context.Background())
	if err.Error() != expectedErr.Error() {
		t.Errorf("SubPub close after closing error expected: %v, got: %v", expectedErr, err)
	}
}

func TestCloseContext(t *testing.T) {
	sp := subpub.NewSubPub()
	subject := "test.closeContext"

	handler := func(msg interface{}) {
		time.Sleep(time.Second)
	}

	_, err := sp.Subscribe(subject, handler)
	if err != nil {
		t.Fatal("Subscribe error")
	}

	err = sp.Publish(subject, "long time message")
	if err != nil {
		t.Fatal("Publish error")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	closeErr := sp.Close(ctx)

	if !errors.Is(closeErr, context.DeadlineExceeded) {
		t.Errorf("Close with context expected: %v, gor: %v", closeErr, context.DeadlineExceeded)
	}
}
