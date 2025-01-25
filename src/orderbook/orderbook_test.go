package main

import (
	"container/heap"
	"testing"
)

func TestMinHeapPop(t *testing.T) {
	h := &MinHeap{2.0, 1.0, 5.0, 0.0, 6.0, 7.0}
	wants := []float64{0.0, 1.0, 2.0, 5.0, 6.0, 7.0}
	heap.Init(h)

	// Test pop
	count := 0
	for h.Len() > 0 {
		got := heap.Pop(h)
		want := wants[count]
		if got != want {
			t.Errorf("got %f, wanted %f for index %d", got, want, count)
		}
		count++
	}
}

func TestMinHeapPeak(t *testing.T) {
	h := &MinHeap{2.0, 1.0, 5.0, 0.0, 6.0, 7.0}
	wants := []float64{0.0, 1.0, 2.0, 5.0, 6.0, 7.0}
	heap.Init(h)

	// Test pop
	count := 0
	for h.Len() > 0 {
		got := h.Peak().(float64)
		want := wants[count]
		if got != want {
			t.Errorf("got %f, wanted %f for index %d", got, want, count)
		}
		count++
		heap.Pop(h)
	}
}

func TestMaxHeapPop(t *testing.T) {
	h := &MaxHeap{2.0, 1.0, 5.0, 0.0, 6.0, 7.0}
	wants := []float64{7.0, 6.0, 5.0, 2.0, 1.0, 0.0}
	heap.Init(h)

	// Test pop
	count := 0
	for h.Len() > 0 {
		got := heap.Pop(h)
		want := wants[count]
		if got != want {
			t.Errorf("got %f, wanted %f for index %d", got, want, count)
		}
		count++
	}
}

func TestMaxHeapPeak(t *testing.T) {
	h := &MaxHeap{2.0, 1.0, 5.0, 0.0, 6.0, 7.0}
	wants := []float64{7.0, 6.0, 5.0, 2.0, 1.0, 0.0}
	heap.Init(h)

	// Test pop
	count := 0
	for h.Len() > 0 {
		got := h.Peak().(float64)
		want := wants[count]
		if got != want {
			t.Errorf("got %f, wanted %f for index %d", got, want, count)
		}
		count++
		heap.Pop(h)
	}
}
