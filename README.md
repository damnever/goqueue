## A GoRoutine safe queue for Golang [![Build Status](https://travis-ci.org/damnever/goqueue.svg?branch=master)](https://travis-ci.org/damnever/goqueue) [![GoDoc](https://godoc.org/github.com/damnever/goqueue?status.svg)](https://godoc.org/github.com/damnever/goqueue)

It is similar to the Queue of Python.

### Installation

```
go get github.com/damnever/goqueue
```

### Example

Just for example, I use `Queue.Get(0)` and `Queue.PutNoWait(value)` more and often, but channel is not a right way do that...

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

    go func() {
        defer wg.Done()
        for i := 0; i <= 27; i += 3 {
            queue.PutNoWait(i)
        }
    }()

    for i := 0; i < 5; i++ {
        go worker(queue)
    }

    wg.Add(6)
    wg.Wait()

    fmt.Println("All task done!!!")
}
```
