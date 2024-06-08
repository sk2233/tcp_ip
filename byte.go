/*
@author: sk
@date: 2024/6/5
*/
package main

import "encoding/binary"

type ByteReader struct {
	data  []byte
	index int
}

func (r *ByteReader) ReadByte(count int) []byte {
	r.index += count
	return r.data[r.index-count : r.index]
}

func (r *ByteReader) ReadU16() uint16 {
	bs := r.ReadByte(2)
	return binary.BigEndian.Uint16(bs)
}

func (r *ByteReader) ReadU32() uint32 {
	bs := r.ReadByte(4)
	return binary.BigEndian.Uint32(bs)
}

func (r *ByteReader) ReadU8() uint8 {
	bs := r.ReadByte(1)
	return bs[0]
}

func (r *ByteReader) ReadLast() []byte {
	return r.data[r.index:]
}

func (r *ByteReader) Len() uint16 {
	return uint16(len(r.data) - r.index)
}

func NewByteReader(data []byte) *ByteReader {
	return &ByteReader{data: data, index: 0}
}

type ByteWriter struct {
	data  []byte
	index int
}

func (w *ByteWriter) WriteByte(data []byte) {
	w.index -= len(data)
	copy(w.data[w.index:], data)
}

func (w *ByteWriter) WriteU16(data uint16) {
	w.index -= 2
	binary.BigEndian.PutUint16(w.data[w.index:], data)
}

func (w *ByteWriter) WriteU8(data uint8) {
	w.index--
	w.data[w.index] = data
}

func (w *ByteWriter) Len() uint16 {
	return uint16(len(w.data) - w.index)
}

func (w *ByteWriter) GetData() []byte {
	return w.data[w.index:]
}

func (w *ByteWriter) Seek(offset int) {
	w.index += offset
}

func (w *ByteWriter) WriteU32(data uint32) {
	w.index -= 4
	binary.BigEndian.PutUint32(w.data[w.index:], data)
}

func NewByteWriter() *ByteWriter {
	return &ByteWriter{data: make([]byte, MaxPackageSize), index: MaxPackageSize}
}
