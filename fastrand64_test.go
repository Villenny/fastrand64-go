package fastrand64

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tsuna/endian"
	"github.com/valyala/fastrand"
	"github.com/yalue/native_endian"
)

func Test_SafeRNG_Bytes(t *testing.T) {
	rng := NewSyncPoolXoshiro256ssRNG()
	bytes := rng.Bytes(255)
	assert.Equal(t, 255, len(bytes))
}

func Test_SafeRNG_UInt32n(t *testing.T) {
	rng := NewSyncPoolXoshiro256ssRNG()
	for i := 0; i < 4096; i++ {
		r := rng.Uint32n(10)
		assert.Less(t, r, uint32(10))
	}
}

func Test_SafeRNG_UInt64(t *testing.T) {
	rng1 := NewSyncPoolRNG(func() UnsafeRNG { return NewUnsafeRandRNG(1) })
	rng2 := NewUnsafeRandRNG(1)
	for i := 0; i < 256; i++ {
		r1 := rng1.Uint64()
		r2 := rng2.Uint64()
		assert.Equal(t, r1, r2)
	}
}

func Test_SafeRNG_Seed(t *testing.T) {
	rng := NewSyncPoolXoshiro256ssRNG()
	assert.Panics(t, func() { rng.Seed(0) })
}

func Test_SafeRNG_Int63(t *testing.T) {
	rng1 := NewSyncPoolRNG(func() UnsafeRNG { return NewUnsafeRandRNG(1) })
	rng2 := NewUnsafeRandRNG(1)
	for i := 0; i < 256; i++ {
		r1 := rng1.Int63()
		r2 := int64(0x7FFFFFFFFFFFFFFF & rng2.Uint64())
		assert.Equal(t, r1, r2)
	}
}

func Test_SafeRNG_Read(t *testing.T) {
	rng1 := NewSyncPoolRNG(func() UnsafeRNG { return NewUnsafeRandRNG(1) })
	rng2 := NewUnsafeRandRNG(1)

	b := make([]byte, 8)
	rng1.Read(b)
	var r1 uint64
	buf := bytes.NewBuffer(b)
	_ = binary.Read(buf, native_endian.NativeEndian(), &r1)

	r2 := rng2.Uint64()

	assert.Equal(t, r1, r2)
}

func Test_UnsafeXoshiro256ssRNG_UInt64(t *testing.T) {
	rng := UnsafeXoshiro256ssRNG{s0: 0x01d353e5f3993bb0, s1: 0x7b9c0df6cb193b20, s2: 0xfdfcaa91110765b6, s3: 0xd2db341f10bb232e}
	var r uint64

	/* this should produce:
	0000  dd 51 b2 b7 d9 30 3a 37  eb d9 63 66 a6 70 fd 50  |.Q...0:7..cf.p.P|
	0010  26 e7 29 1f 21 21 c0 35  36 c1 2d 03 77 b1 41 d3  |&.).!!.56.-.w.A.|
	0020  43 33 2f 77 f7 fe 97 01  1e 93 c3 ce e4 df fc c4  |C3/w............|
	0030  db 6c 06 54 08 25 6f 5a  0e 86 82 4d 1c 72 c9 50  |.l.T.%oZ...M.r.P|
	0040  20 ae ca 84 d9 24 87 b9  51 96 93 ae ae d2 8f ce  | ....$..Q.......|
	0050  57 37 c1 5c f4 cc 5c d6  2a 29 72 cb f0 c5 f8 f8  |W7.\..\.*)r.....|
	0060  46 1e 33 a2 5d b1 66 b4  15 6f 3b ed 93 e4 70 ba  |F.3.].f..o;...p.|
	0070  11 be 24 b0 20 64 13 86  71 72 92 31 d8 be 03 a9  |..$. d..qr.1....|
	0080  78 6f 73 68 69 72 6f 32  35 36 2a 2a 20 62 79 20  |xoshiro256** by |
	0090  42 6c 61 63 6b 6d 61 6e  20 26 20 56 69 67 6e 61  |Blackman & Vigna|
	00a0  bd 9a f9 bd 3a 79 52 d3  76 50 5e 1e 55 6a 36 48  |....:yR.vP^.Uj6H|
	00b0  9f c0 39 c2 5c db 99 a3  5c d5 4b a2 15 35 53 9c  |..9.\...\.K..5S.|
	00c0  da dd c6 0b bf 33 ef a7  82 eb 06 52 6d 6d 31 2b  |.....3.....Rmm1+|
	00d0  24 7a 0c 3f 70 43 d1 6f  aa c6 88 7e f9 30 ee ff  |$z.?pC.o...~.0..|
	00e0  22 31 af c6 1f e5 68 22  e9 6e 30 06 f6 7f 9a 6e  |"1....h".n0....n|
	00f0  be 19 0c f7 ae e2 fa ec  8e c6 22 e1 78 b6 39 d1  |..........".x.9.|
	*/

	r = rng.Uint64()
	assert.Equal(t, r, endian.HostToNetUint64(uint64(0xdd51b2b7d9303a37)))
	r = rng.Uint64()
	assert.Equal(t, r, endian.HostToNetUint64(uint64(0xebd96366a670fd50)))
}

