package fastxml

import (
	"github.com/orisano/gosax"
	"sync"
)

var attrPool sync.Pool

type Attribute struct {
	Key   []byte
	Value []byte
}

// AcquireAttribute returns new Attribute from the pool.
func AcquireAttribute() *Attribute {
	v := attrPool.Get()
	if v == nil {
		return &Attribute{}
	}
	return v.(*Attribute)
}

// ReleaseAttribute returns the given Attribute to the pool.
//
// The Attribute shouldn't be used after returning to the pool.
func ReleaseAttribute(e *Attribute) {
	e.Reset()
	if !e.isOversized() {
		attrPool.Put(e)
	}
}

var zeroAttr Attribute

// Reset resets the attr for subsequent re-use.
func (e *Attribute) Reset() {
	e.CopyFrom(&zeroAttr)
}

func (e *Attribute) CopyFrom(src *Attribute) {
	e.Key = append(e.Key[:0], src.Key...)
	e.Value = append(e.Value[:0], src.Value...)
}

func (e *Attribute) copyFromSax(src *gosax.Attribute) {
	e.Key = append(e.Key[:0], src.Key...)
	e.Value = append(e.Value[:0], src.Value...)
}

const (
	maxKeyLen   = 32
	maxValueLen = 128
)

func (e *Attribute) isOversized() bool {
	if cap(e.Key) > maxKeyLen {
		return true
	}
	if cap(e.Value) > maxValueLen {
		return true
	}

	return false
}

type Attributes []*Attribute
