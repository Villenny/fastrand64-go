// Package fastrand64 implements fast pesudorandom number generator
// that should scale well on multi-CPU systems.
//
// Use crypto/rand instead of this package for generating
// cryptographically secure random numbers.
//
// Example:
//
//  import "github.com/villenny/fastrand64-go"
//
//  make a threadsafe random generator
//  rng := NewSyncPoolXoshiro256ssRNG()
//
//  // somewhere later, in some goproc, one of lots, like a web request handler for example
//  // this (ab)uses a sync.Pool to allocate one generator per thread
//  r1 := rng.Uint32n(10)
//  r2 := rng.Uint64()
//  someBytes := rng.Bytes(8)
//
//  // This will produce R1=<random int 0-9>, R2=<random unsigned 64bit int>, someBytes=<random bytes>
//  fmt.Printf("R1=%v, R2=%v, someBytes=%v", r1, r2, someBytes)
package fastrand64

import (
	"math/rand"
	"sync"
	"time"
)

// ThreadsafePoolRNG core type for the pool backed threadsafe RNG
type ThreadsafePoolRNG struct {
	rngPool sync.Pool
}

// UnsafeRNG is the interface for an unsafe RNG used by the Pool RNG as a source of randomness
type UnsafeRNG interface {
	Uint64() uint64
}

// NewSyncPoolRNG Wraps a sync.Pool around a thread unsafe RNG, thus making it efficiently thread safe
func NewSyncPoolRNG(fn func() UnsafeRNG) *ThreadsafePoolRNG {
	s := &ThreadsafePoolRNG{}
	s.rngPool = sync.Pool{New: func() interface{} { return fn() }}
	return s
}

// NewSyncPoolXoshiro256ssRNG conveniently allocations a thread safe pooled back xoshiro256** generator
// this uses NewSyncPoolRNG internally
func NewSyncPoolXoshiro256ssRNG() *ThreadsafePoolRNG {
	rand.Seed(time.Now().UnixNano())
	return NewSyncPoolRNG(func() UnsafeRNG {
		return NewUnsafeXoshiro256ssRNG(int64(rand.Uint64()))
	})
}

// Uint64 returns pseudorandom uint64. Threadsafe
func (s *ThreadsafePoolRNG) Uint64() uint64 {
	r := s.rngPool.Get().(UnsafeRNG)
	x := r.Uint64()
	s.rngPool.Put(r)
	return x
}

// Int63 is here to match Source64 interface, why not call Int64
func (s *ThreadsafePoolRNG) Int63() int64 {
	return int64(0x7FFFFFFFFFFFFFFF & s.Uint64())
}

// Seed is only here to match the golang std libs Source64 interface
func (s *ThreadsafePoolRNG) Seed(seed int64) {
	// you cant really seed a PoolRNG, since the call order is non-determinate
	panic("Cant seed a ThreadsafePoolRNG")
}

// Bytes allocates a []byte filled with random bytes and returns it. This is convenient
// but caller does the allocation pattern is better way since it can reduce allocation count/GC
func (s *ThreadsafePoolRNG) Bytes(n int) []byte {
	r := s.rngPool.Get().(UnsafeRNG)
	bytes := make([]byte, n)
	result := Bytes(r, bytes)
	s.rngPool.Put(r)
	return result
}

// Read fills a []byte array with random bytes from a thread safe pool backed RNG
func (s *ThreadsafePoolRNG) Read(p []byte) []byte {
	r := s.rngPool.Get().(UnsafeRNG)
	Bytes(r, p)
	s.rngPool.Put(r)
	return p
}

