# Rediscache benchmark

## remote redis + local datastore emulator

```
goos: linux
goarch: amd64
pkg: rediscachebench
BenchmarkGetByRedigoTxPipeline-8             	      88	  14755338 ns/op
BenchmarkGetPipeline-8                       	      87	  14367165 ns/op
BenchmarkGetTxPipeline-8                     	      76	  19569233 ns/op
BenchmarkMGet-8                              	      84	  15134550 ns/op
BenchmarkGetMultiByMercariRedisCache-8       	      63	  16707716 ns/op
BenchmarkGetMultiByJuntakiRedisCache-8       	      72	  16919890 ns/op
PASS
ok  	rediscachebench	16.747s
```

## local redis + local datastore emulator

```
goos: linux
goarch: amd64
pkg: rediscachebench
BenchmarkGetByRedigoTxPipeline-8             	   10000	    163214 ns/op
BenchmarkGetPipeline-8                       	    5493	    222822 ns/op
BenchmarkGetTxPipeline-8                     	    4617	    297613 ns/op
BenchmarkMGet-8                              	    6728	    178958 ns/op
BenchmarkGetMultiByMercariRedisCache-8       	    8868	    254791 ns/op
BenchmarkGetMultiByJuntakiRedisCache-8       	    7587	    141707 ns/op
PASS
ok  	rediscachebench	12.085s
```