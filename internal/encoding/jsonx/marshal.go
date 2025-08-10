// internal/encoding/jsonx/marshal.go
package jsonx

import (
	"bytes"
	"encoding/json"
	"io"
)

type pooledEncoder struct {
	*json.Encoder
}

func NewDecoder(r io.Reader) *json.Decoder {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	return dec
}

func MarshalToBuffer(v any, buf *bytes.Buffer) error {
	// json.Encoder reusa el buffer y evita allocs intermedias
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
