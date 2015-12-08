## A simple thread safe queue for Golang

Golang channel can not support infinite size, I use it as a replacement.

Queue Methods(maybe more):

 - `NewQueue() *Queue`
 - `IsEmpty() bool`
 - `Size() int`
 - `Get(block bool, timeout float64) (*list.Element, error)`
 - `Put(value interface) *list.Element`

---

NOTE: Maybe a lot of BUGs in the code...
