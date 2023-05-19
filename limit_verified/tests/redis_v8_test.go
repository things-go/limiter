package tests

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"

	redisV8 "github.com/things-go/limiter/limit_verified/redis/v8"
)

func Test_RedisV8_Name(t *testing.T) {
	testName(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})),
	)
}

func Test_RedisV8_SendCode_RedisUnavailable(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	addr := mr.Addr()
	mr.Close()
	testSendCode_RedisUnavailable(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: addr})),
	)
}

func Test_RedisV8_SendCode_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testSendCode_Success(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8_SendCode_Err_Failure(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testSendCode_Err_Failure(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8_SendCode_MaxSendPerDay(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testSendCode_MaxSendPerDay(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8_SendCode_Concurrency_MaxSendPerDay(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testSendCode_Concurrency_MaxSendPerDay(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8_SendCode_ResendTooFrequently(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testSendCode_ResendTooFrequently(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8_SendCode_Concurrency_ResendTooFrequently(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testSendCode_Concurrency_ResendTooFrequently(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8_VerifyCode_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testVerifyCode_Success(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8_VerifyCode_CodeRequired(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testVerifyCode_CodeRequired(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

// TODO: mini redis 测试失败, 但redis是成功的
// func Test_RedisV8_VerifyCode_CodeExpired(t *testing.T) {
// 	mr, err := miniredis.Run()
// 	require.Nil(t, err)
// 	defer mr.Close()

// 	testVerifyCode_CodeExpired(
// 		t,
// 		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
// 	)
// }

func Test_RedisV8_VerifyCode_CodeMaxError(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testVerifyCode_CodeMaxError(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8_VerifyCode_Concurrency_CodeMaxError(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testVerifyCode_Concurrency_CodeMaxError(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8__INCR_MaxSendPerDay(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	test_INCR_MaxSendPerDay(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8__INCR_DECR(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	test_INCR_DECR(
		t,
		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}
