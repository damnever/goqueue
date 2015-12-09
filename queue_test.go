package goqueue

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestUnBlockGetPut(t *testing.T) {
	value := 8888

	fmt.Println("Test unblock get/put, should no error...")
	queue := NewQueue(1)
	queue.PutNoWait(value)
	e, err := queue.GetNoWait()
	if err != nil {
		t.Fatalf("Unexpect error: %v\n", err)
	} else if e.Value.(int) != value {
		t.Fatalf("Expect %v, got %v\n", value, e.Value.(int))
	}
	fmt.Println("  ...PASSED")

	fmt.Println("Test unblock get/put, error should show up...")
	emptyQ := NewQueue(1)
	_, err1 := emptyQ.GetNoWait()
	if err1 == nil {
		t.Fatalf("No error show up\n")
	}
	fmt.Println("  ...PASSED")
}

func TestBlockGetWithTimeout(t *testing.T) {
	wg := &sync.WaitGroup{}
	queue := NewQueue(1)

	get := func(timeout float64, value int, noErr bool) {
		defer wg.Done()
		e, err := queue.Get(true, timeout)
		if noErr {
			if err == nil && e.Value.(int) != value {
				t.Fatalf("Expect %v, got %v\n", value, e.Value.(int))
			} else if err != nil {
				t.Fatalf("Unexpect error: %v\n", err)
			}
		} else {
			if err == nil {
				t.Fatalf("Wanted EmptyQueueError, but got nil\n")
			} else if reflect.TypeOf(err) != reflect.TypeOf(&EmptyQueueError{}) {
				t.Fatalf("Expect %v, got %v\n", EmptyQueueError{}, err)
			}
		}
	}

	put := func(d time.Duration, value int) {
		time.Sleep(time.Second * d)
		queue.PutNoWait(value)
	}

	fmt.Println("Test block get without timeout, wait until an element available...")
	wg.Add(1)
	go put(time.Duration(3), 9999)
	go get(0, 9999, true)
	wg.Wait()
	fmt.Println("  ...PASSED")

	fmt.Println("Test block get without timeout, wait forever...")
	done := make(chan bool)
	queue2 := NewQueue(1)
	go func() {
		queue2.Get(true, 0)
		done <- true
	}()
	select {
	case <-done:
		t.Fatalf("Queue.Get returned even though no element in Queue\n")
	case <-time.After(time.Duration(2) * time.Second):
	}
	fmt.Println("  ...PASSED")
	queue2.PutNoWait(0)
	<-done
	close(done)

	fmt.Println("Test block get with timeout, if an element available before timeout, return immediately...")
	wg.Add(1)
	go put(time.Duration(2), 8888)
	go get(3, 8888, true)
	wg.Wait()
	fmt.Println("  ...PASSED")

	fmt.Println("Test block get with timeout, if no element is available within timeout, return EmptyQueueError...")
	wg.Add(1)
	go put(time.Duration(4), 7777)
	go get(2, 7777, false)
	wg.Wait()
	fmt.Println("  ...PASSED")
}

func TestBlockPutWithTimeout(t *testing.T) {
	queue := NewQueue(1)
	queue.PutNoWait(1111)
	wg := &sync.WaitGroup{}

	get := func(d time.Duration) {
		time.Sleep(time.Second * d)
		queue.GetNoWait()
	}

	put := func(timeout float64, value int, noErr bool) {
		defer wg.Done()
		e, err := queue.Put(value, true, timeout)
		if noErr {
			if err == nil && e.Value.(int) != value {
				t.Fatalf("Expect %v, got %v\n", value, e.Value.(int))
			} else if err != nil {
				t.Fatalf("Unexpect error: %v\n", err)
			}
		} else {
			if err == nil {
				t.Fatalf("Wanted FullQueueError, but got nil\n")
			} else if reflect.TypeOf(err) != reflect.TypeOf(&FullQueueError{}) {
				t.Fatalf("Expect %v, got %v\n", FullQueueError{}, err)
			}
		}
	}

	fmt.Println("Test block put without timeout, wait until a free slot available...")
	go get(time.Duration(2))
	wg.Add(1)
	go put(0, 9999, true)
	wg.Wait()
	fmt.Println("  ...PASSED")

	fmt.Println("Test block put without timeout, wait forever...")
	queue2 := NewQueue(1)
	queue2.PutNoWait(1111)
	done := make(chan bool)
	go func() {
		queue2.Put(0, true, 2222)
		done <- true
	}()
	select {
	case <-done:
		t.Fatalf("Queue.Put returned even though no free slot in Queue\n")
	case <-time.After(time.Duration(2) * time.Second):
	}
	fmt.Println("  ...PASSED")
	queue2.GetNoWait()
	<-done

	fmt.Println("Test block put with timeout, if a free slot is available before timeout, return immediately...")
	go get(time.Duration(2))
	wg.Add(1)
	go put(3, 8888, true)
	wg.Wait()
	fmt.Println("  ...PASSED")

	fmt.Println("Test block put with timeout, if no free slot is available within timeout, return FullQueueError...")
	go get(time.Duration(4))
	wg.Add(1)
	go put(2, 7777, false)
	wg.Wait()
	fmt.Println("  ...PASSED")
}

func TestTaskDoneAndFIFO(t *testing.T) {
	size := 3
	queue := NewQueue(size)

	fmt.Println("Test Queue.TaskDone method and FIFO...")
	for i := 0; i < size; i++ {
		queue.PutNoWait(i)
	}
	for i := 0; i < size; i++ {
		e, err := queue.GetNoWait()
		if err != nil {
			t.Fatalf("Unexpect error: %v\n", err)
		} else if e.Value.(int) != i {
			t.Fatalf("Queue is not FIFO")
		}
		queue.TaskDone()
	}
	if queue.Size() != 0 {
		t.Fatalf("Something wrong, %v task(s) not done\n", queue.Size())
	}
	fmt.Println("  ...PASSED")
}

func TestWaitAllComplete(t *testing.T) {
	size := 20
	queue := NewQueue(size)

	fmt.Println("Test Queue.WaitAllComplete method...")
	go func() {
		for !queue.IsEmpty() {
			_, err := queue.Get(true, 0)
			if err != nil {
				t.Fatalf("Unexpect Error: %v\n", err)
			}
			queue.TaskDone()
		}
	}()
	for i := 0; i < size; i++ {
		queue.PutNoWait(i)
	}
	queue.WaitAllComplete()
	if queue.Size() != 0 {
		t.Fatalf("Somthing wrong, %v task(s) not done\n", queue.Size())
	}
	fmt.Println("  ...PASSED")
}
