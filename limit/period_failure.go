package limit

import (
	"context"
	"time"
)

const (
	// inner lua code
	// innerPeriodFailureLimitCodeSuccess means success.
	innerPeriodFailureLimitCodeSuccess = 0
	// innerPeriodFailureLimitCodeInQuota means within the quota.
	innerPeriodFailureLimitCodeInQuota = 1
	// innerPeriodFailureLimitCodeOverQuota means passed the quota.
	innerPeriodFailureLimitCodeOverQuota = 2
)

// A PeriodFailureLimit is used to limit requests when failure during a period of time.
type PeriodFailureLimit[B PeriodFailureBackend] struct {
	// a period seconds of time
	period int
	// limit quota requests during a period seconds of time.
	quota int
	// keyPrefix in redis
	keyPrefix string
	store     B
	isAlign   bool
}

// NewPeriodFailureLimit returns a PeriodFailureLimit with given parameters.
func NewPeriodFailureLimit[B PeriodFailureBackend](store B, opts ...PeriodLimitOption) *PeriodFailureLimit[B] {
	limiter := &PeriodFailureLimit[B]{
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

func (p *PeriodFailureLimit[B]) align()                { p.isAlign = true }
func (p *PeriodFailureLimit[B]) setKeyPrefix(k string) { p.keyPrefix = k }
func (p *PeriodFailureLimit[B]) setPeriod(v time.Duration) {
	if vv := int(v / time.Second); vv > 0 {
		p.period = int(v / time.Second)
	}
}
func (p *PeriodFailureLimit[B]) setQuota(v int) { p.quota = v }

// CheckErr requests a permit state.
// same as Check
func (p *PeriodFailureLimit[B]) CheckErr(ctx context.Context, key string, err error) (PeriodFailureLimitState, error) {
	return p.Check(ctx, key, err == nil)
}

// Check requests a permit.
func (p *PeriodFailureLimit[B]) Check(ctx context.Context, key string, success bool) (PeriodFailureLimitState, error) {
	code, err := p.store.Check(ctx,
		p.formatKey(key),
		p.quota,
		p.calcExpireSeconds(),
		success,
	)
	if err != nil {
		return PeriodFailureLimitStsUnknown, err
	}
	switch code {
	case innerPeriodFailureLimitCodeSuccess:
		return PeriodFailureLimitStsSuccess, nil
	case innerPeriodFailureLimitCodeInQuota:
		return PeriodFailureLimitStsInQuota, nil
	case innerPeriodFailureLimitCodeOverQuota:
		return PeriodFailureLimitStsOverQuota, nil
	default:
		return PeriodFailureLimitStsUnknown, ErrUnknownCode
	}
}

// SetQuotaFull set a permit over quota.
func (p *PeriodFailureLimit[B]) SetQuotaFull(ctx context.Context, key string) error {
	return p.store.SetQuotaFull(ctx,
		p.formatKey(key),
		p.quota,
		p.calcExpireSeconds(),
	)
}

// Del delete a permit
func (p *PeriodFailureLimit[B]) Del(ctx context.Context, key string) error {
	return p.store.Del(ctx, p.formatKey(key))
}

// GetRunValue get run value
// Exist: false if key not exist.
// Count: current failure count
// TTL: not set expire time, t = -1
func (p *PeriodFailureLimit[B]) GetRunValue(ctx context.Context, key string) (*RunValue, error) {
	tb, err := p.store.GetRunValue(ctx, p.formatKey(key))
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

func (p *PeriodFailureLimit[B]) formatKey(key string) string {
	return p.keyPrefix + key
}

func (p *PeriodFailureLimit[B]) calcExpireSeconds() int {
	if p.isAlign {
		now := time.Now()
		_, offset := now.Zone()
		unix := now.Unix() + int64(offset)
		return p.period - int(unix%int64(p.period))
	}
	return p.period
}
