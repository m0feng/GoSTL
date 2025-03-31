package main_test

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	"GoSTL/Stack"
)

func TestNewStack(t *testing.T) {
	s := Stack.NewStack[int]()
	if s == nil {
		t.Fatal("NewStack returned nil")
	}
	if !s.Empty() {
		t.Error("New stack should be empty")
	}
	if s.Capacity() < 8 {
		t.Errorf("Expected capacity >= 8, got %d", s.Capacity())
	}
}

func TestPushPop(t *testing.T) {
	s := Stack.NewStack[int]()

	// Test basic push/pop
	for i := 0; i < 100; i++ {
		s.Push(i)
	}
	for i := 99; i >= 0; i-- {
		val, ok := s.Pop()
		if !ok || val != i {
			t.Errorf("Expected %d, got %d (ok: %v)", i, val, ok)
		}
	}

	// Test empty pop
	if _, ok := s.Pop(); ok {
		t.Error("Pop on empty stack should return false")
	}
}

func TestTop(t *testing.T) {
	s := Stack.NewStack[string]()

	// Test empty
	if _, ok := s.Top(); ok {
		t.Error("Top on empty stack should return false")
	}

	// Test with elements
	s.Push("first")
	if val, ok := s.Top(); !ok || val != "first" {
		t.Errorf("Expected 'first', got '%s' (ok: %v)", val, ok)
	}

	s.Push("second")
	if val, ok := s.Top(); !ok || val != "second" {
		t.Errorf("Expected 'second', got '%s' (ok: %v)", val, ok)
	}

	// Test after pop
	s.Pop()
	if val, ok := s.Top(); !ok || val != "first" {
		t.Errorf("Expected 'first', got '%s' (ok: %v)", val, ok)
	}
}

func TestLengthEmpty(t *testing.T) {
	s := Stack.NewStack[float64]()

	if !s.Empty() {
		t.Error("New stack should be empty")
	}

	for i := 0; i < 1000; i++ {
		s.Push(float64(i))
		if s.Length() != i+1 {
			t.Errorf("Expected length %d, got %d", i+1, s.Length())
		}
	}

	for i := 0; i < 1000; i++ {
		s.Pop()
		if s.Length() != 999-i {
			t.Errorf("Expected length %d, got %d", 999-i, s.Length())
		}
	}

	if !s.Empty() {
		t.Error("Stack should be empty after all pops")
	}
}

func TestResize(t *testing.T) {
	s := Stack.NewStack[int](4) // Small initial capacity to force resizes

	// Fill and check capacity growth
	initialCap := s.Capacity()
	for i := 0; i < 100; i++ {
		s.Push(i)
		if s.Length() != i+1 {
			t.Errorf("Length mismatch, expected %d got %d", i+1, s.Length())
		}
	}
	if s.Capacity() <= initialCap {
		t.Error("Capacity should have increased")
	}

	// Verify all elements
	for i := 99; i >= 0; i-- {
		val, ok := s.Pop()
		if !ok || val != i {
			t.Errorf("Expected %d, got %d (ok: %v)", i, val, ok)
		}
	}

	// Test explicit resize
	s.Resize(10)
	if s.Capacity() != 10 {
		t.Errorf("Expected capacity 10 after resize, got %d", s.Capacity())
	}
}

func TestAt(t *testing.T) {
	s := Stack.NewStack[rune]()

	// Fill with letters 'a' to 'z'
	for i := 0; i < 26; i++ {
		s.Push(rune('a' + i))
	}

	// Test positive indices (0 = top)
	for i := 0; i < 26; i++ {
		val, ok := s.At(i)
		expected := rune('z' - i)
		if !ok || val != expected {
			t.Errorf("At(%d) expected %c, got %c (ok: %v)", i, expected, val, ok)
		}
	}

	// Test negative indices
	for i := -1; i >= -26; i-- {
		val, ok := s.At(i)
		expected := rune('a' - i - 1)
		if !ok || val != expected {
			t.Errorf("At(%d) expected %c, got %c (ok: %v)", i, expected, val, ok)
		}
	}

	// Test out of bounds
	if _, ok := s.At(26); ok {
		t.Error("At(26) should be out of bounds")
	}
	if _, ok := s.At(-27); ok {
		t.Error("At(-27) should be out of bounds")
	}
}