func Test_NewUnsafeRandRNG_UInt64(t *testing.T) {
	rng := NewUnsafeRandRNG(1)
	r := rng.Uint64()
	assert.Equal(t, rand.New(rand.NewSource(1).(rand.Source64)).Uint64(), r)
}

// ///////////////////////////////////////////////////////////////////////////
//
//    B E N C H M A R K S

var BenchSink interface{}

func Benchmark_UnsafeXoshiro256ssRNG(b *testing.B) {
	rng := NewUnsafeXoshiro256ssRNG(time.Now().UnixNano())
	var r uint64
	for i := 0; i < b.N; i++ {
		r = rng.Uint64()
	}
	BenchSink = &r
}

func Benchmark_UnsafeRandRNG(b *testing.B) {
	rng := NewUnsafeRandRNG(time.Now().UnixNano())
	var r uint64
	for i := 0; i < b.N; i++ {
		r = rng.Uint64()
	}
	BenchSink = &r
}

func Benchmark_UnsafePcg32RNG(b *testing.B) {
	rng := NewUnsafePcg32RNG(time.Now().UnixNano())
	var r uint32
	for i := 0; i < b.N; i++ {
		r = rng.Uint32()
	}
	BenchSink = &r
}

func Benchmark_FastModulo(b *testing.B) {
	maxN := uint64(10)
	rng := NewUnsafeRandRNG(time.Now().UnixNano())
	var r uint32
	for i := 0; i < b.N; i++ {
		x := rng.Uint64() & 0x00000000FFFFFFFF
		r = uint32((x * (maxN)) >> 32)
	}
	BenchSink = &r
}

func Benchmark_Modulo(b *testing.B) {
	maxN := uint64(10)
	rng := NewUnsafeRandRNG(time.Now().UnixNano())
	var r uint32
	for i := 0; i < b.N; i++ {
		r = uint32(rng.Uint64() % (maxN))
	}
	BenchSink = &r
}

func Benchmark_SyncPoolXoshiro256ssRNG_Uint32n_Serial(b *testing.B) {
	rng := NewSyncPoolXoshiro256ssRNG()
	var r uint32
	for i := 0; i < b.N; i++ {
		r = rng.Uint32n(10)
	}
	BenchSink = &r
}

func Benchmark_SyncPoolXoshiro256ssRNG_Uint32n_Parallel(b *testing.B) {
	rng := NewSyncPoolXoshiro256ssRNG()
	b.RunParallel(func(pb *testing.PB) {
		r := rng.Uint32n(10)
		for pb.Next() {
			r = rng.Uint32n(10)
		}
		BenchSink = &r
	})
}

