package v8

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/things-go/limiter/limit"
	redis2 "github.com/things-go/limiter/limit/redis"
)

// A PeriodFailureLimit is used to limit requests when failure during a period of time.
type PeriodFailureLimit struct {
	// a period seconds of time
	period int
	// limit quota requests during a period seconds of time.
	quota int
	// keyPrefix in redis
	keyPrefix string
	store     *redis.Client
	isAlign   bool
}

// NewPeriodFailureLimit returns a PeriodFailureLimit with given parameters.
func NewPeriodFailureLimit(store *redis.Client, opts ...PeriodLimitOption) *PeriodFailureLimit {
	limiter := &PeriodFailureLimit{
		period:    int(24 * time.Hour / time.Second),
		quota:     6,
		keyPrefix: "limit:period:failure:", // limit:period:failure:
		store:     store,
	}
	for _, opt := range opts {
		opt(limiter)
	}
	return limiter
}

func (p *PeriodFailureLimit) align()                { p.isAlign = true }
func (p *PeriodFailureLimit) setKeyPrefix(k string) { p.keyPrefix = k }
func (p *PeriodFailureLimit) setPeriod(v time.Duration) {
	if vv := int(v / time.Second); vv > 0 {
		p.period = int(v / time.Second)
	}
}
func (p *PeriodFailureLimit) setQuota(v int) { p.quota = v }

// CheckErr requests a permit state.
// same as Check
func (p *PeriodFailureLimit) CheckErr(ctx context.Context, key string, err error) (limit.PeriodFailureLimitState, error) {
	return p.Check(ctx, key, err == nil)
}

// Check requests a permit.
func (p *PeriodFailureLimit) Check(ctx context.Context, key string, success bool) (limit.PeriodFailureLimitState, error) {
	s := "0"
	if success {
		s = "1"
	}
	result, err := p.store.Eval(ctx,
		redis2.PeriodFailureLimitFixedScript,
		[]string{p.formatKey(key)},
		[]string{
			strconv.Itoa(p.quota),
			strconv.Itoa(p.calcExpireSeconds()),
			s,
		},
	).Result()
	if err != nil {
		return limit.PeriodFailureLimitStsUnknown, err
	}
	code, ok := result.(int64)
	if !ok {
		return limit.PeriodFailureLimitStsUnknown, limit.ErrUnknownCode
	}
	switch code {
	case redis2.InnerPeriodFailureLimitCodeSuccess:
		return limit.PeriodFailureLimitStsSuccess, nil
	case redis2.InnerPeriodFailureLimitCodeInQuota:
		return limit.PeriodFailureLimitStsInQuota, nil
	case redis2.InnerPeriodFailureLimitCodeOverQuota:
		return limit.PeriodFailureLimitStsOverQuota, nil
	default:
		return limit.PeriodFailureLimitStsUnknown, limit.ErrUnknownCode
	}
}

// SetQuotaFull set a permit over quota.
func (p *PeriodFailureLimit) SetQuotaFull(ctx context.Context, key string) error {
	err := p.store.Eval(ctx,
		redis2.PeriodFailureLimitFixedSetQuotaFullScript,
		[]string{p.formatKey(key)},
		[]string{
			strconv.Itoa(p.quota),
			strconv.Itoa(p.calcExpireSeconds()),
		},
	).Err()
	if err == redis.Nil {
		return nil
	}
	return err
}

// Del delete a permit
func (p *PeriodFailureLimit) Del(ctx context.Context, key string) error {
	return p.store.Del(ctx, p.formatKey(key)).Err()
}

// TTL get key ttl
// if key not exist, time = -1.
// if key exist, but not set expire time, t = -2
func (p *PeriodFailureLimit) TTL(ctx context.Context, key string) (time.Duration, error) {
	return p.store.TTL(ctx, p.formatKey(key)).Result()
}

// GetInt get current failure count
func (p *PeriodFailureLimit) GetInt(ctx context.Context, key string) (int, bool, error) {
	v, err := p.store.Get(ctx, p.formatKey(key)).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, false, nil
		}
		return 0, false, err
	}
	return v, true, nil
}

func (p *PeriodFailureLimit) formatKey(key string) string {
	return p.keyPrefix + key
}

func (p *PeriodFailureLimit) calcExpireSeconds() int {
	if p.isAlign {
		now := time.Now()
		_, offset := now.Zone()
		unix := now.Unix() + int64(offset)
		return p.period - int(unix%int64(p.period))
	}
	return p.period
}
