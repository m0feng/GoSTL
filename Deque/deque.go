package Deque

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Deque represents a highly optimized double-ended queue (deque) data structure.
type Deque[T any] struct {
	data    unsafe.Pointer // pointer to slice header (atomic access)
	front   int32          // atomic access
	back    int32          // atomic access
	length  int32          // atomic access
	mu      sync.Mutex     // only for resize operations
	initCap int            // initial capacity
}

type sliceHeader struct {
	data unsafe.Pointer
	len  int
	cap  int
}

// NewDeque creates and initializes a new Deque with optional initial capacity.
func NewDeque[T any](initCap ...int) *Deque[T] {
	q := &Deque[T]{}
	capacity := 8
	if len(initCap) > 0 && initCap[0] > 0 {
		capacity = initCap[0]
	}
	q.initCap = capacity
	q.Init(capacity)
	return q
}

// Init initializes or resets the deque.
func (q *Deque[T]) Init(n int) {
	capacity := 8
	if n > capacity {
		capacity = n
	}
	q.initCap = capacity
	data := make([]T, capacity)
	header := (*sliceHeader)(unsafe.Pointer(&data))
	atomic.StorePointer(&q.data, unsafe.Pointer(header))
	atomic.StoreInt32(&q.front, 0)
	atomic.StoreInt32(&q.back, 0)
	atomic.StoreInt32(&q.length, 0)
}

// Format implements the fmt.Formatter interface.
func (q *Deque[T]) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v', 's':
		length := int(atomic.LoadInt32(&q.length))
		if length == 0 {
			_, _ = io.WriteString(f, "[]")
			return
		}

		// Get display limit rules:
		// 1. Precision (%.3v) has the highest priority
		// 2. Width (%5v) comes next
		// 3. Default show all
		limit := length
		if p, ok := f.Precision(); ok { // %.3v style
			limit = p
		} else if w, ok := f.Width(); ok && w > 0 { // %5v style
			// For width, we use heuristic: show max w/2 elements
			limit = w / 2
			if limit < 3 {
				limit = 3 // Minimum 3 elements to show
			}
		}

		// Always show at least 1 element and at most all elements
		if limit <= 0 {
			limit = 1
		} else if limit > length {
			limit = length
		}

		// Get the elements to display
		elements := make([]T, limit)
		for i := 0; i < limit; i++ {
			elements[i], _ = q.At(i)
		}

		// Build the output string
		var b strings.Builder
		b.WriteByte('[')
		for i, val := range elements {
			if i > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(fmt.Sprint(val))
		}

		// Add ellipsis if we truncated
		if limit < length {
			b.WriteString(fmt.Sprintf(" ...+%d", length-limit))
		}
		b.WriteByte(']')

		_, _ = io.WriteString(f, b.String())
	default:
		_, _ = fmt.Fprintf(f, "%%!%c(deque)", verb)
	}
}

// Empty returns true if the deque contains no elements.
func (q *Deque[T]) Empty() bool {
	return atomic.LoadInt32(&q.length) == 0
}

// Len returns the number of elements in the deque.
func (q *Deque[T]) Len() int {
	return int(atomic.LoadInt32(&q.length))
}

// internalResize resizes the deque (must be called with lock held)
func (q *Deque[T]) internalResize(newCap int) {
	oldHeader := (*sliceHeader)(atomic.LoadPointer(&q.data))
	front := atomic.LoadInt32(&q.front)
	back := atomic.LoadInt32(&q.back)
	length := atomic.LoadInt32(&q.length)

	newData := make([]T, newCap)
	newHeader := (*sliceHeader)(unsafe.Pointer(&newData))

	if front < back {
		copy(newData, (*[1 << 30]T)(oldHeader.data)[front:back])
	} else {
		n := copy(newData, (*[1 << 30]T)(oldHeader.data)[front:oldHeader.cap])
		copy(newData[n:], (*[1 << 30]T)(oldHeader.data)[:back])
	}

	atomic.StorePointer(&q.data, unsafe.Pointer(newHeader))
	atomic.StoreInt32(&q.front, 0)
	atomic.StoreInt32(&q.back, length)
}

// PushBack adds an element to the back of the deque.
func (q *Deque[T]) PushBack(val T) {
	for {
		back := atomic.LoadInt32(&q.back)
		length := atomic.LoadInt32(&q.length)
		header := (*sliceHeader)(atomic.LoadPointer(&q.data))
		capacity := int32(header.cap)

		if length < capacity {
			newBack := (back + 1) % capacity
			if atomic.CompareAndSwapInt32(&q.back, back, newBack) {
				(*[1 << 30]T)(header.data)[back] = val
				atomic.AddInt32(&q.length, 1)
				return
			}
			continue
		}

		// Need to resize
		q.mu.Lock()
		// Double check after acquiring lock
		header = (*sliceHeader)(atomic.LoadPointer(&q.data))
		if atomic.LoadInt32(&q.length) == int32(header.cap) {
			newCap := header.cap * 2
			if newCap == 0 {
				newCap = q.initCap
			}
			q.internalResize(newCap)
			header = (*sliceHeader)(atomic.LoadPointer(&q.data))
		}
		back = atomic.LoadInt32(&q.back)
		capacity = int32(header.cap)
		(*[1 << 30]T)(header.data)[back] = val
		atomic.StoreInt32(&q.back, (back+1)%capacity)
		atomic.AddInt32(&q.length, 1)
		q.mu.Unlock()
		return
	}
}

