/*
  A thread safe queue written in Go.
  Golang do not support infinite size channel.
  It can get element block with timeout.
*/
package goqueue

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

type Queue struct {
	mutex   sync.Mutex
	getCond *sync.Cond
	putCond *sync.Cond
	list    *list.List
}

func NewQueue() *Queue {
	queue := new(Queue)
	queue.mutex = sync.Mutex{}
	queue.getCond = sync.NewCond(&queue.mutex)
	queue.putCond = sync.NewCond(&queue.mutex)
	queue.list = list.New()
	return queue
}

func (queue *Queue) Size() int {
	queue.mutex.Lock()
	size := queue.list.Len()
	queue.mutex.Unlock()
	return size
}

func (queue *Queue) IsEmpty() bool {
	queue.mutex.Lock()
	isEmpty := queue.isEmpty()
	queue.mutex.Unlock()
	return isEmpty
}

func (queue *Queue) Get(block bool, timeout float64) (*list.Element, error) {
	queue.getCond.L.Lock()
	defer queue.getCond.L.Unlock()
	emptyQ := false
	if !block {
		if queue.isEmpty() {
			emptyQ = true
		}
	} else if timeout == float64(0) {
		for queue.isEmpty() {
			queue.getCond.Wait()
		}
	} else if timeout < float64(0) {
		return &list.Element{}, errors.New("'timeout' must be a non-negative number")
	} else {
		timer := time.After(time.Duration(timeout) * time.Second)
		notEmpty := make(chan bool)
		go queue.wait(notEmpty)
		for queue.isEmpty() {
			select {
			case <-timer:
				emptyQ = true
				break
			case <-notEmpty:
				if queue.isEmpty() {
					go queue.wait(notEmpty)
				} else {
					break
				}
			}
		}
	}
	if emptyQ {
		return &list.Element{}, errors.New("Empty Queue")
	}
	e := queue.list.Front()
	queue.list.Remove(e)
	return e, nil
}

func (queue *Queue) Put(element interface{}) *list.Element {
	queue.putCond.L.Lock()
	defer queue.putCond.L.Unlock()
	e := queue.list.PushBack(element)
	queue.getCond.Signal()
	return e
}

func (queue *Queue) wait(c chan<- bool) {
	queue.getCond.Wait()
	c <- true
}

func (queue *Queue) isEmpty() bool {
	return (queue.list.Len() == 0)
}