func TestSet(t *testing.T) {
	s := Stack.NewStack[int]()

	// Fill with 0..9
	for i := 0; i < 10; i++ {
		s.Push(i)
	}

	// Set some values
	if !s.Set(2, 20) {
		t.Error("Set(2, 20) failed")
	}
	if !s.Set(-1, 90) { // last element
		t.Error("Set(-1, 90) failed")
	}

	// Verify
	if val, _ := s.At(2); val != 20 {
		t.Errorf("Expected 20 at index 2, got %d", val)
	}
	if val, _ := s.At(9); val != 90 {
		t.Errorf("Expected 90 at index 9, got %d", val)
	}

	// Test out of bounds
	if s.Set(10, 100) {
		t.Error("Set(10, 100) should fail")
	}
	if s.Set(-11, 100) {
		t.Error("Set(-11, 100) should fail")
	}
}

func TestReverse(t *testing.T) {
	s := Stack.NewStack[int]()

	// Test empty and single element
	s.Reverse() // should do nothing
	s.Push(1)
	s.Reverse()
	if val, _ := s.Top(); val != 1 {
		t.Error("Reverse on single element should not change it")
	}

	// Test multiple elements
	s.Clear()
	for i := 0; i < 5; i++ {
		s.Push(i)
	}
	s.Reverse()
	expected := []int{0, 1, 2, 3, 4} // stack top is now 0
	for i := 0; i < 5; i++ {
		val, _ := s.At(i)
		if val != expected[i] {
			t.Errorf("After reverse, At(%d) expected %d, got %d", i, expected[i], val)
		}
	}
}

func TestRotate(t *testing.T) {
	s := Stack.NewStack[int]()

	// Fill with 0..4
	for i := 0; i < 5; i++ {
		s.Push(i)
	}

	// Rotate right by 2 (top moves down 2)
	s.Rotate(2)                      //3 4 0 1 2
	expected := []int{2, 1, 0, 4, 3} // stack top is now 1
	for i := 0; i < 5; i++ {
		val, _ := s.At(i)
		if val != expected[i] {
			t.Errorf("After rotate(2), At(%d) expected %d, got %d", i, expected[i], val)
		}
	}
	//1 2 3 4 0
	// Rotate left by 3 (same as right by -3)
	s.Rotate(-3)
	expected = []int{0, 4, 3, 2, 1} // stack top is now 3
	for i := 0; i < 5; i++ {
		val, _ := s.At(i)
		if val != expected[i] {
			t.Errorf("After rotate(-3), At(%d) expected %d, got %d", i, expected[i], val)
		}
	}

	// Rotate by length (should be no-op)
	s.Rotate(5)
	for i := 0; i < 5; i++ {
		val, _ := s.At(i)
		if val != expected[i] {
			t.Errorf("After rotate(5), At(%d) expected %d, got %d", i, expected[i], val)
		}
	}
}

func TestTrimToSize(t *testing.T) {
	s := Stack.NewStack[int](64)
	initCap := 64
	// Fill with some data
	for i := 0; i < 32; i++ {
		s.Push(i)
	}

	// Remove some elements
	for i := 0; i < 16; i++ {
		s.Pop()
	}

	capBefore := s.Capacity()
	s.TrimToSize()
	capAfter := s.Capacity()

	if capAfter > capBefore {
		t.Errorf("Expected capacity to shrink, before: %d, after: %d", capBefore, capAfter)
	}
	if s.Length() != 16 {
		t.Errorf("Expected length 16 after trim, got %d", s.Length())
	}

	// Verify elements
	for i := 0; i < 16; i++ {
		val, _ := s.At(i)
		if val != 15-i {
			t.Errorf("At(%d) expected %d, got %d", i, 15-i, val)
		}
	}

	s.Clear()
	s.TrimToSize()
	if s.Capacity() != initCap {
		t.Errorf("Empty stack should trim to initCap (%d), got %d", initCap, s.Capacity())
	}
}

