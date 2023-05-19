package tests

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	redisV9 "github.com/things-go/limiter/verified/redis/v9"
)

func TestReflux_RedisV9_Improve_Cover(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()
	testReflux_Improve_Cover(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestReflux_RedisV9_RedisUnavailable(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	addr := mr.Addr()
	mr.Close()
	testReflux_RedisUnavailable(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: addr})),
	)
}

func TestReflux_RedisV9_One_Time(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	testReflux_One_Time(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestReflux_RedisV9_In_Quota(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	testReflux_In_Quota(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestReflux_RedisV9_Over_Quota(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	testReflux_Over_Quota(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

// TODO: success in redis, but failed in miniredis
// func TestReflux_RedisV9_OneTime_Timeout(t *testing.T) {
// 	mr, err := miniredis.Run()
// 	assert.NoError(t, err)

// 	defer mr.Close()

// 	testReflux_OneTime_Timeout(
// 		t,
// 		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
// 		// redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "123456", DB: 4})),
// 	)
// }
