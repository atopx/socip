package socdb

import (
	"bytes"
)

func newKeyWriter() *keyWriter {
	return &keyWriter{Buffer: &bytes.Buffer{}}
}

// This is just a quick hack. I am sure there is
// something better
func (kw *keyWriter) key(t DataType) ([]byte, error) {
	kw.Truncate(0)
	_, err := t.WriteTo(kw)
	if err != nil {
		return nil, err
	}
	return kw.Bytes(), nil
}

func (kw *keyWriter) WriteOrWritePointer(t DataType) (int64, error) {
	return t.WriteTo(kw)
}

func newDataWriter(dataMap *dataMap) *dataWriter {
	return &dataWriter{
		Buffer:    &bytes.Buffer{},
		dataMap:   dataMap,
		pointers:  map[string]writtenType{},
		keyWriter: newKeyWriter(),
	}
}

func (dw *dataWriter) maybeWrite(key dataMapKey) (int, error) {
	written, ok := dw.pointers[string(key)]
	if ok {
		return int(written.pointer), nil
	}

	offset := dw.Len()
	size, err := dw.dataMap.get(key).WriteTo(dw)
	if err != nil {
		return 0, err
	}

	written = writtenType{
		pointer: Pointer(offset),
		size:    size,
	}

	dw.pointers[string(key)] = written

	return int(written.pointer), nil
}

func (dw *dataWriter) WriteOrWritePointer(t DataType) (int64, error) {
	key, err := dw.keyWriter.key(t)
	if err != nil {
		return 0, err
	}

	written, ok := dw.pointers[string(key)]
	if ok && written.size > written.pointer.WrittenSize() {
		// Only use a pointer if it would take less space than writing the
		// type again.
		return written.pointer.WriteTo(dw)
	}
	// We can't use the pointers[string(key)] optimization below
	// as the backing buffer for key may change when we call
	// t.WriteTo. That said, this is the less common code path
	// so it doesn't matter too much.
	keyStr := string(key)

	// TODO: A possible optimization here for simple types would be to just
	// write key to the dataWriter. This won't necessarily work for Map and
	// Slice though as they may have internal pointers missing from key.
	// I briefly tested this and didn't see much difference, but it might
	// be worth exploring more.
	offset := dw.Len()
	size, err := t.WriteTo(dw)
	if err != nil || ok {
		return size, err
	}

	dw.pointers[keyStr] = writtenType{
		pointer: Pointer(offset),
		size:    size,
	}
	return size, nil
}
