## A thread safe queue for Golang [![Build Status](https://travis-ci.org/Damnever/goqueue.svg?branch=master)](https://travis-ci.org/Damnever/goqueue) [![GoDoc](https://godoc.org/github.com/Damnever/goqueue?status.svg)(https://godoc.org/github.com/Damnever/goqueue)]

Golang channel can not support infinite size, I use it as a replacement.

It is similar to the Queue of Python, as a practice...

**NOTE**: Maybe a lots of BUGs in the code...

### Installation

```
go get github.com/Damnever/goqueue
```

### Example

Just for example, I use `Queue.Get(true, 0)` and `Queue.PutNoWait(value)` more and often, but channel can not do that...

```Go
package main

import (
	"fmt"
	"sync"

	"github.com/Damnever/goqueue"
)

func main() {
	queue := goqueue.New(0)
	wg := &sync.WaitGroup{}

	worker := func(queue *goqueue.Queue) {
		defer wg.Done()
		for !queue.IsEmpty() {
			val, err := queue.Get(0)
			if err != nil {
				fmt.Println("Unexpect Error: %v\n", err)
			}
			num := val.(int)
			fmt.Printf("-> %v\n", num)
			if num%3 == 0 {
				for i := num + 1; i < num+3; i++ {
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

	wg.Add(5)
	wg.Wait()

	fmt.Println("All task done!!!")
}
```
