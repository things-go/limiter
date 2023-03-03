package v9

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	redisScript "github.com/things-go/limiter/limit"
)

const (
	seconds = time.Second
	quota   = 5
	total   = 100
)

func TestPeriodLimit_Take(t *testing.T) {
	testPeriodLimit(t,
		WithKeyPrefix("limit:period"),
		WithPeriod(seconds),
		WithQuota(quota),
	)
}

func TestPeriodLimit_TakeWithAlign(t *testing.T) {
	testPeriodLimit(t,
		WithKeyPrefix("limit:period"),
		WithAlign(),
		WithPeriod(seconds),
		WithQuota(quota),
	)
}

func TestPeriodLimit_RedisUnavailable(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	l := NewPeriodLimit(
		redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	)
	mr.Close()
	val, err := l.Take(context.Background(), "first")
	assert.Error(t, err)
	assert.Equal(t, redisScript.PeriodLimitStsUnknown, val)
}

func testPeriodLimit(t *testing.T, opts ...PeriodLimitOption) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	l := NewPeriodLimit(
		redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		opts...,
	)
	var allowed, hitQuota, overQuota int
	for i := 0; i < total; i++ {
		val, err := l.Take(context.Background(), "first")
		assert.NoError(t, err)
		switch val {
		case redisScript.PeriodLimitStsAllowed:
			allowed++
		case redisScript.PeriodLimitStsHitQuota:
			hitQuota++
		case redisScript.PeriodLimitStsOverQuota:
			overQuota++
		case redisScript.PeriodLimitStsUnknown:
			fallthrough
		default:
			t.Error("unknown status")
		}
	}

	assert.Equal(t, quota-1, allowed)
	assert.Equal(t, 1, hitQuota)
	assert.Equal(t, total-quota, overQuota)
}

func TestPeriodLimit_QuotaFull(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	l := NewPeriodLimit(
		redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		WithPeriod(1),
		WithQuota(1),
	)
	val, err := l.Take(context.Background(), "first")
	assert.NoError(t, err)
	assert.True(t, val.IsHitQuota())
}

func TestPeriodLimit_SetQuotaFull(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	l := NewPeriodLimit(
		redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	)

	err = l.SetQuotaFull(context.Background(), "first")
	assert.NoError(t, err)

	val, err := l.Take(context.Background(), "first")
	assert.NoError(t, err)
	assert.Equal(t, redisScript.PeriodLimitStsOverQuota, val)
}

func TestPeriodLimit_Del(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	l := NewPeriodLimit(
		redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		WithPeriod(seconds),
		WithQuota(quota),
	)

	// 第一次, key不存在
	v, b, err := l.GetInt(context.Background(), "first")
	assert.NoError(t, err)
	assert.False(t, b)
	assert.Equal(t, 0, v)

	tt, err := l.TTL(context.Background(), "first")
	assert.NoError(t, err)
	assert.Equal(t, int(tt), -2)

	runValue, err := l.GetRunValue(context.Background(), "first")
	assert.Nil(t, err)
	assert.Equal(t, runValue.Exist, false)
	assert.Equal(t, runValue.Count, int64(0))
	assert.Equal(t, runValue.TTL, time.Duration(0))

	err = l.SetQuotaFull(context.Background(), "first")
	assert.NoError(t, err)

	// 第二次, key 存在
	v, b, err = l.GetInt(context.Background(), "first")
	assert.NoError(t, err)
	assert.True(t, b)
	assert.Equal(t, quota, v)

	tt, err = l.TTL(context.Background(), "first")
	assert.NoError(t, err)
	assert.LessOrEqual(t, tt, seconds)

	runValue, err = l.GetRunValue(context.Background(), "first")
	assert.Nil(t, err)
	assert.Equal(t, runValue.Exist, true)
	assert.Equal(t, runValue.Count, int64(quota))
	assert.Equal(t, runValue.TTL, seconds)

	val, err := l.Take(context.Background(), "first")
	assert.NoError(t, err)
	assert.True(t, val.IsOverQuota())

	err = l.Del(context.Background(), "first")
	assert.NoError(t, err)

	val, err = l.Take(context.Background(), "first")
	assert.NoError(t, err)
	assert.True(t, val.IsAllowed())
}
