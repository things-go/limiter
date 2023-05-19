package tests

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	redisV9 "github.com/things-go/limiter/limit_verified/redis/v9"
)

func Test_RedisV9_Name(t *testing.T) {
	testName(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})),
	)
}

func Test_RedisV9_SendCode_RedisUnavailable(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	addr := mr.Addr()
	mr.Close()
	testSendCode_RedisUnavailable(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: addr})),
	)
}

func Test_RedisV9_SendCode_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testSendCode_Success(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_SendCode_Err_Failure(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testSendCode_Err_Failure(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_SendCode_MaxSendPerDay(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testSendCode_MaxSendPerDay(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_SendCode_Concurrency_MaxSendPerDay(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testSendCode_Concurrency_MaxSendPerDay(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_SendCode_ResendTooFrequently(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testSendCode_ResendTooFrequently(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_SendCode_Concurrency_ResendTooFrequently(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testSendCode_Concurrency_ResendTooFrequently(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_VerifyCode_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testVerifyCode_Success(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_VerifyCode_CodeRequired(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testVerifyCode_CodeRequired(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

// TODO: mini redis 测试失败, 但redis是成功的
// func Test_RedisV9_VerifyCode_CodeExpired(t *testing.T) {
// 	mr, err := miniredis.Run()
// 	require.Nil(t, err)
// 	defer mr.Close()

// 	testVerifyCode_CodeExpired(
// 		t,
// 		redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
// 	)
// }

func Test_RedisV9_VerifyCode_CodeMaxError(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testVerifyCode_CodeMaxError(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_VerifyCode_Concurrency_CodeMaxError(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	testVerifyCode_Concurrency_CodeMaxError(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9__INCR_MaxSendPerDay(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	test_INCR_MaxSendPerDay(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9__INCR_DECR(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	test_INCR_DECR(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}
