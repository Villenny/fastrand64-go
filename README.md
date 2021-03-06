[![GitHub issues](https://img.shields.io/github/issues/Villenny/fastrand64-go)](https://github.com/Villenny/fastrand64-go/issues)
[![GitHub forks](https://img.shields.io/github/forks/Villenny/fastrand64-go)](https://github.com/Villenny/fastrand64-go/network)
[![GitHub stars](https://img.shields.io/github/stars/Villenny/fastrand64-go)](https://github.com/Villenny/fastrand64-go/stargazers)
[![GitHub license](https://img.shields.io/github/license/Villenny/fastrand64-go)](https://github.com/Villenny/fastrand64-go/blob/master/LICENSE)
![Go](https://github.com/Villenny/fastrand64-go/workflows/Go/badge.svg?branch=master)
![Codecov branch](https://img.shields.io/codecov/c/github/villenny/fastrand64-go/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/Villenny/fastrand64-go)](https://goreportcard.com/report/github.com/Villenny/fastrand64-go)
[![Documentation](https://godoc.org/github.com/Villenny/fastrand64-go?status.svg)](http://godoc.org/github.com/Villenny/fastrand64-go)

# fastrand64-go
Helper library for full uint64 randomness, pool backed for efficient concurrency

Inspired by https://github.com/valyala/fastrand which is 20% faster, but not as good a source of randomness.


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
- The Pool wrapped version of Xoshiro is roughly half as fast as the native threadsafe golang random generator, almost entirely due to the cost of checking into and out of the pool. But I benchmarked on my own machine, and linux might have a faster sync.Pool.
- BUT, the pool wrapped Xoshiro generator murders the native in a multicore environment where there would otherwise be lots of contention. 4X faster on my 4 core machine in the pathological case of every core doing nothing but generate random numbers.
- It would probably be faster still (although I havent tested this, unsafe xoshiro256** is 10X faster than the pooled xoshiro256**) to feed each goproc its own unsafe generator in their context arg and not use the pool.


```
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


```

## Contact

Ryan Haksi [ryan.haksi@gmail.com]

## License

Available under the MIT [License](/LICENSE).
