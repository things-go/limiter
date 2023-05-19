package limit

import (
	"context"
	"time"
)

const (
	// inner lua code
	innerPeriodLimitAllowed   = 0
	innerPeriodLimitHitQuota  = 1
	innerPeriodLimitOverQuota = 2
)

// A PeriodLimit is used to limit requests during a period of time.
type PeriodLimit[B PeriodBackend] struct {
	// keyPrefix in redis
	keyPrefix string
	// a period seconds of time
	period int
	// limit quota requests during a period seconds of time.
	quota   int
	isAlign bool
	store   B
}

// NewPeriodLimit returns a PeriodLimit with given parameters.
func NewPeriodLimit[B PeriodBackend](store B, opts ...PeriodLimitOption) *PeriodLimit[B] {
	limiter := &PeriodLimit[B]{
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

func (p *PeriodLimit[B]) align()                { p.isAlign = true }
func (p *PeriodLimit[B]) setKeyPrefix(k string) { p.keyPrefix = k }
func (p *PeriodLimit[B]) setPeriod(v time.Duration) {
	if vv := int(v / time.Second); vv > 0 {
		p.period = int(v / time.Second)
	}
}
func (p *PeriodLimit[B]) setQuota(v int) { p.quota = v }

// Take requests a permit with context, it returns the permit state.
func (p *PeriodLimit[B]) Take(ctx context.Context, key string) (PeriodLimitState, error) {
	code, err := p.store.Take(
		ctx,
		p.formatKey(key),
		p.quota,
		p.calcExpireSeconds(),
	)
	if err != nil {
		return PeriodLimitStsUnknown, err
	}
	switch code {
	case innerPeriodLimitAllowed:
		return PeriodLimitStsAllowed, nil
	case innerPeriodLimitHitQuota:
		return PeriodLimitStsHitQuota, nil
	case innerPeriodLimitOverQuota:
		return PeriodLimitStsOverQuota, nil
	default:
		return PeriodLimitStsUnknown, ErrUnknownCode
	}
}

// SetQuotaFull set a permit over quota.
func (p *PeriodLimit[B]) SetQuotaFull(ctx context.Context, key string) error {
	return p.store.SetQuotaFull(ctx,
		p.formatKey(key),
		p.quota,
		p.calcExpireSeconds(),
	)
}

// Del delete a permit
func (p *PeriodLimit[B]) Del(ctx context.Context, key string) error {
	return p.store.Del(ctx, p.formatKey(key))
}

// GetRunValue get run value
// Exist: false if key not exist.
// Count: current count
// TTL: not set expire time, t = -1
func (p *PeriodLimit[B]) GetRunValue(ctx context.Context, key string) (*RunValue, error) {
	tb, err := p.store.GetRunValue(
		ctx,
		p.formatKey(key),
	)
	if err != nil {
		return nil, err
	}
	switch {
	case len(tb) == 1 && tb[0] == 0:
		return &RunValue{
			Exist: false,
			Count: 0,
			TTL:   0,
		}, nil
	case len(tb) == 3:
		var t time.Duration

		switch n := tb[2]; n {
		// -2 if the key does not exist
		// -1 if the key exists but has no associated expire
		case -2, -1:
			t = time.Duration(n)
		default:
			t = time.Duration(n) * time.Second
		}
		return &RunValue{
			Exist: tb[0] == 1 && t != -2,
			Count: tb[1],
			TTL:   t,
		}, nil
	default:
		return nil, ErrUnknownCode
	}
}

func (p *PeriodLimit[B]) formatKey(key string) string {
	return p.keyPrefix + key
}

func (p *PeriodLimit[B]) calcExpireSeconds() int {
	if p.isAlign {
		now := time.Now()
		_, offset := now.Zone()
		unix := now.Unix() + int64(offset)
		return p.period - int(unix%int64(p.period))
	}
	return p.period
}
