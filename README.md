[![GitHub issues](https://img.shields.io/github/issues/Villenny/concurrency-go)](https://github.com/Villenny/concurrency-go/issues)
[![GitHub forks](https://img.shields.io/github/forks/Villenny/concurrency-go)](https://github.com/Villenny/concurrency-go/network)
[![GitHub stars](https://img.shields.io/github/stars/Villenny/concurrency-go)](https://github.com/Villenny/concurrency-go/stargazers)
[![GitHub license](https://img.shields.io/github/license/Villenny/concurrency-go)](https://github.com/Villenny/concurrency-go/blob/master/LICENSE)
![Go](https://github.com/Villenny/concurrency-go/workflows/Go/badge.svg?branch=master)
![Codecov branch](https://img.shields.io/codecov/c/github/villenny/concurrency-go/master)

# concurrency-go
Helper library for effective concurrency in golang


## Install

```
go get -u github.com/Villenny/concurrency-go
```

## Notable members:
`ParallelForLimit()`, 
`AtomicInt64`, 
`Pool`

Using ParallelForLimit:
- Quickly use all the cores to process an array of work.
```
	import "github.com/villenny/concurrency-go"

	results := make([]MyStruct, len(input))

	ParallelForLimit(runtime.NumCPU(), len(input), func(n int) {
		in := input[n]
		results[n] = MyWorkFn(in)
	})

```

Using AtomicInt64:
- Supports json marshal/unmarshal implicitly, just like regular int64's
- Automatically eliminates false sharing
```
	import "github.com/villenny/concurrency-go"

	allocCount := NewAtomicInt64()
	allocCount.Add(1)
	bytes, _ := json.Marshal(&allocCount)
	_ = json.Unmarshal(bytes, &allocCount)
	itsOne := allocCount.Get()
```

Using the Pool:
- Simplifies use by handling reset on put implicitly
```
import (
	"github.com/villenny/concurrency-go"
	"github.com/cespare/xxhash/v2"
)

var pool = concurrency.NewPool(&concurrency.PoolInfo{
	New:   func() interface{} { return xxhash.New() },
	Reset: func(t interface{}) { t.(*xxhash.Digest).Reset() },
})

func HashString(s string) uint64 {
	digest := pool.Get().(*xxhash.Digest)
	_, _ = digest.WriteString(s) // always returns len(s), nil
	hash := digest.Sum64()
	pool.Put(digest)
	return hash
}
```

## Benchmark

Assuming you never call go fn() inside your work function, ParallelForLimit() is pretty tough to beat assuming you're using it for batch processing which is its intended use case.

Unfortunately without the ability to do something along the lines of go thiscore fn(), theres no way to do this optimally if you have asynchronous calls in your work function.

And while its easy to use, you absolutely can beat it, by feeding your input into one disruptor per long lived goroutine so each has totally independent in order circular array buffers with zero contention.

```
Running tool: C:\Go\bin\go.exe test -benchmem -run=^$ github.com/villenny/concurrency-go -bench .

goos: windows
goarch: amd64
pkg: github.com/villenny/concurrency-go
BenchmarkFor_InlineFor_SQRT-8                   	231453978	         5.42 ns/op	       0 B/op	       0 allocs/op
BenchmarkFor_SerialFor_SQRT-8                   	327313581	         4.06 ns/op	       0 B/op	       0 allocs/op
BenchmarkFor_InlineFor_SQRT2-8                  	353307472	         3.54 ns/op	       0 B/op	       0 allocs/op
BenchmarkFor_ParallelForLimit_SQRT_1-8          	326426421	         3.51 ns/op	       0 B/op	       0 allocs/op
BenchmarkFor_ParallelForLimit_SQRT_2-8          	556136150	         2.87 ns/op	       0 B/op	       0 allocs/op
BenchmarkFor_ParallelForLimit_SQRT_4-8          	704042665	        10.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkFor_ParallelForLimit_SQRT_8-8          	751173544	        11.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkFor_ParallelForLimit_SQRT_16-8         	200899519	         7.09 ns/op	       0 B/op	       0 allocs/op
BenchmarkFor_ParallelFor_SQRT-8                 	 5315282	       230 ns/op	       0 B/op	       0 allocs/op
BenchmarkFor_SerialFor_SHA256-8                 	 1696674	       702 ns/op	      32 B/op	       1 allocs/op
BenchmarkFor_ParallelForLimit_SHA256-8          	 5976416	       205 ns/op	      32 B/op	       1 allocs/op
BenchmarkInt64_Add-8                            	188578995	         6.10 ns/op	       0 B/op	       0 allocs/op
BenchmarkInt64_Get-8                            	552139134	         2.18 ns/op	       8 B/op	       0 allocs/op
BenchmarkSafeInt64_Add-8                        	 7507633	       160 ns/op	       0 B/op	       0 allocs/op
BenchmarkSafeInt64_Get-8                        	28602208	        35.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkAtomicInt64_Add-8                      	75078831	        16.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkAtomicInt64_Get-8                      	1000000000	         1.53 ns/op	       0 B/op	       0 allocs/op
Benchmark_Inc500x2_SafeInt64-8                  	   10000	    114481 ns/op	     491 B/op	       1 allocs/op
Benchmark_Get500x2_Int64-8                      	 3048847	       391 ns/op	      31 B/op	       0 allocs/op
Benchmark_Get500x2_SafeInt64-8                  	   58312	     23007 ns/op	       0 B/op	       0 allocs/op
Benchmark_Inc500x2_AtomicInt64-8                	  138075	      8581 ns/op	       0 B/op	       0 allocs/op
Benchmark_Get500x2_AtomicInt64-8                	 1568206	       766 ns/op	      50 B/op	       0 allocs/op
Benchmark_Inc500x2_RawAtomic_falseSharing-8     	   55357	     21456 ns/op	       0 B/op	       0 allocs/op
Benchmark_Inc500x2_RawAtomic_noFalseSharing-8   	  138074	      8711 ns/op	       0 B/op	       0 allocs/op
Benchmark_Inc500x2_Int64_noFalseSharing-8       	  266944	      4558 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/villenny/concurrency-go	89.611s
```

## Contact

Ryan Haksi [ryan.haksi@gmail.com]

## License

Available under the MIT [License](/LICENSE).
