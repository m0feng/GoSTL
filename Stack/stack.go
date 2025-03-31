package Stack

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Stack is a generic LIFO (Last-In-First-Out) data structure implementation with thread-safe operations.
type Stack[T any] struct {
	data    unsafe.Pointer // atomic pointer to slice header
	top     int32          // atomic stack pointer
	mu      sync.Mutex     // only for resize operations
	initCap int            // initial capacity
}

type sliceHeader struct {
	data unsafe.Pointer
	len  int
	cap  int
}

// NewStack creates and initializes a new Deque with optional initial capacity.
func NewStack[T any](initCap ...int) *Stack[T] {
	q := &Stack[T]{}
	capacity := 8
	if len(initCap) > 0 && initCap[0] > 0 {
		capacity = initCap[0]
	}
	q.initCap = capacity
	q.Init(capacity)
	return q
}

// Init initializes or resets the stack with an initial capacity hint.
func (s *Stack[T]) Init(n int) {
	capacity := 8
	if n > capacity {
		capacity = n
	}
	s.initCap = capacity
	data := make([]T, capacity)
	header := (*sliceHeader)(unsafe.Pointer(&data))
	atomic.StorePointer(&s.data, unsafe.Pointer(header))
	atomic.StoreInt32(&s.top, 0)
}

// Format implements the fmt.Formatter interface.
func (s *Stack[T]) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v', 's':
		top := int(atomic.LoadInt32(&s.top))
		if top == 0 {
			_, _ = io.WriteString(f, "[]")
			return
		}

		// Get display limit from width
		limit := top
		if width, ok := f.Width(); ok && width > 0 {
			// Heuristic: show width/2 elements (minimum 3)
			limit = width / 2
			if limit < 3 {
				limit = 3
			}
		}

		// Ensure limit is valid
		if limit > top {
			limit = top
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		header := (*sliceHeader)(atomic.LoadPointer(&s.data))
		data := (*[1 << 30]T)(header.data)[:header.cap]

		var b strings.Builder
		b.WriteByte('[')

		// Show elements from top (newest) to oldest
		for i := 0; i < limit; i++ {
			if i > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(fmt.Sprint(data[top-1-i]))
		}

		// Add ellipsis if we truncated
		if limit < top {
			b.WriteString(fmt.Sprintf(" ...+%d", top-limit))
		}

		b.WriteByte(']')
		_, _ = io.WriteString(f, b.String())
	default:
		_, _ = fmt.Fprintf(f, "%%!%c(stack)", verb)
	}
}

// stringWithLimit generates the string representation with optional truncation.
func (s *Stack[T]) stringWithLimit(limit int) string {
	top := atomic.LoadInt32(&s.top)
	if top == 0 {
		return "[]"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	header := (*sliceHeader)(atomic.LoadPointer(&s.data))
	data := (*[1 << 30]T)(header.data)[:header.cap]

	var b strings.Builder
	b.WriteByte('[')

	showCount := int(top)
	if limit > 0 && showCount > limit {
		showCount = limit
	}

	for i := int(top) - 1; i >= int(top)-showCount; i-- {
		if i != int(top)-1 {
			b.WriteByte(' ')
		}
		b.WriteString(fmt.Sprint(data[i]))
	}

	if limit > 0 && int(top) > limit {
		_, _ = fmt.Fprintf(&b, " ...+%d", int(top)-limit)
	}

	b.WriteByte(']')
	return b.String()
}

// Empty returns true if the stack contains no elements.
func (s *Stack[T]) Empty() bool {
	return atomic.LoadInt32(&s.top) == 0
}

// internalResize resizes the stack (must be called with lock held)
func (s *Stack[T]) internalResize(newCap int) {
	oldHeader := (*sliceHeader)(atomic.LoadPointer(&s.data))
	top := atomic.LoadInt32(&s.top)

	newData := make([]T, newCap)
	newHeader := (*sliceHeader)(unsafe.Pointer(&newData))
	copy(newData, (*[1 << 30]T)(oldHeader.data)[:top])

	atomic.StorePointer(&s.data, unsafe.Pointer(newHeader))
}

// Push adds an element to the top of the stack.
func (s *Stack[T]) Push(val T) {
	for {
		top := atomic.LoadInt32(&s.top)
		header := (*sliceHeader)(atomic.LoadPointer(&s.data))

		if int(top) < header.cap {
			if atomic.CompareAndSwapInt32(&s.top, top, top+1) {
				(*[1 << 30]T)(header.data)[top] = val
				return
			}
			continue
		}

		s.mu.Lock()
		header = (*sliceHeader)(atomic.LoadPointer(&s.data))
		if int(atomic.LoadInt32(&s.top)) == header.cap {
			newCap := header.cap * 2
			if newCap == 0 {
				newCap = s.initCap
			}
			s.internalResize(newCap)
			header = (*sliceHeader)(atomic.LoadPointer(&s.data))
		}
		top = atomic.LoadInt32(&s.top)
		(*[1 << 30]T)(header.data)[top] = val
		atomic.StoreInt32(&s.top, top+1)
		s.mu.Unlock()
		return
	}
}

// Pop removes and returns the element from the top of the stack.
func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	for {
		top := atomic.LoadInt32(&s.top)
		if top <= 0 {
			return zero, false
		}

		if atomic.CompareAndSwapInt32(&s.top, top, top-1) {
			header := (*sliceHeader)(atomic.LoadPointer(&s.data))
			return (*[1 << 30]T)(header.data)[top-1], true
		}
	}
}