func Benchmark_SyncPoolXoshiro256ssRNG_Uint64_Serial(b *testing.B) {
	rng := NewSyncPoolXoshiro256ssRNG()
	var r uint64
	for i := 0; i < b.N; i++ {
		r = rng.Uint64()
	}
	BenchSink = &r
}

func Benchmark_SyncPoolUnsafeRandRNG_Uint64_Serial(b *testing.B) {
	rand.Seed(1)
	rng := NewSyncPoolRNG(func() UnsafeRNG { return NewUnsafeRandRNG(int64(rand.Uint64())) })
	var r uint64
	for i := 0; i < b.N; i++ {
		r = rng.Uint64()
	}
	BenchSink = &r
}

func Benchmark_SyncPoolXoshiro256ssRNG_Uint64_Parallel(b *testing.B) {
	rng := NewSyncPoolXoshiro256ssRNG()
	b.RunParallel(func(pb *testing.PB) {
		r := rng.Uint64()
		for pb.Next() {
			r = rng.Uint64()
		}
		BenchSink = &r
	})
}

func Benchmark_SyncPoolUnsafeRandRNG_Uint64_Parallel(b *testing.B) {
	rand.Seed(1)
	rng := NewSyncPoolRNG(func() UnsafeRNG { return NewUnsafeRandRNG(int64(rand.Uint64())) })
	b.RunParallel(func(pb *testing.PB) {
		r := rng.Uint64()
		for pb.Next() {
			r = rng.Uint64()
		}
		BenchSink = &r
	})
}

func Benchmark_Rand_Int31n_Serial(b *testing.B) {
	var r int32
	for i := 0; i < b.N; i++ {
		r = rand.Int31n(10)
	}
	BenchSink = &r
}

func Benchmark_Rand_Int31n_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		r := rand.Int31n(10)
		for pb.Next() {
			r = rand.Int31n(10)
		}
		BenchSink = &r
	})
}
func Benchmark_ValyalaFastrand_Int31n_Serial(b *testing.B) {
	r := fastrand.Uint32n(10)
	for i := 0; i < b.N; i++ {
		r = fastrand.Uint32n(10)
	}
	BenchSink = &r
}

func Benchmark_ValyalaFastrand_Int31n_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		r := fastrand.Uint32n(10)
		for pb.Next() {
			r = fastrand.Uint32n(10)
		}
		BenchSink = &r
	})
}
func Benchmark_Rand_Uint64_Serial(b *testing.B) {
	var r uint64
	for i := 0; i < b.N; i++ {
		r = rand.Uint64()
	}
	BenchSink = &r
}

func Benchmark_Rand_Uint64_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		r := rand.Uint64()
		for pb.Next() {
			r = rand.Uint64()
		}
		BenchSink = &r
	})
}

func Benchmark_SyncPoolBytes_Serial_64bytes(b *testing.B) {
	rng := NewSyncPoolXoshiro256ssRNG()
	for i := 0; i < b.N; i++ {
		bytes := rng.Bytes(64)
		assert.Equal(b, 64, len(bytes))
	}
}

func Benchmark_SyncPoolBytes_Serial_1024bytes(b *testing.B) {
	rng := NewSyncPoolXoshiro256ssRNG()
	for i := 0; i < b.N; i++ {
		bytes := rng.Bytes(1024)
		assert.Equal(b, 1024, len(bytes))
	}
}

func Benchmark_SyncPoolBytes_Parallel_1024bytes(b *testing.B) {
	rng := NewSyncPoolXoshiro256ssRNG()
	b.RunParallel(func(pb *testing.PB) {
		bytes := rng.Bytes(1024)
		for pb.Next() {
			bytes = rng.Bytes(1024)
		}
		BenchSink = &bytes
	})
}

