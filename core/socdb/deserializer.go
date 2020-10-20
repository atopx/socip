package socdb

import (
	"math/big"

	"github.com/pkg/errors"
)

// Potentially, it would make sense to add this to mmdbtypes and make
// it public, but we should wait until the API stabilized here and in
// maxminddb first.

func newDeserializer() *deserializer {
	return &deserializer{
		cache:      map[uintptr]DataType{},
		lastOffset: noOffset,
	}
}

func (d *deserializer) ShouldSkip(offset uintptr) (bool, error) {
	v, ok := d.cache[offset]
	if ok {
		d.lastOffset = noOffset
		return true, d.simpleAdd(v)
	}
	d.lastOffset = offset
	return false, nil
}

func (d *deserializer) StartSlice(size uint) error {
	// We make the slice its finalize size to avoid
	// appending, which could interfere with the caching.
	return d.add(make(Slice, size))
}

func (d *deserializer) StartMap(size uint) error {
	return d.add(make(Map, size))
}

func (d *deserializer) End() error {
	if len(d.stack) == 0 {
		return errors.New("received an End but the stack in empty")
	}
	d.stack = d.stack[:len(d.stack)-1]
	return nil
}

func (d *deserializer) String(v string) error {
	return d.add(String(v))
}

func (d *deserializer) Float64(v float64) error {
	return d.add(Float64(v))
}

func (d *deserializer) Bytes(v []byte) error {
	return d.add(Bytes(v))
}

func (d *deserializer) Uint16(v uint16) error {
	return d.add(Uint16(v))
}

func (d *deserializer) Uint32(v uint32) error {
	return d.add(Uint32(v))
}

func (d *deserializer) Int32(v int32) error {
	return d.add(Int32(v))
}

func (d *deserializer) Uint64(v uint64) error {
	return d.add(Uint64(v))
}

func (d *deserializer) Uint128(v *big.Int) error {
	t := Uint128(*v)
	return d.add(&t)
}

func (d *deserializer) Bool(v bool) error {
	return d.add(Bool(v))
}

func (d *deserializer) Float32(v float32) error {
	return d.add(Float32(v))
}

func (d *deserializer) simpleAdd(v DataType) error {
	if len(d.stack) == 0 {
		d.rv = v
	} else {
		top := d.stack[len(d.stack)-1]
		switch parent := top.value.(type) {
		case Map:
			if d.key == nil {
				key, ok := v.(String)
				if !ok {
					return errors.Errorf("expected a String Map key but received %T", v)
				}
				d.key = &key
			} else {
				parent[*d.key] = v
				d.key = nil
				top.curSize++
			}

		case Slice:
			parent[top.curSize] = v
			top.curSize++
		default:
		}
	}
	return nil
}

func (d *deserializer) add(v DataType) error {
	err := d.simpleAdd(v)
	if err != nil {
		return err
	}

	switch v := v.(type) {
	case Map, Slice:
		d.stack = append(d.stack, &stackValue{value: v})
	default:
	}

	if d.lastOffset != noOffset {
		d.cache[d.lastOffset] = v
		d.lastOffset = noOffset
	}
	return nil
}

func (d *deserializer) clear() {
	d.rv = nil

	// Although these shouldn't be necessary normally, they could be needed
	// if we are recovering from an error.
	d.key = nil
	if len(d.stack) > 0 {
		d.stack = d.stack[:0]
	}
}