func TestClear(t *testing.T) {
	s := Stack.NewStack[int]()

	// Fill with data
	for i := 0; i < 100; i++ {
		s.Push(i)
	}

	s.Clear()
	if s.Length() != 0 {
		t.Errorf("After clear, length should be 0, got %d", s.Length())
	}
	if !s.Empty() {
		t.Error("After clear, stack should be empty")
	}

	// Should be able to reuse
	for i := 0; i < 10; i++ {
		s.Push(i)
	}
	if s.Length() != 10 {
		t.Errorf("After reuse, expected length 10, got %d", s.Length())
	}
}

func TestCopy(t *testing.T) {
	s := Stack.NewStack[int]()

	// Fill with data
	for i := 0; i < 100; i++ {
		s.Push(i)
	}

	// Make a copy
	copyS := s.Copy()

	// Modify original
	for i := 0; i < 50; i++ {
		s.Pop()
	}
	for i := 100; i < 150; i++ {
		s.Push(i)
	}

	// Verify copy is unchanged
	if copyS.Length() != 100 {
		t.Errorf("Copy length should be 100, got %d", copyS.Length())
	}
	for i := 0; i < 100; i++ { //0-99
		val, _ := copyS.At(i)
		if val != 99-i {
			t.Errorf("Copy At(%d) expected %d, got %d", i, 99-i, val)
		}
	}

	// Verify original is modified as expected
	if s.Length() != 100 {
		t.Errorf("Original length should be 100, got %d", s.Length())
	}
	for i := 0; i < 50; i++ {
		val, _ := s.At(i)
		expected := 149 - i
		if val != expected {
			t.Errorf("Original At(%d) expected %d, got %d", i, expected, val)
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := Stack.NewStack[int]()
	var wg sync.WaitGroup
	count := 1000

	// Concurrent pushers
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < count; i++ {
			s.Push(i)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < count; i++ {
			s.Push(-i)
		}
	}()

	// Concurrent popper
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < count; i++ {
			s.Pop()
		}
	}()

	// Concurrent reader
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < count*2; i++ {
			s.Length()
			runtime.Gosched()
		}
	}()

	wg.Wait()

	// Verify final state is consistent
	if s.Length() < 0 {
		t.Errorf("Invalid length after concurrent access: %d", s.Length())
	}
}

func TestFormat(t *testing.T) {
	s := Stack.NewStack[int]()
	for i := 0; i < 5; i++ {
		s.Push(i)
	}

	// Test default format
	expected := "[4 3 2 1 0]"
	if str := fmt.Sprintf("%v", s); str != expected {
		t.Errorf("Format %%v expected %q, got %q", expected, str)
	}

	// Test width limit
	expected = "[4 3 2 ...+2]"
	if str := fmt.Sprintf("%5v", s); str != expected {
		t.Errorf("Format %%5v expected %q, got %q", expected, str)
	}

	// Test empty stack
	s.Clear()
	expected = "[]"
	if str := fmt.Sprintf("%v", s); str != expected {
		t.Errorf("Format %%v expected %q, got %q", expected, str)
	}
}

func BenchmarkPushPop(b *testing.B) {
	s := Stack.NewStack[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Push(i)
		s.Pop()
	}
}

func BenchmarkConcurrentPushPop(b *testing.B) {
	s := Stack.NewStack[int]()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			if r.Intn(2) == 0 {
				s.Push(1)
			} else {
				s.Pop()
			}
		}
	})
}

func BenchmarkRotate(b *testing.B) {
	s := Stack.NewStack[int]()
	for i := 0; i < 1000; i++ {
		s.Push(i)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Rotate(1)
	}
}