// Top returns the top element without removing it.
func (s *Stack[T]) Top() (T, bool) {
	var zero T
	top := atomic.LoadInt32(&s.top)
	if top <= 0 {
		return zero, false
	}
	header := (*sliceHeader)(atomic.LoadPointer(&s.data))
	return (*[1 << 30]T)(header.data)[top-1], true
}

// Length returns the number of elements in the stack.
func (s *Stack[T]) Length() int {
	return int(atomic.LoadInt32(&s.top))
}

// Capacity returns the current capacity of the underlying storage.
func (s *Stack[T]) Capacity() int {
	header := (*sliceHeader)(atomic.LoadPointer(&s.data))
	return header.cap
}

// Clear removes all elements from the stack.
func (s *Stack[T]) Clear() {
	atomic.StoreInt32(&s.top, 0)
}

// Resize changes the stack's capacity.
func (s *Stack[T]) Resize(newCap int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	top := atomic.LoadInt32(&s.top)
	if newCap < int(top) {
		atomic.StoreInt32(&s.top, int32(newCap))
	}
	s.internalResize(newCap)
}

// TrimToSize reduces the stack's capacity to match its length.
func (s *Stack[T]) TrimToSize() {
	s.mu.Lock()
	defer s.mu.Unlock()

	top := atomic.LoadInt32(&s.top)
	header := (*sliceHeader)(atomic.LoadPointer(&s.data))

	if top == 0 {
		s.internalResize(s.initCap)
		return
	}

	if top < int32(header.cap) {
		s.internalResize(int(top))
	}
}

// Reverse reverses the elements in the stack.
func (s *Stack[T]) Reverse() {
	s.mu.Lock()
	defer s.mu.Unlock()

	top := int(atomic.LoadInt32(&s.top))
	header := (*sliceHeader)(atomic.LoadPointer(&s.data))
	data := (*[1 << 30]T)(header.data)[:header.cap]

	for i := 0; i < top/2; i++ {
		j := top - 1 - i
		data[i], data[j] = data[j], data[i]
	}
}

// Rotate rotates the stack by n positions.
// Positive n rotates right (top element moves n positions down)
// Negative n rotates left (top element moves |n| positions up)
func (s *Stack[T]) Rotate(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	top := int(atomic.LoadInt32(&s.top))
	if top <= 1 {
		return
	}

	n = n % top
	if n < 0 {
		n += top
	}
	if n == 0 {
		return
	}

	header := (*sliceHeader)(atomic.LoadPointer(&s.data))
	data := (*[1 << 30]T)(header.data)[:top]

	// 三次反转实现（完全原地）
	reverse := func(data []T, start, end int) {
		for i, j := start, end-1; i < j; i, j = i+1, j-1 {
			data[i], data[j] = data[j], data[i]
		}
	}

	reverse(data, 0, top) // 全反转：[0,1,2,3,4] → [4,3,2,1,0]
	reverse(data, 0, n)   // 反转前n个：[4,3,2,1,0] → [3,4,2,1,0]
	reverse(data, n, top) // 反转剩余：[3,4,2,1,0] → [3,4,0,1,2]
}

// Copy creates a new independent copy of the stack.
func (s *Stack[T]) Copy() *Stack[T] {
	s.mu.Lock()
	defer s.mu.Unlock()

	newStack := &Stack[T]{initCap: s.initCap}
	top := int(atomic.LoadInt32(&s.top))
	header := (*sliceHeader)(atomic.LoadPointer(&s.data))
	data := (*[1 << 30]T)(header.data)[:header.cap]

	newData := make([]T, top)
	copy(newData, data[:top])
	newHeader := (*sliceHeader)(unsafe.Pointer(&newData))
	atomic.StorePointer(&newStack.data, unsafe.Pointer(newHeader))
	atomic.StoreInt32(&newStack.top, int32(top))
	return newStack
}

// At returns the element at the specified index.
func (s *Stack[T]) At(index int) (T, bool) {
	var zero T
	top := int(atomic.LoadInt32(&s.top))
	if index < 0 {
		index += top
	}
	if index < 0 || index >= top {
		return zero, false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	header := (*sliceHeader)(atomic.LoadPointer(&s.data))
	data := (*[1 << 30]T)(header.data)[:header.cap]
	return data[top-1-index], true
}

// Set updates the element at the specified index with the given value.
func (s *Stack[T]) Set(index int, val T) bool {
	top := int(atomic.LoadInt32(&s.top))
	if index < 0 {
		index += top
	}
	if index < 0 || index >= top {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	header := (*sliceHeader)(atomic.LoadPointer(&s.data))
	data := (*[1 << 30]T)(header.data)[:header.cap]
	data[top-1-index] = val
	return true
}
