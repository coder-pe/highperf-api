package jsonx

import (
	"bytes"
	"sync"
	"testing"
)

func TestGetBuffer(t *testing.T) {
	buf := GetBuffer()
	if buf == nil {
		t.Fatal("GetBuffer returned nil")
	}

	// Buffer should be empty/reset
	if buf.Len() != 0 {
		t.Errorf("Expected empty buffer, got length %d", buf.Len())
	}

	// Should be able to write to it
	buf.WriteString("test")
	if buf.Len() != 4 {
		t.Errorf("Expected buffer length 4 after write, got %d", buf.Len())
	}
}

func TestPutBuffer(t *testing.T) {
	buf := GetBuffer()
	buf.WriteString("test data")

	// Put buffer back in pool
	PutBuffer(buf)

	// Get a new buffer - should be reset
	buf2 := GetBuffer()
	if buf2.Len() != 0 {
		t.Errorf("Expected reset buffer, got length %d", buf2.Len())
	}

	// Clean up
	PutBuffer(buf2)
}

func TestPutBufferLargeCapacity(t *testing.T) {
	buf := GetBuffer()

	// Write a lot of data to increase capacity
	data := make([]byte, 2<<20) // 2MB
	buf.Write(data)

	if buf.Cap() <= 1<<20 {
		t.Skip("Buffer capacity not large enough for this test")
	}

	// PutBuffer should not return large buffers to pool
	PutBuffer(buf)

	// Get a new buffer - should be a fresh one, not the large one
	buf2 := GetBuffer()
	if buf2.Cap() >= 1<<20 {
		t.Errorf("Expected fresh buffer with smaller capacity, got %d", buf2.Cap())
	}

	PutBuffer(buf2)
}

func TestBufferPoolConcurrency(t *testing.T) {
	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup

	// Test concurrent get/put operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				buf := GetBuffer()
				buf.WriteString("concurrent test")
				PutBuffer(buf)
			}
		}()
	}

	wg.Wait()
}

func TestBufferReuse(t *testing.T) {
	buf1 := GetBuffer()
	buf1.WriteString("first use")

	// Store reference to verify reuse
	ptr1 := &buf1.Bytes()[0]

	PutBuffer(buf1)

	buf2 := GetBuffer()
	buf2.WriteString("second use")

	// Note: We can't guarantee the same buffer is returned due to pool implementation,
	// but we can test that it works correctly
	if buf2.Len() != len("second use") {
		t.Errorf("Expected buffer length %d, got %d", len("second use"), buf2.Len())
	}

	PutBuffer(buf2)

	// Test that buffer content is properly reset
	buf3 := GetBuffer()
	if buf3.Len() != 0 {
		t.Errorf("Expected reset buffer, got length %d with content: %q", buf3.Len(), buf3.String())
	}

	PutBuffer(buf3)
	_ = ptr1 // Use ptr1 to avoid unused variable error
}

func TestBufferGrowth(t *testing.T) {
	buf := GetBuffer()
	initialCap := buf.Cap()

	// Write data to force growth
	data := make([]byte, initialCap+100)
	for i := range data {
		data[i] = byte(i % 256)
	}
	buf.Write(data)

	if buf.Cap() <= initialCap {
		t.Errorf("Expected buffer to grow beyond %d, got %d", initialCap, buf.Cap())
	}

	if buf.Len() != len(data) {
		t.Errorf("Expected buffer length %d, got %d", len(data), buf.Len())
	}

	PutBuffer(buf)
}

func BenchmarkGetPutBuffer(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := GetBuffer()
		buf.WriteString("benchmark test data")
		PutBuffer(buf)
	}
}

func BenchmarkBufferWithoutPool(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		buf.WriteString("benchmark test data")
		// No pooling - let GC handle it
	}
}

func BenchmarkConcurrentBufferPool(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := GetBuffer()
			buf.WriteString("concurrent benchmark")
			PutBuffer(buf)
		}
	})
}

func TestPoolCapacityLimit(t *testing.T) {
	// Create a buffer larger than the limit
	buf := GetBuffer()

	// Force the buffer to grow beyond 1MB
	largeData := make([]byte, 2<<20) // 2MB
	buf.Write(largeData)

	if buf.Cap() <= 1<<20 {
		t.Skip("Buffer didn't grow large enough for this test")
	}

	// This should not put the buffer back in the pool
	PutBuffer(buf)

	// Verify that a new buffer is created (not the large one)
	newBuf := GetBuffer()
	if newBuf.Cap() >= 1<<20 {
		t.Errorf("Large buffer was returned from pool when it should have been discarded")
	}

	PutBuffer(newBuf)
}
