package subpub

import "testing"

func TestNewSubPub(t *testing.T) {
	var sp interface{} = NewSubPub()

	if structType, ok := sp.(*subPub); !ok {
		t.Errorf("NewSubPub() expected a *subPub struct, got %v", structType)
	}
}

func TestSubscribe(t *testing.T) {

	sp := NewSubPub()

	handler := func(msg interface{}) {
		t.Logf("internal Subscribe() handler: %v", msg)
	}

	sub, err := sp.Subscribe("test.subject1", handler)
	if err != nil {
		t.Fatalf("internal Subscribe() error: %v", err)
	}

	if structType, ok := sub.(*subscription); !ok {
		t.Errorf("Subscribe() expected a *subscription struct, got %v", structType)
	}
}
