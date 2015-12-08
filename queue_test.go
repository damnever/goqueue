package goqueue

import (
	"fmt"
	"testing"
	"time"
)

func TestUnBlockGet(t *testing.T) {
	value := 8888

	fmt.Println("Test unblock get, should no error...")
	queue := NewQueue()
	queue.Put(value)
	e, err := queue.Get(false, 0)
	if err != nil {
		t.Fatalf("Unexpect error: %v\n", err)
	} else if e.Value.(int) != value {
		t.Fatalf("Expect %v, got %v\n", value, e.Value.(int))
	}
	fmt.Println("  ...PASSED")

	fmt.Println("Test unblock get, error should show up...")
	emptyQ := NewQueue()
	_, err1 := emptyQ.Get(false, 0)
	if err1 == nil {
		t.Fatalf("No error show up\n")
	}
	fmt.Println("  ...PASSED")
}

func TestBlockGetWithTimeout(t *testing.T) {
	w := make(chan bool)
	queue := NewQueue()

	get := func(timeout float64, value int, noErr bool) {
		e, err := queue.Get(true, timeout)
		if noErr {
			if err == nil && e.Value.(int) != value {
				t.Fatalf("Expect %v, got %v\n", value, e.Value.(int))
			} else if err != nil {
				t.Fatalf("Unexpect error: %v\n", err)
			}
		} else if err == nil {
			t.Fatalf("Wanted an error, but got nil\n")
		}
		w <- true
	}

	put := func(d time.Duration, value int) {
		time.Sleep(time.Second * d)
		queue.Put(value)
	}

	fmt.Println("Test block get with timeout, should no error...")
	go put(time.Duration(2), 8888)
	go get(3, 8888, true)
	<-w
	fmt.Println("  ...PASSED")

	fmt.Println("Test block get with timeout, error should show up...")
	go put(time.Duration(4), 8888)
	go get(3, 8888, false)
	<-w
	fmt.Println("  ...PASSED")
}
