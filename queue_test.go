package goqueue

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestUnBlockGetPut(t *testing.T) {
	value := 8888

	fmt.Println("Test unblock get/put, should no error...")
	queue := New(1)
	queue.PutNoWait(value)
	val, err := queue.GetNoWait()
	if err != nil {
		t.Fatalf("Unexpect error: %v\n", err)
	} else if val.(int) != value {
		t.Fatalf("Expect %v, got %v\n", value, val.(int))
	}
	fmt.Println("  ...PASSED")

	fmt.Println("Test unblock get/put, error should show up...")
	queue = New(1)
	_, err = queue.GetNoWait()
	if err == nil {
		t.Fatalf("No error show up\n")
	}
	queue.PutNoWait(value)
	err = queue.PutNoWait(value + 1)
	if err == nil {
		t.Fatalf("No error show up\n")
	}
	val, err = queue.GetNoWait()
	if err != nil {
		t.Fatalf("Unexpect error: %v\n", err)
	} else if val.(int) != value {
		t.Fatalf("Expect %v, got %v\n", value, val.(int))
	}

	fmt.Println("  ...PASSED")
}

func TestFIFO(t *testing.T) {
	size := 3
	queue := New(size)

	fmt.Println("Test FIFO...")
	for i := 0; i < size; i++ {
		queue.PutNoWait(i)
	}
	for i := 0; i < size; i++ {
		val, err := queue.GetNoWait()
		if err != nil {
			t.Fatalf("Unexpect error: %v\n", err)
		} else if val.(int) != i {
			t.Fatalf("Queue is not FIFO")
		}
	}
	if queue.Size() != 0 {
		t.Fatalf("Something wrong, Queue has %d more items.\n", queue.Size())
	}
	fmt.Println("  ...PASSED")
}

func TestBlockGetWithTimeout(t *testing.T) {
	wg := &sync.WaitGroup{}
	queue := New(1)

	get := func(timeout float64, value int, noErr bool) {
		defer wg.Done()
		val, err := queue.Get(timeout)
		if noErr {
			if err == nil && val.(int) != value {
				t.Fatalf("Expect %v, got %v\n", value, val.(int))
			} else if err != nil {
				t.Fatalf("Unexpect error: %v\n", err)
			}
		} else {
			if err == nil {
				t.Fatalf("Wanted EmptyQueueError, but got nil\n")
			} else if err != ErrEmptyQueue {
				t.Fatalf("Expect %v, got %v\n", ErrEmptyQueue, err)
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
	queue2 := New(1)
	go func() {
		queue2.Get(0)
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
	queue := New(1)
	queue.PutNoWait(1111)
	wg := &sync.WaitGroup{}

	get := func(d time.Duration) {
		defer wg.Done()
		time.Sleep(time.Second * d)
		queue.GetNoWait()
	}

	put := func(timeout float64, value int, noErr bool) {
		defer wg.Done()
		err := queue.Put(value, timeout)
		if noErr {
			if err != nil {
				t.Fatalf("Unexpect error: %v\n", err)
			}
		} else {
			if err == nil {
				t.Fatalf("Wanted FullQueueError, but got nil\n")
			} else if err != ErrFullQueue {
				t.Fatalf("Expect %v, got %v\n", ErrFullQueue, err)
			}
		}
	}

	fmt.Println("Test block put without timeout, wait until a free slot available...")
	go get(time.Duration(2))
	go put(0, 9999, true)
	wg.Add(2)
	wg.Wait()
	fmt.Println("  ...PASSED")

	fmt.Println("Test block put without timeout, wait forever...")
	queue2 := New(1)
	queue2.PutNoWait(1111)
	done := make(chan bool)
	go func() {
		queue2.Put(0, 2222)
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
	go put(3, 8888, true)
	wg.Add(2)
	wg.Wait()
	fmt.Println("  ...PASSED")

	fmt.Println("Test block put with timeout, if no free slot is available within timeout, return FullQueueError...")
	go get(time.Duration(4))
	go put(2, 7777, false)
	wg.Add(2)
	wg.Wait()
	fmt.Println("  ...PASSED")
}

func TestConcurrentPutGet(t *testing.T) {
	queue := New(50)
	wg := &sync.WaitGroup{}

	fmt.Println("Test concurrent Get/Put...")
	go func() {
		defer wg.Done()
		dones := [10]chan bool{}
		for i := 0; i < 10; i++ {
			dones[i] = make(chan bool, 1)
			go func(i int) {
				for j := 1; j <= 100; j++ {
					queue.Put(i*j, 0)
				}
				dones[i] <- true
			}(i)
		}
		for i := 0; i < 10; i++ {
			<-dones[i]
		}
	}()

	go func() {
		defer wg.Done()
		dones := [9]chan bool{}
		for i := 0; i < 9; i++ {
			dones[i] = make(chan bool, 1)
			go func(i int) {
				for j := 1; j <= 100; j++ {
					queue.Get(0)
				}
				dones[i] <- true
			}(i)
		}
		for i := 0; i < 9; i++ {
			<-dones[i]
		}
		for i := 0; i < 50; i++ {
			queue.Get(0)
		}
	}()

	wg.Add(2)
	wg.Wait()

	if queue.Size() != 50 {
		t.Fatalf("Except Queue size %d, got %d.\n", 50, queue.Size())
	}
	if queue.putters.Len() != 0 {
		t.Fatalf("Not right, still has %d Put operators is pending.\n", queue.putters.Len())
	}
	if queue.getters.Len() != 0 {
		t.Fatalf("Not right, still has %d Get operators is pending.\n", queue.getters.Len())
	}

	for i := 0; i < 30; i++ {
		queue.Get(0)
	}
	if queue.Size() != 20 {
		t.Fatalf("Expect Queue size is %d, got %d.\n", 30, queue.Size())
	}

	for i := 0; i < 20; i++ {
		queue.GetNoWait()
	}
	if queue.Size() != 0 {
		t.Fatalf("Queue is not empty, still has %d items.\n", queue.Size())
	}
	fmt.Println("  ...PASSED")
}
