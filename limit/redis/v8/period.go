package v8

import (
	"context"
	"strconv"

	"github.com/go-redis/redis/v8"

	"github.com/things-go/limiter/limit"
	redisScript "github.com/things-go/limiter/limit/redis"
)

var _ limit.PeriodBackend = (*PeriodBackend)(nil)

// A PeriodBackend is used to limit requests during a period of time.
type PeriodBackend struct {
	store *redis.Client
}

// NewPeriodBackend returns a PeriodLimit with given parameters.
func NewPeriodBackend(store *redis.Client) *PeriodBackend {
	return &PeriodBackend{
		store: store,
	}
}

// Take requests a permit with context, it returns the permit state.
func (p *PeriodBackend) Take(ctx context.Context, key string, quota, expireSec int) (int64, error) {
	return p.store.Eval(ctx,
		redisScript.PeriodLimitScript,
		[]string{
			key,
		},
		[]string{
			strconv.Itoa(quota),
			strconv.Itoa(expireSec),
		},
	).Int64()
}

// SetQuotaFull set a permit over quota.
func (p *PeriodBackend) SetQuotaFull(ctx context.Context, key string, quota, expireSec int) error {
	return p.store.Eval(ctx,
		redisScript.PeriodLimitSetQuotaFullScript,
		[]string{
			key,
		},
		[]string{
			strconv.Itoa(quota),
			strconv.Itoa(expireSec),
		},
	).Err()
}

// Del delete a permit
func (p *PeriodBackend) Del(ctx context.Context, key string) error {
	return p.store.Del(ctx, key).Err()
}

// GetRunValue get run value
// Exist: false if key not exist.
// Count: current count
// TTL: not set expire time, t = -1
func (p *PeriodBackend) GetRunValue(ctx context.Context, key string) ([]int64, error) {
	return p.store.Eval(ctx,
		redisScript.PeriodLimitRunValueScript,
		[]string{
			key,
		},
	).Int64Slice()

}
