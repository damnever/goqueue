/*
  A thread safe queue for Golang.
*/
package goqueue

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

type EmptyQueueError struct{}

type FullQueueError struct{}

func (err *EmptyQueueError) Error() string {
	return "Queue is Empty"
}

func (err *FullQueueError) Error() string {
	return "Queue is Full"
}

type Queue struct {
	mutex          sync.Mutex
	getCond        *sync.Cond
	putCond        *sync.Cond
	taskDoneCond   *sync.Cond
	list           *list.List
	maxSize        int
	unfinishedTask int
}

func NewQueue(maxSize int) *Queue {
	queue := new(Queue)
	queue.mutex = sync.Mutex{}
	queue.getCond = sync.NewCond(&queue.mutex)
	queue.putCond = sync.NewCond(&queue.mutex)
	queue.taskDoneCond = sync.NewCond(&queue.mutex)
	queue.list = list.New()
	queue.maxSize = maxSize
	queue.unfinishedTask = 0
	return queue
}

func (queue *Queue) Size() int {
	queue.mutex.Lock()
	size := queue.qsize()
	queue.mutex.Unlock()
	return size
}

func (queue *Queue) IsEmpty() bool {
	queue.mutex.Lock()
	isEmpty := (queue.qsize() == 0)
	queue.mutex.Unlock()
	return isEmpty
}

func (queue *Queue) IsFull() bool {
	queue.mutex.Lock()
	isFull := (queue.maxSize > 0 && queue.qsize() == queue.maxSize)
	queue.mutex.Unlock()
	return isFull
}

func (queue *Queue) GetNoWait() (*list.Element, error) {
	return queue.Get(false, 0)
}

func (queue *Queue) Get(block bool, timeout float64) (*list.Element, error) {
	queue.getCond.L.Lock()
	defer queue.getCond.L.Unlock()
	emptyQ := false
	if !block {
		if queue.qsize() == 0 {
			emptyQ = true
		}
	} else if timeout == float64(0) {
		for queue.qsize() == 0 {
			queue.getCond.Wait()
		}
	} else if timeout < float64(0) {
		return &list.Element{}, errors.New("'timeout' must be a non-negative number")
	} else {
		timer := time.After(time.Duration(timeout) * time.Second)
		notEmpty := make(chan bool)
		go queue.waitSignal(queue.getCond, notEmpty)
		for queue.qsize() == 0 {
			select {
			case <-timer:
				emptyQ = true
				break
			case <-notEmpty:
				if queue.qsize() == 0 {
					go queue.waitSignal(queue.getCond, notEmpty)
				} else {
					break
				}
			}
		}
	}
	if emptyQ {
		return &list.Element{}, &EmptyQueueError{}
	}
	e := queue.list.Front()
	queue.list.Remove(e)
	queue.putCond.Signal()
	return e, nil
}

func (queue *Queue) PutNoWait(element interface{}) (*list.Element, error) {
	return queue.Put(element, false, 0)
}

func (queue *Queue) Put(element interface{}, block bool, timeout float64) (*list.Element, error) {
	queue.putCond.L.Lock()
	defer queue.putCond.L.Unlock()
	fullQ := false
	if queue.maxSize > 0 {
		if !block {
			if queue.qsize() == queue.maxSize {
				fullQ = true
			}
		} else if timeout == float64(0) {
			for queue.qsize() == queue.maxSize {
				queue.putCond.Wait()
			}
		} else if timeout < float64(0) {
			return &list.Element{}, errors.New("'timeout' must be a non-negative number")
		} else {
			timer := time.After(time.Duration(timeout) * time.Second)
			notFull := make(chan bool)
			go queue.waitSignal(queue.putCond, notFull)
			for queue.qsize() == queue.maxSize {
				select {
				case <-timer:
					fullQ = true
					break
				case <-notFull:
					if queue.qsize() == queue.maxSize {
						go queue.waitSignal(queue.putCond, notFull)
					} else {
						break
					}
				}
			}
		}
	}
	if fullQ {
		return &list.Element{}, &FullQueueError{}
	}
	e := queue.list.PushBack(element)
	queue.unfinishedTask += 1
	queue.getCond.Signal()
	return e, nil
}

func (queue *Queue) TaskDone() {
	queue.taskDoneCond.L.Lock()
	defer queue.taskDoneCond.L.Unlock()
	unfinished := queue.unfinishedTask - 1
	if unfinished <= 0 {
		if unfinished < 0 {
			panic(errors.New("TaskDone() called too many times"))
		}
		queue.taskDoneCond.Broadcast()
	}
	queue.unfinishedTask = unfinished
}

func (queue *Queue) WaitAllComplete() {
	queue.taskDoneCond.L.Lock()
	defer queue.taskDoneCond.L.Unlock()
	for queue.unfinishedTask != 0 {
		queue.taskDoneCond.Wait()
	}
}

func (queue *Queue) waitSignal(cond *sync.Cond, c chan<- bool) {
	cond.Wait()
	c <- true
}

func (queue *Queue) qsize() int {
	return queue.list.Len()
}
