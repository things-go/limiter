package v8

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"github.com/things-go/limiter/limit"
)

var internalErr = errors.New("internal error")

func TestPeriodFailureLimit_RedisV8_Check(t *testing.T) {
	testPeriodFailureLimit_RedisV8(t,
		limit.WithKeyPrefix("limit:period:failure:"),
		limit.WithPeriod(seconds),
		limit.WithQuota(quota),
	)
}

func TestPeriodFailureLimit_RedisV8_CheckWithAlign(t *testing.T) {
	testPeriodFailureLimit_RedisV8(t, limit.WithAlign(),
		limit.WithKeyPrefix("limit:period:failure:"),
		limit.WithAlign(),
		limit.WithPeriod(seconds),
		limit.WithQuota(quota),
	)
}

func TestPeriodFailureLimit_RedisV8_RedisUnavailable(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)

	l := limit.NewPeriodFailureLimit(
		NewPeriodFailureStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
	)
	mr.Close()
	sts, err := l.CheckErr(context.Background(), "first", nil)
	assert.Error(t, err)
	assert.Equal(t, limit.PeriodFailureLimitStsUnknown, sts)
}

func testPeriodFailureLimit_RedisV8(t *testing.T, opts ...limit.PeriodLimitOption) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)

	defer mr.Close()

	l := limit.NewPeriodFailureLimit(
		NewPeriodFailureStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
		opts...,
	)
	var inLimitCnt, overFailureTimeCnt int
	for i := 0; i < total; i++ {
		sts, err := l.CheckErr(context.Background(), "first", internalErr)
		assert.NoError(t, err)
		switch sts {
		case limit.PeriodFailureLimitStsInQuota:
			inLimitCnt++
		case limit.PeriodFailureLimitStsOverQuota:
			overFailureTimeCnt++
		default:
			t.Errorf("unknown status, must be on of [%d, %d]", limit.PeriodFailureLimitStsInQuota, limit.PeriodFailureLimitStsOverQuota)
		}
	}
	assert.Equal(t, quota, inLimitCnt)
	assert.Equal(t, total-quota, overFailureTimeCnt)

	sts, err := l.CheckErr(context.Background(), "first", nil)
	assert.NoError(t, err)
	assert.Equal(t, limit.PeriodFailureLimitStsOverQuota, sts)
}

func TestPeriodFailureLimit_RedisV8_Check_In_Limit_Failure_Time_Then_Success(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)

	defer mr.Close()

	l := limit.NewPeriodFailureLimit(
		NewPeriodFailureStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
	)
	var inLimitCnt, overFailureTimeCnt int
	for i := 0; i < quota-1; i++ {
		sts, err := l.CheckErr(context.Background(), "first", internalErr)
		assert.NoError(t, err)
		switch sts {
		case limit.PeriodFailureLimitStsInQuota:
			inLimitCnt++
		case limit.PeriodFailureLimitStsOverQuota:
			overFailureTimeCnt++
		default:
			t.Errorf("unknown status, must be on of [%d, %d]", limit.PeriodFailureLimitStsInQuota, limit.PeriodFailureLimitStsOverQuota)
		}
	}
	assert.Equal(t, quota-1, inLimitCnt)
	assert.Equal(t, 0, overFailureTimeCnt)

	sts, err := l.CheckErr(context.Background(), "first", nil)
	assert.NoError(t, err)
	assert.Equal(t, limit.PeriodFailureLimitStsSuccess, sts)

	rv, err := l.GetRunValue(context.Background(), "first")
	assert.NoError(t, err)
	assert.False(t, rv.Exist)
	assert.Zero(t, rv.Count)
}

func TestPeriodFailureLimit_RedisV8_Check_Over_Limit_Failure_Time_Then_Success_Always_OverFailureTimeError(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)

	defer mr.Close()

	l := limit.NewPeriodFailureLimit(
		NewPeriodFailureStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
		limit.WithQuota(quota),
	)
	var inLimitCnt, overFailureTimeCnt int
	for i := 0; i < quota+1; i++ {
		sts, err := l.CheckErr(context.Background(), "first", internalErr)
		assert.NoError(t, err)
		switch sts {
		case limit.PeriodFailureLimitStsInQuota:
			inLimitCnt++
		case limit.PeriodFailureLimitStsOverQuota:
			overFailureTimeCnt++
		default:
			t.Errorf("unknown status, must be on of [%d, %d]", limit.PeriodFailureLimitStsInQuota, limit.PeriodFailureLimitStsOverQuota)
		}
	}
	assert.Equal(t, quota, inLimitCnt)
	assert.Equal(t, 1, overFailureTimeCnt)

	sts, err := l.CheckErr(context.Background(), "first", nil)
	assert.NoError(t, err)
	assert.Equal(t, limit.PeriodFailureLimitStsOverQuota, sts)

	rv, err := l.GetRunValue(context.Background(), "first")
	assert.NoError(t, err)
	assert.True(t, rv.Exist)
	assert.Equal(t, int64(quota+1), rv.Count)
}

func TestPeriodFailureLimit_RedisV8_SetQuotaFull(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)
	defer mr.Close()

	l := limit.NewPeriodFailureLimit(
		NewPeriodFailureStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
	)

	err = l.SetQuotaFull(context.Background(), "first")
	assert.Nil(t, err)

	sts, err := l.CheckErr(context.Background(), "first", nil)
	assert.NoError(t, err)
	assert.Equal(t, limit.PeriodFailureLimitStsOverQuota, sts)
}

func TestPeriodFailureLimit_RedisV8_Del(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)
	defer mr.Close()

	l := limit.NewPeriodFailureLimit(
		NewPeriodFailureStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
		limit.WithPeriod(seconds),
		limit.WithQuota(quota),
	)

	// 第一次, key不存在
	rv, err := l.GetRunValue(context.Background(), "first")
	assert.Nil(t, err)
	assert.False(t, rv.Exist)
	assert.Zero(t, rv.Count)
	assert.Zero(t, int(rv.TTL))

	runValue, err := l.GetRunValue(context.Background(), "first")
	assert.Nil(t, err)
	assert.Equal(t, runValue.Exist, false)
	assert.Equal(t, runValue.Count, int64(0))
	assert.Equal(t, runValue.TTL, time.Duration(0))

	err = l.SetQuotaFull(context.Background(), "first")
	assert.Nil(t, err)

	// 第二次, key 存在
	rv, err = l.GetRunValue(context.Background(), "first")
	assert.Nil(t, err)
	assert.Equal(t, int64(quota), rv.Count)
	assert.LessOrEqual(t, seconds, rv.TTL)

	runValue, err = l.GetRunValue(context.Background(), "first")
	assert.Nil(t, err)
	assert.Equal(t, runValue.Exist, true)
	assert.Equal(t, runValue.Count, int64(quota))
	assert.Equal(t, runValue.TTL, seconds)

	sts, err := l.CheckErr(context.Background(), "first", internalErr)
	assert.NoError(t, err)
	assert.Equal(t, limit.PeriodFailureLimitStsOverQuota, sts)
	assert.True(t, sts.IsOverQuota())

	err = l.Del(context.Background(), "first")
	assert.Nil(t, err)

	sts, err = l.CheckErr(context.Background(), "first", internalErr)
	assert.NoError(t, err)
	assert.Equal(t, limit.PeriodFailureLimitStsInQuota, sts)
	assert.True(t, sts.IsWithinQuota())

	sts, err = l.CheckErr(context.Background(), "first", nil)
	assert.NoError(t, err)
	assert.Equal(t, limit.PeriodFailureLimitStsSuccess, sts)
	assert.True(t, sts.IsSuccess())
}
