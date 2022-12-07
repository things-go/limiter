package v8

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/things-go/limiter/limit"
	redisScript "github.com/things-go/limiter/limit/redis"
)

// A PeriodLimit is used to limit requests during a period of time.
type PeriodLimit struct {
	// keyPrefix in redis
	keyPrefix string
	// a period seconds of time
	period int
	// limit quota requests during a period seconds of time.
	quota   int
	isAlign bool
	store   *redis.Client
}

// NewPeriodLimit returns a PeriodLimit with given parameters.
func NewPeriodLimit(store *redis.Client, opts ...PeriodLimitOption) *PeriodLimit {
	limiter := &PeriodLimit{
		keyPrefix: "limit:period:",
		period:    int(24 * time.Hour / time.Second),
		quota:     6,
		isAlign:   false,
		store:     store,
	}
	for _, opt := range opts {
		opt(limiter)
	}
	return limiter
}

func (p *PeriodLimit) align()                { p.isAlign = true }
func (p *PeriodLimit) setKeyPrefix(k string) { p.keyPrefix = k }
func (p *PeriodLimit) setPeriod(v time.Duration) {
	if vv := int(v / time.Second); vv > 0 {
		p.period = int(v / time.Second)
	}
}
func (p *PeriodLimit) setQuota(v int) { p.quota = v }

// Take requests a permit with context, it returns the permit state.
func (p *PeriodLimit) Take(ctx context.Context, key string) (limit.PeriodLimitState, error) {
	result, err := p.store.Eval(ctx,
		redisScript.PeriodLimitScript,
		[]string{
			p.formatKey(key),
		},
		[]string{
			strconv.Itoa(p.quota),
			strconv.Itoa(p.calcExpireSeconds()),
		},
	).Result()
	if err != nil {
		return limit.PeriodLimitStsUnknown, err
	}

	code, ok := result.(int64)
	if !ok {
		return limit.PeriodLimitStsUnknown, limit.ErrUnknownCode
	}
	switch code {
	case redisScript.InnerPeriodLimitAllowed:
		return limit.PeriodLimitStsAllowed, nil
	case redisScript.InnerPeriodLimitHitQuota:
		return limit.PeriodLimitStsHitQuota, nil
	case redisScript.InnerPeriodLimitOverQuota:
		return limit.PeriodLimitStsOverQuota, nil
	default:
		return limit.PeriodLimitStsUnknown, limit.ErrUnknownCode
	}
}

// SetQuotaFull set a permit over quota.
func (p *PeriodLimit) SetQuotaFull(ctx context.Context, key string) error {
	return p.store.Eval(ctx,
		redisScript.PeriodLimitSetQuotaFullScript,
		[]string{
			p.formatKey(key),
		},
		[]string{
			strconv.Itoa(p.quota),
			strconv.Itoa(p.calcExpireSeconds()),
		},
	).Err()
}

// Del delete a permit
func (p *PeriodLimit) Del(ctx context.Context, key string) error {
	return p.store.Del(ctx, p.formatKey(key)).Err()
}

// TTL get key ttl
// if key not exist, time = -1.
// if key exist, but not set expire time, t = -2
func (p *PeriodLimit) TTL(ctx context.Context, key string) (time.Duration, error) {
	return p.store.TTL(ctx, p.formatKey(key)).Result()
}

// GetInt get current count
func (p *PeriodLimit) GetInt(ctx context.Context, key string) (int, bool, error) {
	v, err := p.store.Get(ctx, p.formatKey(key)).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, false, nil
		}
		return 0, false, err
	}
	return v, true, nil
}

func (p *PeriodLimit) formatKey(key string) string {
	return p.keyPrefix + key
}

func (p *PeriodLimit) calcExpireSeconds() int {
	if p.isAlign {
		now := time.Now()
		_, offset := now.Zone()
		unix := now.Unix() + int64(offset)
		return p.period - int(unix%int64(p.period))
	}
	return p.period
}