// PushFront adds an element to the front of the deque.
func (q *Deque[T]) PushFront(val T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	header := (*sliceHeader)(atomic.LoadPointer(&q.data))
	if atomic.LoadInt32(&q.length) == int32(header.cap) {
		newCap := header.cap * 2
		if newCap == 0 {
			newCap = q.initCap
		}
		q.internalResize(newCap)
		header = (*sliceHeader)(atomic.LoadPointer(&q.data))
	}

	front := atomic.LoadInt32(&q.front)
	newFront := (front - 1 + int32(header.cap)) % int32(header.cap)
	(*[1 << 30]T)(header.data)[newFront] = val
	atomic.StoreInt32(&q.front, newFront)
	atomic.AddInt32(&q.length, 1)
}

// PopBack removes and returns the element from the back of the deque.
func (q *Deque[T]) PopBack() (T, bool) {
	var zero T
	for {
		length := atomic.LoadInt32(&q.length)
		if length == 0 {
			return zero, false
		}

		back := atomic.LoadInt32(&q.back)
		newBack := (back - 1 + int32(len(q.currentData()))) % int32(len(q.currentData()))
		if atomic.CompareAndSwapInt32(&q.back, back, newBack) {
			if atomic.AddInt32(&q.length, -1) >= 0 {
				return q.currentData()[newBack], true
			}
			// CAS failed, revert
			atomic.StoreInt32(&q.back, back)
			atomic.AddInt32(&q.length, 1)
		}
	}
}

// PopFront removes and returns the element from the front of the deque.
func (q *Deque[T]) PopFront() (T, bool) {
	var zero T
	for {
		length := atomic.LoadInt32(&q.length)
		if length == 0 {
			return zero, false
		}

		front := atomic.LoadInt32(&q.front)
		if atomic.CompareAndSwapInt32(&q.front, front, (front+1)%int32(len(q.currentData()))) {
			if atomic.AddInt32(&q.length, -1) >= 0 {
				return q.currentData()[front], true
			}
			// CAS failed, revert
			atomic.StoreInt32(&q.front, front)
			atomic.AddInt32(&q.length, 1)
		}
	}
}

// currentData returns the current underlying slice (unsafe, for internal use only)
func (q *Deque[T]) currentData() []T {
	header := (*sliceHeader)(atomic.LoadPointer(&q.data))
	return *(*[]T)(unsafe.Pointer(header))
}

// Front returns the element at the front of the deque without removing it.
func (q *Deque[T]) Front() (T, bool) {
	var zero T
	length := atomic.LoadInt32(&q.length)
	if length == 0 {
		return zero, false
	}
	front := atomic.LoadInt32(&q.front)
	return q.currentData()[front], true
}

// Back returns the element at the back of the deque without removing it.
func (q *Deque[T]) Back() (T, bool) {
	var zero T
	length := atomic.LoadInt32(&q.length)
	if length == 0 {
		return zero, false
	}
	back := atomic.LoadInt32(&q.back)
	data := q.currentData()
	return data[(back-1+int32(len(data)))%int32(len(data))], true
}

// Capacity returns the current capacity of the deque.
func (q *Deque[T]) Capacity() int {
	header := (*sliceHeader)(atomic.LoadPointer(&q.data))
	return header.cap
}

// Clear removes all elements from the deque.
func (q *Deque[T]) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	atomic.StoreInt32(&q.front, 0)
	atomic.StoreInt32(&q.back, 0)
	atomic.StoreInt32(&q.length, 0)
}

// At returns the element at the specified index.
func (q *Deque[T]) At(index int) (T, bool) {
	var zero T
	length := atomic.LoadInt32(&q.length)
	if index < 0 {
		index += int(length)
	}
	if index < 0 || index >= int(length) {
		return zero, false
	}
	front := atomic.LoadInt32(&q.front)
	data := q.currentData()
	return data[(front+int32(index))%int32(len(data))], true
}

// ShrinkToFit reduces capacity to fit the current size.
func (q *Deque[T]) ShrinkToFit() {
	q.mu.Lock()
	defer q.mu.Unlock()

	length := atomic.LoadInt32(&q.length)
	if length == 0 {
		q.Init(q.initCap)
		return
	}

	header := (*sliceHeader)(atomic.LoadPointer(&q.data))
	if int(length) == header.cap {
		return
	}

	newCap := int(length)
	if newCap < q.initCap {
		newCap = q.initCap
	}
	q.internalResize(newCap)
}

