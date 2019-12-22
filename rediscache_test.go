package rediscachebench

import (
	"context"
	"github.com/go-redis/redis/v7"
	redigo "github.com/gomodule/redigo/redis"
	jrediscache "github.com/juntaki/datastore/dsmiddleware/rediscache"
	"go.mercari.io/datastore"
	"go.mercari.io/datastore/clouddatastore"
	mrediscache "go.mercari.io/datastore/dsmiddleware/rediscache"
	"os"
	"strings"
	"testing"
)

// Copy from mercari/datastore/testutils
var EmitCleanUpLog = false

func SetupCloudDatastore(t *testing.B) (context.Context, datastore.Client, func()) {
	ctx := context.Background()
	client, err := clouddatastore.FromContext(ctx)
	if err != nil {
		t.Fatal(err)
	}

	return ctx, client, func() {
		defer client.Close()

		q := client.NewQuery("__kind__").KeysOnly()
		keys, err := client.GetAll(ctx, q, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(keys) == 0 {
			return
		}

		kinds := make([]string, 0, len(keys))
		for _, key := range keys {
			kinds = append(kinds, key.Name())
		}

		if EmitCleanUpLog {
			t.Logf("remove %s", strings.Join(kinds, ", "))
		}

		for _, kind := range kinds {

			cnt := 0
			for {
				q := client.NewQuery(kind).Limit(1000).KeysOnly()
				keys, err := client.GetAll(ctx, q, nil)
				if err != nil {
					t.Fatal(err)
				}
				err = client.DeleteMulti(ctx, keys)
				if err != nil {
					t.Fatal(err)
				}

				cnt += len(keys)

				if len(keys) != 1000 {
					if EmitCleanUpLog {
						t.Logf("remove %s entity: %d", kind, cnt)
					}
					break
				}
			}
		}
	}
}

func BenchmarkGetMultiByMercariRedisCache(b *testing.B) {
	ctx, client, cleanUp := SetupCloudDatastore(b)
	defer cleanUp()

	conn, err := redigo.Dial("tcp", os.Getenv("REDIS_HOST")+":"+os.Getenv("REDIS_PORT"), redigo.DialPassword(os.Getenv("REDIS_PASSWORD")))
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()
	ch := mrediscache.New(
		conn,

	)
	client.AppendMiddleware(ch)
	defer func() {
		_, err := conn.Do("FLUSHALL")
		if err != nil {
			b.Fatal(err)
		}
		// stop logging before cleanUp func called.
		client.RemoveMiddleware(ch)
	}()

	// exec.

	type Data struct {
		Name string
	}

	// Put. add to cache.
	key := client.IDKey("Data", 111, nil)
	objBefore := &Data{Name: "Data"}
	_, err = client.Put(ctx, key, objBefore)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	// Get. from cache.
	for i := 0; i < b.N; i++ {
		objAfter := &Data{}
		err = client.Get(ctx, key, objAfter)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetMultiByJuntakiRedisCache(b *testing.B) {
	ctx, client, cleanUp := SetupCloudDatastore(b)
	defer cleanUp()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ch := jrediscache.New(
		redisClient,
	)
	client.AppendMiddleware(ch)
	defer func() {
		_, err := redisClient.FlushAll().Result()
		if err != nil {
			b.Fatal(err)
		}
		// stop logging before cleanUp func called.
		client.RemoveMiddleware(ch)
	}()

	// exec.

	type Data struct {
		Name string
	}

	// Put. add to cache.
	key := client.IDKey("Data", 111, nil)
	objBefore := &Data{Name: "Data"}
	_, err := client.Put(ctx, key, objBefore)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	// Get. from cache.
	for i := 0; i < b.N; i++ {

		objAfter := &Data{}
		err = client.Get(ctx, key, objAfter)
		if err != nil {
			b.Fatal(err)
		}
	}
}