// Bytes fills a []byte array with random bytes from a thread unsafe RNG
func Bytes(r UnsafeRNG, bytes []byte) []byte {
	n := len(bytes)

	/*
		header := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
		ptr := (*uint64)(unsafe.Pointer(header.Data))
		offsetMax := uintptr(n - (n % 8))
		ptrMax := (*uint64)(unsafe.Pointer(header.Data + offsetMax))

		for {
			if ptr == ptrMax {
				break
			}
			x := r.Uint64()
			*ptr = x
			ptr = (*uint64)(unsafe.Pointer((uintptr)(unsafe.Pointer(ptr)) + 8))
		}

		i := int(offsetMax)
	*/

	i := 0
	iMax := n - (n % 8)
	for {
		if i == iMax {
			break
		}
		x := r.Uint64()

		bytes[i] = byte(x)
		bytes[i+1] = byte(x >> 8)
		bytes[i+2] = byte(x >> 16)
		bytes[i+3] = byte(x >> 24)
		bytes[i+4] = byte(x >> 32)
		bytes[i+5] = byte(x >> 40)
		bytes[i+6] = byte(x >> 48)
		bytes[i+7] = byte(x >> 56)
		i += 8
	}

	x := r.Uint64()
	for {
		if i >= n {
			break
		}
		bytes[i] = byte(x)
		x >>= 8
		i += 1
	}

	return bytes
}

// Uint32n returns pseudorandom Uint32n in the range [0..maxN).
//
// It is safe calling this function from concurrent goroutines.
func (s *ThreadsafePoolRNG) Uint32n(maxN int) uint32 {
	x := s.Uint64() & 0x00000000FFFFFFFF
	// See http://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
	return uint32((x * uint64(maxN)) >> 32)
}

// UnsafeXoshiro256ssRNG It is unsafe to call UnsafeRNG methods from concurrent goroutines.
//
// UnsafeXoshiro256** is a pseudorandom number generator.
// For an interesting commentary on xoshiro256**
// https://www.pcg-random.org/posts/a-quick-look-at-xoshiro256.html
//
// It is however very fast, and strong enough for most practical purposes
type UnsafeXoshiro256ssRNG struct {
	s0 uint64
	s1 uint64
	s2 uint64
	s3 uint64
}

func rol64(x uint64, k uint64) uint64 {
	return (x << k) | (x >> (64 - k))
}

// Splitmix64 is typically used to convert a potentially zero seed, into better non-zero seeds
// ie seeding a stronger RNG
func Splitmix64(index uint64) uint64 {
	z := (index + uint64(0x9E3779B97F4A7C15))
	z = (z ^ (z >> 30)) * uint64(0xBF58476D1CE4E5B9)
	z = (z ^ (z >> 27)) * uint64(0x94D049BB133111EB)
	z = z ^ (z >> 31)
	return z
}

// Uint64 generates a random Uin64, (not thread safe)
func (r *UnsafeXoshiro256ssRNG) Uint64() uint64 {
	// See https://en.wikipedia.org/wiki/Xorshift
	result := rol64(r.s1*5, 7) * 9
	t := r.s1 << 17

	r.s2 ^= r.s0
	r.s3 ^= r.s1
	r.s1 ^= r.s2
	r.s0 ^= r.s3

	r.s2 ^= t
	r.s3 = rol64(r.s3, 45)

	return result
}

// Seed takes a single uint64 and runs it through splitmix64 to seed the 256 bit starting state for the RNG
func (r *UnsafeXoshiro256ssRNG) Seed(seed int64) {
	i := 0
	for r.s0 = 0; r.s0 == 0; i++ {
		r.s0 = Splitmix64(uint64(seed) + uint64(i))
	}
	for r.s1 = 0; r.s1 == 0; i++ {
		r.s1 = Splitmix64(uint64(seed) + uint64(i))
	}
	for r.s2 = 0; r.s2 == 0; i++ {
		r.s2 = Splitmix64(uint64(seed) + uint64(i))
	}
	for r.s3 = 0; r.s3 == 0; i++ {
		r.s3 = Splitmix64(uint64(seed) + uint64(i))
	}
}

// NewUnsafeXoshiro256ssRNG creates a new Thread unsafe PRNG generator
func NewUnsafeXoshiro256ssRNG(seed int64) *UnsafeXoshiro256ssRNG {
	r := &UnsafeXoshiro256ssRNG{}
	r.Seed(seed)
	return r
}

// NewUnsafeRandRNG creates a new Thread unsafe PRNG generator using the native golang 64bit RNG generator
// (thus avoiding using any global state)
func NewUnsafeRandRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed).(rand.Source64))
}