// Copy creates a new independent copy of the deque.
func (q *Deque[T]) Copy() *Deque[T] {
	q.mu.Lock()
	defer q.mu.Unlock()

	newDeque := NewDeque[T](q.Capacity())
	length := atomic.LoadInt32(&q.length)
	if length == 0 {
		return newDeque
	}

	header := (*sliceHeader)(atomic.LoadPointer(&q.data))
	front := atomic.LoadInt32(&q.front)
	back := atomic.LoadInt32(&q.back)
	capacity := header.cap

	data := make([]T, length)
	if front < back {
		copy(data, (*[1 << 30]T)(header.data)[front:back])
	} else {
		n := copy(data, (*[1 << 30]T)(header.data)[front:capacity])
		copy(data[n:], (*[1 << 30]T)(header.data)[:back])
	}

	newHeader := (*sliceHeader)(unsafe.Pointer(&data))
	atomic.StorePointer(&newDeque.data, unsafe.Pointer(newHeader))
	atomic.StoreInt32(&newDeque.front, 0)
	atomic.StoreInt32(&newDeque.back, length)
	atomic.StoreInt32(&newDeque.length, length)
	return newDeque
}

// Set sets the element at the specified index to the given value.
func (q *Deque[T]) Set(index int, value T) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	length := atomic.LoadInt32(&q.length)
	if index < 0 {
		index += int(length)
	}
	if index < 0 || index >= int(length) {
		return false
	}

	header := (*sliceHeader)(atomic.LoadPointer(&q.data))
	front := atomic.LoadInt32(&q.front)
	pos := (front + int32(index)) % int32(header.cap)
	(*[1 << 30]T)(header.data)[pos] = value
	return true
}

// Swap swaps the elements at the specified indices.
func (q *Deque[T]) Swap(i, j int) bool {
	if i == j {
		return false
	}
	q.mu.Lock()
	defer q.mu.Unlock()

	length := atomic.LoadInt32(&q.length)
	if i < 0 {
		i += int(length)
	}
	if j < 0 {
		j += int(length)
	}
	if i < 0 || i >= int(length) || j < 0 || j >= int(length) {
		return false
	}

	header := (*sliceHeader)(atomic.LoadPointer(&q.data))
	front := atomic.LoadInt32(&q.front)
	capacity := int32(header.cap)

	posI := (front + int32(i)) % capacity
	posJ := (front + int32(j)) % capacity

	// Swap the elements
	data := (*[1 << 30]T)(header.data)
	data[posI], data[posJ] = data[posJ], data[posI]
	return true
}

// reverseSlice reverses elements in slice s from index start to end-1
func (q *Deque[T]) reverseSlice(s []T, start, end int) {
	for i, j := start, end-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// Rotate rotates the deque by n positions to the right (positive n) or left (negative n).
func (q *Deque[T]) Rotate(n int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Load all atomic values once at the beginning
	length := int(atomic.LoadInt32(&q.length))
	if length <= 1 {
		return
	}

	// Normalize n to be within [0, length)
	n = n % length
	if n < 0 {
		n += length
	}
	if n == 0 {
		return
	}

	header := (*sliceHeader)(atomic.LoadPointer(&q.data))
	capacity := header.cap
	front := int(atomic.LoadInt32(&q.front))
	data := (*[1 << 30]T)(header.data)[:capacity]

	// Unified approach that works for both cases
	q.reverseCircular(data, front, front+length, length)
	q.reverseCircular(data, front, front+n, n)
	q.reverseCircular(data, front+n, front+length, length-n)
}

// reverseCircular reverses count elements in circular buffer starting at start (mod capacity)
func (q *Deque[T]) reverseCircular(data []T, start, end, count int) {
	capacity := cap(data)
	for i := 0; i < count/2; i++ {
		left := (start + i) % capacity
		right := (end - 1 - i) % capacity
		data[left], data[right] = data[right], data[left]
	}
}

// Reverse reverses the order of elements in the deque.
func (q *Deque[T]) Reverse() {
	q.mu.Lock()
	defer q.mu.Unlock()

	length := atomic.LoadInt32(&q.length)
	if length <= 1 {
		return
	}

	header := (*sliceHeader)(atomic.LoadPointer(&q.data))
	front := atomic.LoadInt32(&q.front)
	capacity := int32(header.cap)
	data := (*[1 << 30]T)(header.data)

	// Reverse in-place
	for i := 0; i < int(length)/2; i++ {
		left := (front + int32(i)) % capacity
		right := (front + length - 1 - int32(i)) % capacity
		data[left], data[right] = data[right], data[left]
	}
}
