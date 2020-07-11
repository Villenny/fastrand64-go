[![GitHub issues](https://img.shields.io/github/issues/Villenny/fastrand64-go)](https://github.com/Villenny/fastrand64-go/issues)
[![GitHub forks](https://img.shields.io/github/forks/Villenny/fastrand64-go)](https://github.com/Villenny/fastrand64-go/network)
[![GitHub stars](https://img.shields.io/github/stars/Villenny/fastrand64-go)](https://github.com/Villenny/fastrand64-go/stargazers)
[![GitHub license](https://img.shields.io/github/license/Villenny/fastrand64-go)](https://github.com/Villenny/fastrand64-go/blob/master/LICENSE)
![Go](https://github.com/Villenny/fastrand64-go/workflows/Go/badge.svg?branch=master)
![Codecov branch](https://img.shields.io/codecov/c/github/villenny/fastrand64-go/master)

# fastrand64-go
Helper library for full uint64 randomness, pool backed for efficient concurrency

Inspired by https://github.com/valyala/fastrand which is faster, but not 64 bit.


## Install

```
go get -u github.com/Villenny/fastrand64-go
```

## Notable members:
`Xoshiro256ssRNG`, 
`SyncPoolRNG`, 

The expected use case:
- If you are doing a lot of random indexing on a lot of cores
```
	import "github.com/villenny/fastrand64-go"

	rng := NewSyncPoolXoshiro256ssRNG()

	// somewhere later, in some goproc, one of lots, like a web request handler for example

	r1 := rng.Uint32n(10)
	r2 := rng.Uint64()
	someBytes := rng.Bytes(256)
```

Using SyncPoolRNG:
- I tried to keep everything safe for composition, this way you can use your own random generator if you have one
- Note the pool uses the the builtin golang threadsafe uint64 rand function to generate seeds for each allocated generator in the pool.
```
	import "github.com/villenny/concurrency-go"

	// use the helper function to generate a system rand based source
	rand.Seed(1)
	rng := NewSyncPoolRNG(func() UnsafeRNG { return NewUnsafeRandRNG(int64(rand.Uint64())) })

	// use some random thing that has a Uint64() function 	
	rand.Seed(1)
	rng := NewSyncPoolRNG(func() UnsafeRNG { return rand.New(rand.NewSource(rand.Uint64()).(rand.Source64)) })
	
```


## Benchmark

- Xoshiro256ss is roughly 3X faster than whatever golang uses natively
- The Pool wrapped version of Xoshiro is roughly half as fast as the native threadsafe golang random generator, almost entirely due to the cost of checking into and out of the pool.
- BUT, the pool wrapped Xoshiro generator murders the native in a multicore environment where there would otherwise be lots of contention. 4X faster on my 4 core machine in the pathological case of every core doing nothing but generate random numbers.
- It would probably be faster still (although I havent tested this) to feed each goproc its own unsafe generator in their context arg and not use the pool.

```
goos: windows
goarch: amd64
pkg: github.com/villenny/concurrency-go
Benchmark_UnsafeXoshiro256ssRNG
Benchmark_UnsafeXoshiro256ssRNG-8                       910025752                2.66 ns/op            0 B/op          0 allocs/op
Benchmark_UnsafeRandRNG
Benchmark_UnsafeRandRNG-8                               332294956                7.25 ns/op            0 B/op          0 allocs/op
Benchmark_FastModulo
Benchmark_FastModulo-8                                  309598480                7.49 ns/op            0 B/op          0 allocs/op
Benchmark_Modulo
Benchmark_Modulo-8                                      299561928                8.07 ns/op            0 B/op          0 allocs/op
Benchmark_SyncPoolXoshiro256ssRNG_Uint32n_Serial
Benchmark_SyncPoolXoshiro256ssRNG_Uint32n_Serial-8      58596038                35.5 ns/op             0 B/op          0 allocs/op
Benchmark_SyncPoolXoshiro256ssRNG_Uint32n_Parallel
Benchmark_SyncPoolXoshiro256ssRNG_Uint32n_Parallel-8    286352749                8.43 ns/op            0 B/op          0 allocs/op
Benchmark_SyncPoolXoshiro256ssRNG_Uint64_Serial
Benchmark_SyncPoolXoshiro256ssRNG_Uint64_Serial-8       68647528                35.0 ns/op             0 B/op          0 allocs/op
Benchmark_SyncPoolUnsafeRandRNG_Uint64_Serial
Benchmark_SyncPoolUnsafeRandRNG_Uint64_Serial-8         63221281                37.3 ns/op             0 B/op          0 allocs/op
Benchmark_SyncPoolXoshiro256ssRNG_Uint64_Parallel
Benchmark_SyncPoolXoshiro256ssRNG_Uint64_Parallel-8     291210572               10.2 ns/op             0 B/op          0 allocs/op
Benchmark_SyncPoolUnsafeRandRNG_Uint64_Parallel
Benchmark_SyncPoolUnsafeRandRNG_Uint64_Parallel-8       265468694                8.87 ns/op            0 B/op          0 allocs/op
Benchmark_Rand_Int31n_Serial
Benchmark_Rand_Int31n_Serial-8                          137442118               17.6 ns/op             0 B/op          0 allocs/op
Benchmark_Rand_Int31n_Parallel
Benchmark_Rand_Int31n_Parallel-8                        24767979                95.5 ns/op             0 B/op          0 allocs/op
Benchmark_Rand_Uint64_Serial
Benchmark_Rand_Uint64_Serial-8                          143346825               16.8 ns/op             0 B/op          0 allocs/op
Benchmark_Rand_Uint64_Parallel
Benchmark_Rand_Uint64_Parallel-8                        26994104                89.0 ns/op             0 B/op          0 allocs/op
Benchmark_SyncPoolBytes_Serial_64bytes
Benchmark_SyncPoolBytes_Serial_64bytes-8                 2385789              1019 ns/op             288 B/op          4 allocs/op
Benchmark_SyncPoolBytes_Serial_1024bytes
Benchmark_SyncPoolBytes_Serial_1024bytes-8               1000000              2216 ns/op            1248 B/op          4 allocs/op
Benchmark_SyncPoolBytes_Parallel_1024bytes
Benchmark_SyncPoolBytes_Parallel_1024bytes-8             4113847               587 ns/op            1024 B/op          1 allocs/op
PASS
ok      github.com/villenny/concurrency-go      51.672s

```

## Contact

Ryan Haksi [ryan.haksi@gmail.com]

## License

Available under the MIT [License](/LICENSE).
