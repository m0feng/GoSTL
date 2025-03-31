package queue

import (
	"GoSTL/Deque"
	"fmt"
)

// Queue implements a FIFO (First-In-First-Out) data structure using a Deque as its underlying storage.
// It provides O(1) time complexity for push/pop operations at both ends.
type Queue[T any] struct {
	d *Deque.Deque[T] // underlying deque that stores the queue elements
}

// NewQueue creates and initializes a new Queue with an initial capacity of 8.
// The queue uses a deque internally for efficient operations at both ends.
// Returns a pointer to the newly created Queue.
func NewQueue[T any]() *Queue[T] {
	Q := &Queue[T]{d: Deque.NewDeque[T]()}
	return Q
}

// Init initializes or clears the queue with the specified initial capacity.
// This operation will remove all existing elements in the queue.
func (q *Queue[T]) Init(n int) {
	q.d.Init(n)
}

// Pop removes and returns the front element of the queue (FIFO operation).
// Panics if the queue is empty.
func (q *Queue[T]) Pop() (T, bool) {
	return q.d.PopFront()
}

// Front returns the front element of the queue without removing it.
// Panics if the queue is empty.
func (q *Queue[T]) Front() (T, bool) {
	return q.d.Front()
}

// Push adds an element to the back of the queue.
func (q *Queue[T]) Push(value T) {
	q.d.PushBack(value)
}

// Len returns the number of elements in the queue.
func (q *Queue[T]) Len() int {
	return q.d.Len()
}

// Empty returns true if the queue contains no elements.
func (q *Queue[T]) Empty() bool {
	return q.d.Empty()
}

// Capacity returns the current capacity of the underlying storage.
func (q *Queue[T]) Capacity() int {
	return q.d.Capacity()
}

// At returns the element at the specified index (0 = front, Len()-1 = back).
// Supports negative indices (-1 = last element, -2 = second last, etc.).
// Panics if index is out of range.
func (q *Queue[T]) At(index int) (T, bool) {
	return q.d.At(index)
}

// Set updates the element at the specified index.
// Supports negative indices (-1 = last element, -2 = second last, etc.).
// Panics if index is out of range.
func (q *Queue[T]) Set(index int, value T) bool {
	return q.d.Set(index, value)
}

// Swap exchanges the elements at indices i and j.
// Panics if either index is out of range.
func (q *Queue[T]) Swap(i, j int) bool {
	return q.d.Swap(i, j)
}

// Rotate rotates the queue elements by n positions:
// - Positive n rotates right (back elements move to front)
// - Negative n rotates left (front elements move to back)
func (q *Queue[T]) Rotate(n int) {
	q.d.Rotate(n)
}

// ShrinkToFit reduces the queue's capacity to fit its current length.
// This may help reduce memory usage for queues that have grown large but now contain few elements.
func (q *Queue[T]) ShrinkToFit() {
	q.d.ShrinkToFit()
}

// Copy creates a deep copy of the queue with the same elements and capacity.
func (q *Queue[T]) Copy() *Queue[T] {
	newDeque := q.d.Copy()
	newQueue := &Queue[T]{
		d: newDeque,
	}
	return newQueue
}

// Reverse reverses the order of elements in the queue in-place.
// After reversal, the former front becomes the back and vice versa.
func (q *Queue[T]) Reverse() {
	q.d.Reverse()
}

// Clear removes all elements from the queue while maintaining its current capacity.
func (q *Queue[T]) Clear() {
	q.d.Clear()
}

// String returns a string representation of the queue's elements.
// The format is similar to a slice representation.
func (q *Queue[T]) String() string {
	return fmt.Sprintf("%v", &q.d)
}

// Format implements custom formatting for the queue.
// Supports all the same format verbs as fmt.Print functions.
func (q *Queue[T]) Format(f fmt.State, verb rune) {
	q.d.Format(f, verb)
}
