package rediscachebench

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	redigo "github.com/gomodule/redigo/redis"

	"os"
	"sync"
	"testing"
	"time"
)

const n = 10

var redisClient *redis.Client
var message = "a"

func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}

func BenchmarkGetByRedigoTxPipeline(b *testing.B) {
	for i := 0; i < n; i++ {
		redisClient.Set(fmt.Sprintln(i), message, time.Minute)
	}
	defer func() {
		_, err := redisClient.FlushAll().Result()
		if err != nil {
			b.Fatal(err)
		}
	}()

	conn, err := redigo.Dial("tcp", os.Getenv("REDIS_HOST")+":"+os.Getenv("REDIS_PORT"),
		redigo.DialPassword(os.Getenv("REDIS_PASSWORD")))
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := conn.Send("MULTI")
		if err != nil {
			b.Fatal(err)
		}
		for i := 0; i < n; i++ {
			err := conn.Send("GET", fmt.Sprintln(i))
			if err != nil {
				b.Fatal(err)
			}

		}
		_, err = conn.Do("EXEC")
		if err != nil {
			b.Fatal(err)
		}
		conn.Flush()
	}
}

func BenchmarkGetByRedigoTxPipelineConcurrent(b *testing.B) {
	b.Skip() // it may fail
	for i := 0; i < n; i++ {
		redisClient.Set(fmt.Sprintln(i), message, time.Minute)
	}
	defer func() {
		_, err := redisClient.FlushAll().Result()
		if err != nil {
			b.Fatal(err)
		}
	}()

	conn, err := redigo.Dial("tcp", os.Getenv("REDIS_HOST")+":"+os.Getenv("REDIS_PORT"),
		redigo.DialPassword(os.Getenv("REDIS_PASSWORD")))
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			err := conn.Send("MULTI")
			if err != nil {
				b.Fatal(err)
			}
			for j := 0; j < n/2; j++ {
				err := conn.Send("GET", fmt.Sprintln(j))
				if err != nil {
					b.Fatal(err)
				}
			}
			_, err = conn.Do("EXEC")
			if err != nil {
				b.Fatal(err)
			}
			conn.Flush()
			wg.Done()
		}()
		go func() {
			err := conn.Send("MULTI")
			if err != nil {
				b.Fatal(err)
			}
			for j := 0; j < n/2; j++ {
				err := conn.Send("GET", fmt.Sprintln(j))
				if err != nil {
					b.Fatal(err)
				}
			}
			_, err = conn.Do("EXEC")
			if err != nil {
				b.Fatal(err)
			}
			conn.Flush()
			wg.Done()
		}()
		// wg.Wait()
	}

}

func BenchmarkGet(b *testing.B) {
	b.Skip() // Too slow
	for i := 0; i < n; i++ {
		redisClient.Set(fmt.Sprintln(i), message, time.Minute)
	}
	defer func() {
		_, err := redisClient.FlushAll().Result()
		if err != nil {
			b.Fatal(err)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < n; i++ {
			redisClient.Get(fmt.Sprintln(i))
		}
	}
}

func BenchmarkGetConcurrent(b *testing.B) {
	b.Skip() // Slow
	for i := 0; i < n; i++ {
		redisClient.Set(fmt.Sprintln(i), message, time.Minute)
	}
	defer func() {
		_, err := redisClient.FlushAll().Result()
		if err != nil {
			b.Fatal(err)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg := sync.WaitGroup{}
		wg.Add(n)
		for i := 0; i < n; i++ {
			i := i
			go func() {
				redisClient.Get(fmt.Sprintln(i))
				wg.Done()
			}()
		}
		wg.Wait()
	}

}

func BenchmarkGetPipeline(b *testing.B) {
	for i := 0; i < n; i++ {
		redisClient.Set(fmt.Sprintln(i), message, time.Minute)
	}
	defer func() {
		_, err := redisClient.FlushAll().Result()
		if err != nil {
			b.Fatal(err)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline := redisClient.Pipeline()
		for i := 0; i < n; i++ {
			pipeline.Get(fmt.Sprintln(i))
		}
		pipeline.Exec()
	}
}

func BenchmarkGetTxPipeline(b *testing.B) {
	for i := 0; i < n; i++ {
		redisClient.Set(fmt.Sprintln(i), message, time.Minute)
	}
	defer func() {
		_, err := redisClient.FlushAll().Result()
		if err != nil {
			b.Fatal(err)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline := redisClient.TxPipeline()
		for i := 0; i < n; i++ {
			pipeline.Get(fmt.Sprintln(i))
		}
		pipeline.Exec()
	}
}

func BenchmarkMGet(b *testing.B) {
	var keys []string
	for i := 0; i < n; i++ {
		redisClient.Set(fmt.Sprintln(i), message, time.Minute)
		keys = append(keys, fmt.Sprintln(i))
	}
	defer func() {
		_, err := redisClient.FlushAll().Result()
		if err != nil {
			b.Fatal(err)
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		redisClient.MGet(keys...)
	}
}
