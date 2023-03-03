package v9

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/things-go/limiter/verified"
)

var _ verified.RefluxProvider = (*TestVerifiedRefluxProvider)(nil)

type TestVerifiedRefluxProvider struct{}

func (t TestVerifiedRefluxProvider) Name() string { return "test_provider" }

func (t TestVerifiedRefluxProvider) GenerateUniqueId() string {
	return randString(6)
}

func TestReflux_Improve_Cover(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()
	l := verified.NewVerifiedCaptcha(
		new(TestVerifiedCaptchaProvider),
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
	l.Name(defaultKind)
}

func TestReflux_RedisUnavailable(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)

	l := verified.NewVerifiedReflux(
		new(TestVerifiedRefluxProvider),
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
	mr.Close()

	randKey := randString(6)
	_, err = l.Generate(context.Background(), defaultKind, randKey)
	assert.Error(t, err)
}

func TestReflux_One_Time(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	l := verified.NewVerifiedReflux(
		new(TestVerifiedRefluxProvider),
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
		verified.WithKeyPrefix("verified:reflux:"),
		verified.WithKeyExpires(time.Minute*3),
	)

	randKey := randString(6)

	value, err := l.Generate(context.Background(), defaultKind, randKey, verified.WithGenerateKeyExpires(time.Minute*5))
	assert.NoError(t, err)

	b := l.Verify(context.Background(), defaultKind, randKey, value)
	require.True(t, b)

	b = l.Verify(context.Background(), defaultKind, randKey, value)
	require.False(t, b)
}

func TestReflux_In_Quota(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	l := verified.NewVerifiedReflux(
		new(TestVerifiedRefluxProvider),
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
		verified.WithKeyPrefix("verified:reflux:"),
		verified.WithKeyExpires(time.Minute*3),
		verified.WithMaxErrQuota(3),
	)

	randKey := randString(6)
	value, err := l.Generate(context.Background(), defaultKind, randKey, verified.WithGenerateKeyExpires(time.Minute*5))
	assert.NoError(t, err)

	badValue := value + "xxx"

	b := l.Verify(context.Background(), defaultKind, randKey, badValue)
	require.False(t, b)
	b = l.Verify(context.Background(), defaultKind, randKey, badValue)
	require.False(t, b)
	b = l.Verify(context.Background(), defaultKind, randKey, value)
	require.True(t, b)
}

func TestReflux_Over_Quota(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	l := verified.NewVerifiedReflux(
		new(TestVerifiedRefluxProvider),
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
		verified.WithKeyPrefix("verified:reflux:"),
		verified.WithKeyExpires(time.Minute*3),
		verified.WithMaxErrQuota(3),
	)

	randKey := randString(6)
	value, err := l.Generate(context.Background(), defaultKind, randKey,
		verified.WithGenerateKeyExpires(time.Minute*5),
		verified.WithGenerateMaxErrQuota(6),
	)
	assert.NoError(t, err)

	badValue := value + "xxx"

	for i := 0; i < 6; i++ {
		b := l.Verify(context.Background(), defaultKind, randKey, badValue)
		require.False(t, b)
	}
	b := l.Verify(context.Background(), defaultKind, randKey, value)
	require.False(t, b)
}

// TODO: success in redis, but failed in miniredis
// func TestReflux_OneTime_Timeout(t *testing.T) {
//     mr, err := miniredis.Run()
//     assert.NoError(t, err)
//
//     defer mr.Close()
//
//     l := verified.NewVerifiedReflux(
//         new(TestVerifiedRefluxProvider),
//         NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
//         // NewRedisStore(redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "123456", DB: 0})),
//     )
//     randKey := randString(6)
//     value, err := l.Generate(context.Background(),defaultKind, randKey, verified.WithGenerateKeyExpires(time.Second*1))
//     assert.NoError(t, err)
//
//     time.Sleep(time.Second * 2)
//
//     b := l.Verify(context.Background(),defaultKind, randKey, value)
//     require.False(t, b)
// }
