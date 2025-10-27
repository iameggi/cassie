package bucket

import (
	"bytes"
	"strings"
)

// DefaultCapacity defines the default initial capacity for pooled buffers (4KB).
const DefaultCapacity = 4096

// --- Factory Functions ---

// NewBytePool creates a new *Pool[bytes.Buffer] with the given initial capacity.
// The buffer will be automatically reset when returned to the pool.
func NewBytePool(initialCapacity int) *Pool[bytes.Buffer] {
	if initialCapacity <= 0 {
		initialCapacity = DefaultCapacity
	}
	return New(
		func() *bytes.Buffer {
			return bytes.NewBuffer(make([]byte, 0, initialCapacity))
		},
		func(b *bytes.Buffer) {
			b.Reset()
		},
	)
}

// NewStringBuilderPool creates a new *Pool[strings.Builder] with the given initial capacity.
// The builder will be automatically reset when returned to the pool.
func NewStringBuilderPool(initialCapacity int) *Pool[strings.Builder] {
	if initialCapacity <= 0 {
		initialCapacity = DefaultCapacity
	}
	return New(
		func() *strings.Builder {
			var b strings.Builder
			b.Grow(initialCapacity)
			return &b
		},
		func(b *strings.Builder) {
			b.Reset()
		},
	)
}

// --- Global Pools ---

// ByteBucket provides a ready-to-use global pool of *bytes.Buffer
// with a default capacity of 4KB.
var ByteBucket = NewBytePool(DefaultCapacity)

// StringBuilderBucket provides a ready-to-use global pool of *strings.Builder
// with a default capacity of 4KB.
var StringBuilderBucket = NewStringBuilderPool(DefaultCapacity)

// --- Safe Callback Helpers ---

// WithByteBuffer executes the given function f with a pooled *bytes.Buffer
// from ByteBucket. The buffer is automatically returned to the pool after use.
func WithByteBuffer(f func(buf *bytes.Buffer)) {
	ByteBucket.With(f)
}

// WithByteBufferErr executes the given function f with a pooled *bytes.Buffer
// from ByteBucket. The buffer is automatically returned to the pool after use.
// Any error returned by f is propagated to the caller.
func WithByteBufferErr(f func(buf *bytes.Buffer) error) error {
	return ByteBucket.WithErr(f)
}

// WithStringBuilder executes the given function f with a pooled *strings.Builder
// from StringBuilderBucket. The builder is automatically returned to the pool after use.
func WithStringBuilder(f func(sb *strings.Builder)) {
	StringBuilderBucket.With(f)
}

// WithStringBuilderErr executes the given function f with a pooled *strings.Builder
// from StringBuilderBucket. The builder is automatically returned to the pool after use.
// Any error returned by f is propagated to the caller.
func WithStringBuilderErr(f func(sb *strings.Builder) error) error {
	return StringBuilderBucket.WithErr(f)
}
