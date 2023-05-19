package v8

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/things-go/limiter/verified/tests"
)

func TestCaptcha_RedisV8_Improve_Cover(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()
	tests.GenericTestCaptcha_Improve_Cover(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestCaptcha_RedisV8_Unsupported_Driver(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	addr := mr.Addr()
	mr.Close()
	tests.GenericTestCaptcha_Unsupported_Driver(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: addr})),
	)
}

func TestCaptcha_RedisV8_RedisUnavailable(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	addr := mr.Addr()
	mr.Close()
	tests.GenericTestCaptcha_RedisUnavailable(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: addr})),
	)
}

func TestCaptcha_RedisV8_OneTime(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	tests.GenericTestCaptcha_OneTime(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestCaptcha_RedisV8_In_Quota(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	tests.GenericTestCaptcha_In_Quota(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestCaptcha_RedisV8_Over_Quota(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	tests.GenericTestCaptcha_Over_Quota(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

// TODO: success in redis, but failed in miniredis
// func TestCaptcha_RedisV8_Onetime_Timeout(t *testing.T) {
// 	mr, err := miniredis.Run()
// 	assert.NoError(t, err)

// 	defer mr.Close()

// 	testCaptcha_Onetime_Timeout(
// 		t,
// 		// NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
// 		NewRedisStore(redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "123456", DB: 0})),
// 	)
// }
