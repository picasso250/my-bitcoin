package gocoin

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

const HeaderLen = 4

// Encode wraps msg with 4-byte BE length and JSON-serializes it
func Encode(w io.Writer, msg interface{}) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	lenBuf := make([]byte, HeaderLen)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(body)))
	if _, err := w.Write(lenBuf); err != nil {
		return err
	}
	_, err = w.Write(body)
	return err
}

// Decode reads 4-byte length then JSON-decodes into v
func Decode(r io.Reader, v interface{}) error {
	lenBuf := make([]byte, HeaderLen)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return err
	}
	length := binary.BigEndian.Uint32(lenBuf)
	if length > 32*1024*1024 { // max 32 MB for sanity
		return fmt.Errorf("message too large: %d", length)
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}