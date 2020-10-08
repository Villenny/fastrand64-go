package fastrand64

import (
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type UnsafePcg32RNG struct {
	state uint64
	inc   uint64
}

func (r *UnsafePcg32RNG) SetState(initstate uint64, initseq uint64) {
	r.state = 0
	r.inc = (initseq << 1) | 1
	r.Uint32()
	r.state += initstate
	r.Uint32()
}

func (r *UnsafePcg32RNG) Seed(seed int64) {
	r.SetState(
		Splitmix64(uint64(seed)+uint64(0)),
		Splitmix64(uint64(seed)+uint64(1)),
	)
}

func (r *UnsafePcg32RNG) Uint32() uint32 {
	oldstate := r.state
	r.state = oldstate*6364136223846793005 + r.inc
	xorshifted := ((oldstate >> 18) ^ oldstate) >> 27
	rot := oldstate >> 59
	result := (xorshifted >> rot) | (xorshifted << ((-rot) & 31))
	return uint32(result)
}

func NewUnsafePcg32RNG(seed int64) *UnsafePcg32RNG {
	r := &UnsafePcg32RNG{}
	r.Seed(seed)
	return r
}

type UnsafePcg32x2RNG struct {
	gen0 UnsafePcg32RNG
	gen1 UnsafePcg32RNG
}

func (r *UnsafePcg32x2RNG) SetState(seed1 uint64, seq1 uint64, seed2 uint64, seq2 uint64) {
	mask := ^uint64(0) >> 1
	// The stream for each of the two generators *must* be distinct
	if (seq1 & mask) == (seq2 & mask) {
		seq2 = ^seq2
	}
	r.gen0.SetState(seed1, seq1)
	r.gen1.SetState(seed2, seq2)
}

func (r *UnsafePcg32x2RNG) Seed(seed int64) {
	r.SetState(
		Splitmix64(uint64(seed)+uint64(0)),
		Splitmix64(uint64(seed)+uint64(1)),
		Splitmix64(uint64(seed)+uint64(2)),
		Splitmix64(uint64(seed)+uint64(3)),
	)
}

func (r *UnsafePcg32x2RNG) Uint64() uint64 {
	return (uint64(r.gen0.Uint32()) << 32) | uint64(r.gen1.Uint32())
}

func NewUnsafePcg32x2RNG(seed int64) *UnsafePcg32x2RNG {
	r := &UnsafePcg32x2RNG{}
	r.Seed(seed)
	return r
}

type UnsafeJsf64RNG struct {
	a uint64
	b uint64
	c uint64
	d uint64
}

func rot64(x uint64, k uint64) uint64 {
	return ((x << k) | (x >> (64 - k)))
}

func (x *UnsafeJsf64RNG) Uint64() uint64 {
	e := x.a - rot64(x.b, 7)
	x.a = x.b ^ rot64(x.c, 13)
	x.b = x.c + rot64(x.d, 37)
	x.c = x.d + e
	x.d = e + x.a
	return x.d
}

func (x *UnsafeJsf64RNG) Seed(seed int64) {
	x.a = 0xf1ea5eed
	x.b = uint64(seed)
	x.c = uint64(seed)
	x.d = uint64(seed)
	for i := 0; i < 20; i++ {
		x.Uint64()
	}
}

func NewUnsafeJsf64RNG(seed int64) *UnsafeJsf64RNG {
	r := &UnsafeJsf64RNG{}
	r.Seed(seed)
	return r
}

func Test_UnsafeCast(t *testing.T) {
	var i uint64 = 1
	b := make([]byte, 16)

	header := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	ptr := (*uint64)(unsafe.Pointer(header.Data))

	ptr = (*uint64)(unsafe.Pointer((uintptr)(unsafe.Pointer(ptr)) + 8))
	*ptr = i

	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0}, b)
}

// ////////////////////////////////////////////////////////////////

/*

Running tool: C:\Go\bin\go.exe test -benchmem -run=^$ github.com/villenny/concurrency-go -bench ^(Benchmark_UnsafePcg32x2RNG|Benchmark_UnsafeJsf64RNG)$

goos: windows
goarch: amd64
pkg: github.com/villenny/concurrency-go
Benchmark_UnsafePcg32x2RNG-8   	223279117	         5.28 ns/op	       0 B/op	       0 allocs/op
Benchmark_UnsafeJsf64RNG-8     	295147552	         4.10 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/villenny/concurrency-go	3.550s

*/
func Benchmark_UnsafePcg32x2RNG(b *testing.B) {
	rng := NewUnsafePcg32x2RNG(time.Now().UnixNano())
	var r uint64
	for i := 0; i < b.N; i++ {
		r = rng.Uint64()
	}
	BenchSink = &r
}

func Benchmark_UnsafeJsf64RNG(b *testing.B) {
	rng := NewUnsafeJsf64RNG(time.Now().UnixNano())
	var r uint64
	for i := 0; i < b.N; i++ {
		r = rng.Uint64()
	}
	BenchSink = &r
}
