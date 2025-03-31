package main

import (
	"GoSTL/Deque"
	"fmt"
	"time"
)

func main() {
	time1 := time.Now()
	st := Deque.NewDeque[int]()
	for i := 0; i < 1e8; i++ {
		st.PushBack(i)
	}
	for i := 0; i < 1e8-1; i++ {
		st.PopBack()
	}
	fmt.Println(st)
	for i := 0; i < 10; i++ {
		st.PushFront(i)
	}
	fmt.Println(st)
	fmt.Printf("%5v\n", st)
	time2 := time.Now()
	fmt.Printf("Time elapsed: %v\n", time2.Sub(time1))
}
