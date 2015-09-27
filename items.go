package kv

import (
	"encoding/gob"
	"encoding/json"
	"io"
	"io/ioutil"
)

// Simple raw item encoding - copies value bytes as is
type ByteItem struct {
	Value []byte
}

func (b *ByteItem) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, string(b.Value))
	return int64(n), err
}

func (b *ByteItem) ReadFrom(r io.Reader) (int64, error) {
	data, err := ioutil.ReadAll(r)
	if err == nil {
		b.Value = data
	}
	return int64(len(data)), err
}

// JSON encoding - serializes arbitrary value to/from JSON
type JSONItem struct {
	Value interface{}
}

func (e *JSONItem) WriteTo(w io.Writer) (int64, error) {
	if err := json.NewEncoder(w).Encode(e.Value); err != nil {
		return 0, err
	}
	return 0, nil
}

func (e *JSONItem) ReadFrom(r io.Reader) (int64, error) {
	if err := json.NewDecoder(r).Decode(e.Value); err != nil {
		return 0, err
	}
	return 0, nil
}

// Gob encoding - serializes arbitrary value to/from gob format
type GobItem struct {
	Value interface{}
}

func (e *GobItem) WriteTo(w io.Writer) (int64, error) {
	if err := gob.NewEncoder(w).Encode(e.Value); err != nil {
		return 0, err
	}
	return 0, nil
}

func (e *GobItem) ReadFrom(r io.Reader) (int64, error) {
	if err := gob.NewDecoder(r).Decode(e.Value); err != nil {
		return 0, err
	}
	return 0, nil
}
