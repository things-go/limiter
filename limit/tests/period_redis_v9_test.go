package tests

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/things-go/limiter/limit"
	redisV9 "github.com/things-go/limiter/limit/redis/v9"
)

var _ limit.PeriodLimitDriver = (*limit.PeriodLimit[*redisV9.PeriodStore])(nil)

func TestPeriodLimit_RedisV9_Take(t *testing.T) {
	testPeriodLimit_RedisV9(t,
		limit.WithKeyPrefix("limit:period"),
		limit.WithPeriod(seconds),
		limit.WithQuota(quota),
	)
}

func TestPeriodLimit_RedisV9_TakeWithAlign(t *testing.T) {
	testPeriodLimit_RedisV9(
		t,
		limit.WithKeyPrefix("limit:period"),
		limit.WithAlign(),
		limit.WithPeriod(seconds),
		limit.WithQuota(quota),
	)
}

func TestPeriodLimit_RedisV9_RedisUnavailable(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	l := limit.NewPeriodLimit(
		redisV9.NewPeriodStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
	)
	mr.Close()
	val, err := l.Take(context.Background(), "first")
	assert.Error(t, err)
	assert.Equal(t, limit.PeriodLimitStsUnknown, val)
}

func testPeriodLimit_RedisV9(t *testing.T, opts ...limit.PeriodLimitOption) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	l := limit.NewPeriodLimit(
		redisV9.NewPeriodStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
		opts...,
	)
	var allowed, hitQuota, overQuota int
	for i := 0; i < total; i++ {
		val, err := l.Take(context.Background(), "first")
		assert.NoError(t, err)
		switch val {
		case limit.PeriodLimitStsAllowed:
			allowed++
		case limit.PeriodLimitStsHitQuota:
			hitQuota++
		case limit.PeriodLimitStsOverQuota:
			overQuota++
		case limit.PeriodLimitStsUnknown:
			fallthrough
		default:
			t.Error("unknown status")
		}
	}

	assert.Equal(t, quota-1, allowed)
	assert.Equal(t, 1, hitQuota)
	assert.Equal(t, total-quota, overQuota)
}

func TestPeriodLimit_RedisV9_QuotaFull(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	l := limit.NewPeriodLimit(
		redisV9.NewPeriodStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
		limit.WithPeriod(1),
		limit.WithQuota(1),
	)
	val, err := l.Take(context.Background(), "first")
	assert.NoError(t, err)
	assert.True(t, val.IsHitQuota())
}

func TestPeriodLimit_RedisV9_SetQuotaFull(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	l := limit.NewPeriodLimit(
		redisV9.NewPeriodStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
	)

	err = l.SetQuotaFull(context.Background(), "first")
	assert.NoError(t, err)

	val, err := l.Take(context.Background(), "first")
	assert.NoError(t, err)
	assert.Equal(t, limit.PeriodLimitStsOverQuota, val)
}

func TestPeriodLimit_RedisV9_Del(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	l := limit.NewPeriodLimit(
		redisV9.NewPeriodStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
		limit.WithPeriod(seconds),
		limit.WithQuota(quota),
	)

	// 第一次, key不存在
	rv, err := l.GetRunValue(context.Background(), "first")
	assert.NoError(t, err)
	assert.False(t, rv.Exist)
	assert.Zero(t, rv.Count)
	assert.Zero(t, int(rv.TTL))

	runValue, err := l.GetRunValue(context.Background(), "first")
	assert.Nil(t, err)
	assert.Equal(t, runValue.Exist, false)
	assert.Equal(t, runValue.Count, int64(0))
	assert.Equal(t, runValue.TTL, time.Duration(0))

	err = l.SetQuotaFull(context.Background(), "first")
	assert.NoError(t, err)

	// 第二次, key 存在
	rv, err = l.GetRunValue(context.Background(), "first")
	assert.NoError(t, err)
	assert.True(t, rv.Exist)
	assert.Equal(t, int64(quota), rv.Count)
	assert.LessOrEqual(t, seconds, rv.TTL)

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
