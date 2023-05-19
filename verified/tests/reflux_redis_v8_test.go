package tests

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	redisV8 "github.com/things-go/limiter/verified/redis/v8"
)

func TestReflux_RedisV8_Improve_Cover(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()
	testReflux_Improve_Cover(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestReflux_RedisV8_RedisUnavailable(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	addr := mr.Addr()
	mr.Close()
	testReflux_RedisUnavailable(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: addr})),
	)
}

func TestReflux_RedisV8_One_Time(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	testReflux_One_Time(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestReflux_RedisV8_In_Quota(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	testReflux_In_Quota(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestReflux_RedisV8_Over_Quota(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	testReflux_Over_Quota(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

// TODO: success in redis, but failed in miniredis
// func TestReflux_RedisV8_OneTime_Timeout(t *testing.T) {
// 	mr, err := miniredis.Run()
// 	assert.NoError(t, err)

// 	defer mr.Close()

// 	testReflux_OneTime_Timeout(
// 		t,
// 		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
// 		// redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "123456", DB: 4})),
// 	)
// }
