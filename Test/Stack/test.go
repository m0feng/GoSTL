package main

import (
	"GoSTL/Stack"
	"fmt"
	"time"
)

func main() {
	time1 := time.Now()
	st := Stack.NewStack[int]()
	for i := 0; i < 1e8; i++ {
		st.Push(i)
	}
	for i := 0; i < 1e8-2; i++ {
		st.Pop()
	}
	fmt.Println(st)
	time2 := time.Now()
	fmt.Printf("Time elapsed: %v\n", time2.Sub(time1))
}
