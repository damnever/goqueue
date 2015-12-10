## A thread safe queue for Golang

Golang channel can not support infinite size, I use it as a replacement.

It is very very very similar to the Queue of Python, as a practice...

**NOTE**:
 - I am a newbie, I know nothing about goroutine lifecycle...
 - Maybe a lots of BUGs in the code...

### Installation

```
go get github.com/Damnever/goqueue
```

### Example

Just for example, I use `Queue.Get(true, 0)` and `Queue.PutNoWait(value)` more and often, but channel can not do that...

```Go
package main

import "github.com/Damnever/goqueue"
import "fmt"

func main() {
	queue := goqueue.NewQueue(0)

	worker := func(queue *goqueue.Queue) {
        for {
		    e, err := queue.Get(true, 0)
		    if err != nil {
			    fmt.Println("Unexpect Error: %v\n", err)
		    }
            num := e.Value.(int)
            fmt.Printf("-> %v\n", num)
		    queue.TaskDone()
            if num % 3 == 0 {
                for i := num + 1; i < num + 3; i ++ {
                     queue.PutNoWait(i)
                }
            }
        }
	}

    for i := 0; i <= 27; i += 3 {
		queue.PutNoWait(i)
    }

    for i := 0; i < 5; i++ {
        go worker(queue)
    }

    queue.WaitAllComplete()
	fmt.Println("All task done!!!")
}
```

### Function and methods

 - `NewQueue(maxSize int) *Queue`

   Create a Queue. If `maxSize` is zero, return a Queue with infinite size.

Queue Methods:

 - `(queue *Queue) Size() int`

   Return the Queue size.

 - `(queue *Queue) IsEmpty() bool`

   Return `true` if Queue is empty, return `false` otherwise.

 - `(queue *Queue) IsFull() bool`

   Return `true` if Queue is full(as long as `maxSize` is not zero), return `false` otherwise.

 - `(queue *Queue) Get(block bool, timeout float64) (*list.Element, error)`

   If `block` is false, return `(&list.Element{Value: ...}, nil)` immediately if Queue is not empty, return `(&list.Element{}, EmptyQueueError)` otherwise.

   If `block` is `true` and `timeout` is zero, wait until an element available.

   If `block` is `true` and `timeout` greater than zero, return immediately if an element available before timeout, return `EmptyQueueError` if no element is available within `timeout`.

 - `(queue *Queue) GetNoWait() (*list.Element, error)`

   Same as `Get` with `block=false`.

 - `(queue *Queue) Put(value interface{}, block bool, timeout float64) (*list.Element, error)`

   If `maxSize` is zero, put `value` into Queue immediately.

   If `maxSize` is not zero:
   (1) If `block` is false, put `value` into Queue immediately and return `(&list.Element{Value: ...}, nil)` if Queue is not full, return `(&list.Element{}, FullQueueError)` otherwise.
   (2) If `block` is `true` and `timeout` is zero, wait until free slot available.
   (3) If `block` is `true` and `timeout` greater than zero, put `value` into Queue and return immediately if a slot available before timeout, return `EmptyQueueError` if no free slot is available within `timeout`.

 - `(queue *Queue) PutNoWait(value interface{}) (*list.Element, error)`

   Same as `Put` with `block=false`.

 - `(queue *Queue) TaskDone()`

   If you use `WaitAllComplete` method, you must call this method after you invoked `Get`/`GetNoWait` to indicate a formerly enqueue task is done.

   If you called this method more times than number of elements in Queue, then an error will `panic`...

 - `(queue *Queue) WaitAllComplete()`

   Blocks until all elements in the Queue have been gotten and processed.

Errors:

 - `type EmptyQueueError struct{}`

 - `type FullQueueError struct{}`

---
