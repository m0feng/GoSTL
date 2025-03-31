package main

import (
	"GoSTL/Queue"
	"fmt"
)

func main() {
	q := queue.NewQueue[int]()
	for i := 0; i < 1e8; i++ {
		q.Push(i)
	}
	for i := 0; i < 1e8-1; i++ {
		q.Pop()
	}
	fmt.Printf("%v", q)
}