/*
goos: windows
goarch: amd64
pkg: github.com/villenny/concurrency-go
Benchmark_UnsafeXoshiro256ssRNG
Benchmark_UnsafeXoshiro256ssRNG-8                       886575583                2.65 ns/op            0 B/op          0 allocs/op
Benchmark_UnsafeRandRNG
Benchmark_UnsafeRandRNG-8                               331377972                7.34 ns/op            0 B/op          0 allocs/op
Benchmark_FastModulo
Benchmark_FastModulo-8                                  321618081                7.49 ns/op            0 B/op          0 allocs/op
Benchmark_Modulo
Benchmark_Modulo-8                                      290857617                8.18 ns/op            0 B/op          0 allocs/op
Benchmark_SyncPoolXoshiro256ssRNG_Uint32n_Serial
Benchmark_SyncPoolXoshiro256ssRNG_Uint32n_Serial-8      66759944                35.4 ns/op             0 B/op          0 allocs/op
Benchmark_SyncPoolXoshiro256ssRNG_Uint32n_Parallel
Benchmark_SyncPoolXoshiro256ssRNG_Uint32n_Parallel-8    283647373                8.68 ns/op            0 B/op          0 allocs/op
Benchmark_SyncPoolXoshiro256ssRNG_Uint64_Serial
Benchmark_SyncPoolXoshiro256ssRNG_Uint64_Serial-8       68642816                35.0 ns/op             0 B/op          0 allocs/op
Benchmark_SyncPoolUnsafeRandRNG_Uint64_Serial
Benchmark_SyncPoolUnsafeRandRNG_Uint64_Serial-8         64929618                37.9 ns/op             0 B/op          0 allocs/op
Benchmark_SyncPoolXoshiro256ssRNG_Uint64_Parallel
Benchmark_SyncPoolXoshiro256ssRNG_Uint64_Parallel-8     276466162                8.75 ns/op            0 B/op          0 allocs/op
Benchmark_SyncPoolUnsafeRandRNG_Uint64_Parallel
Benchmark_SyncPoolUnsafeRandRNG_Uint64_Parallel-8       263430684                8.87 ns/op            0 B/op          0 allocs/op
Benchmark_Rand_Int31n_Serial
Benchmark_Rand_Int31n_Serial-8                          140250397               17.3 ns/op             0 B/op          0 allocs/op
Benchmark_Rand_Int31n_Parallel
Benchmark_Rand_Int31n_Parallel-8                        24767929                94.8 ns/op             0 B/op          0 allocs/op
Benchmark_ValyalaFastrand_Int31n_Serial
Benchmark_ValyalaFastrand_Int31n_Serial-8               100000000               24.5 ns/op             0 B/op          0 allocs/op
Benchmark_ValyalaFastrand_Int31n_Parallel
Benchmark_ValyalaFastrand_Int31n_Parallel-8             340295736                8.60 ns/op            0 B/op          0 allocs/op
Benchmark_Rand_Uint64_Serial
Benchmark_Rand_Uint64_Serial-8                          160272944               14.8 ns/op             0 B/op          0 allocs/op
Benchmark_Rand_Uint64_Parallel
Benchmark_Rand_Uint64_Parallel-8                        26694428                88.7 ns/op             0 B/op          0 allocs/op
Benchmark_SyncPoolBytes_Serial_64bytes
Benchmark_SyncPoolBytes_Serial_64bytes-8                 2222467              1069 ns/op             288 B/op          4 allocs/op
Benchmark_SyncPoolBytes_Serial_1024bytes
Benchmark_SyncPoolBytes_Serial_1024bytes-8               1000000              2248 ns/op            1248 B/op          4 allocs/op
Benchmark_SyncPoolBytes_Parallel_1024bytes
Benchmark_SyncPoolBytes_Parallel_1024bytes-8             4044595               543 ns/op            1024 B/op          1 allocs/op
PASS
ok      github.com/villenny/concurrency-go      57.292s


*/
