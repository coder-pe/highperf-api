package jsonx

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

type TestStruct struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestNewDecoder(t *testing.T) {
	jsonData := `{"id": 123, "name": "test"}`
	reader := strings.NewReader(jsonData)

	decoder := NewDecoder(reader)
	if decoder == nil {
		t.Fatal("NewDecoder returned nil")
	}

	var ts TestStruct
	err := decoder.Decode(&ts)
	if err != nil {
		t.Errorf("Decode failed: %v", err)
	}

	if ts.ID != 123 {
		t.Errorf("Expected ID 123, got %d", ts.ID)
	}

	if ts.Name != "test" {
		t.Errorf("Expected name 'test', got %q", ts.Name)
	}
}

func TestNewDecoderDisallowUnknownFields(t *testing.T) {
	jsonData := `{"id": 123, "name": "test", "unknown": "field"}`
	reader := strings.NewReader(jsonData)

	decoder := NewDecoder(reader)
	var ts TestStruct
	err := decoder.Decode(&ts)

	// Should fail due to unknown field
	if err == nil {
		t.Error("Expected error due to unknown field, got nil")
	}
}

func TestMarshalToBuffer(t *testing.T) {
	ts := TestStruct{ID: 123, Name: "test"}
	buf := &bytes.Buffer{}

	err := MarshalToBuffer(ts, buf)
	if err != nil {
		t.Errorf("MarshalToBuffer failed: %v", err)
	}

	// Parse the result back to verify
	var result TestStruct
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Errorf("Failed to unmarshal result: %v", err)
	}

	if result.ID != ts.ID {
		t.Errorf("Expected ID %d, got %d", ts.ID, result.ID)
	}

	if result.Name != ts.Name {
		t.Errorf("Expected name %q, got %q", ts.Name, result.Name)
	}
}

func TestMarshalToBufferEscapeHTML(t *testing.T) {
	ts := TestStruct{ID: 1, Name: "<script>alert('xss')</script>"}
	buf := &bytes.Buffer{}

	err := MarshalToBuffer(ts, buf)
	if err != nil {
		t.Errorf("MarshalToBuffer failed: %v", err)
	}

	result := buf.String()
	// Should NOT escape HTML since SetEscapeHTML(false) is called
	if !strings.Contains(result, "<script>") {
		t.Errorf("Expected HTML not to be escaped, got %q", result)
	}
}

func TestMarshalToBufferWithInvalidData(t *testing.T) {
	// Create something that can't be marshaled
	invalidData := make(chan int)
	buf := &bytes.Buffer{}

	err := MarshalToBuffer(invalidData, buf)
	if err == nil {
		t.Error("Expected error when marshaling invalid data, got nil")
	}
}

func TestMarshalToBufferReusesBuffer(t *testing.T) {
	ts1 := TestStruct{ID: 1, Name: "first"}
	ts2 := TestStruct{ID: 2, Name: "second"}
	buf := &bytes.Buffer{}

	// First marshal
	err := MarshalToBuffer(ts1, buf)
	if err != nil {
		t.Errorf("First marshal failed: %v", err)
	}

	firstLen := buf.Len()
	if firstLen == 0 {
		t.Error("Buffer should not be empty after first marshal")
	}

	// Second marshal should append to the buffer
	err = MarshalToBuffer(ts2, buf)
	if err != nil {
		t.Errorf("Second marshal failed: %v", err)
	}

	if buf.Len() <= firstLen {
		t.Error("Buffer should be longer after second marshal")
	}

	// Buffer should contain both JSON objects
	content := buf.String()
	if !strings.Contains(content, "first") || !strings.Contains(content, "second") {
		t.Errorf("Buffer should contain both objects, got %q", content)
	}
}

func BenchmarkNewDecoder(b *testing.B) {
	jsonData := `{"id": 123, "name": "test"}`
	reader := strings.NewReader(jsonData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader.Reset(jsonData)
		decoder := NewDecoder(reader)
		var ts TestStruct
		decoder.Decode(&ts)
	}
}

func BenchmarkMarshalToBuffer(b *testing.B) {
	ts := TestStruct{ID: 123, Name: "test"}
	buf := &bytes.Buffer{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		MarshalToBuffer(ts, buf)
	}
}

func BenchmarkMarshalToBufferVsStdlib(b *testing.B) {
	ts := TestStruct{ID: 123, Name: "test"}

	b.Run("MarshalToBuffer", func(b *testing.B) {
		buf := &bytes.Buffer{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			MarshalToBuffer(ts, buf)
		}
	})

	b.Run("json.Marshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(ts)
		}
	})
}
