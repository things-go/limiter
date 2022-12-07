package v8

import (
	"math/bits"
	"math/rand"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/things-go/limiter/verified"
)

const defaultKind = "default"
const unsupportedKind = "unsupported"

const question = "1+1"
const answer = "2"
const badAnswer = "3"

var defaultAlphabet = []byte("QWERTYUIOPLKJHGFDSAZXCVBNMabcdefghijklmnopqrstuvwxyz")

func randString(length int) string {
	b := make([]byte, length)
	bn := bits.Len(uint(len(defaultAlphabet)))
	mask := int64(1)<<bn - 1
	max := 63 / bn
	r := rand.New(rand.NewSource(time.Now().UnixNano() + rand.Int63() + rand.Int63()))

	// A rand.Int63() generates 63 random bits, enough for alphabets letters!
	for i, cache, remain := 0, r.Int63(), max; i < length; {
		if remain == 0 {
			cache, remain = r.Int63(), max
		}
		if idx := int(cache & mask); idx < len(defaultAlphabet) {
			b[i] = defaultAlphabet[idx]
			i++
		}
		cache >>= bn
		remain--
	}
	return string(b)
}

var _ verified.CaptchaProvider = (*TestVerifiedCaptchaProvider)(nil)

type TestVerifiedCaptchaProvider struct{}

func (t TestVerifiedCaptchaProvider) AcquireDriver(kind string) verified.CaptchaDriver {
	if kind == unsupportedKind {
		return new(verified.UnsupportedVerifiedCaptchaDriver)
	}
	return new(TestVerifiedCaptchaDriver)
}

type TestVerifiedCaptchaDriver struct{}

func (t TestVerifiedCaptchaDriver) Name() string { return "test_provider" }

func (t TestVerifiedCaptchaDriver) GenerateQuestionAnswer() (*verified.QuestionAnswer, error) {
	return &verified.QuestionAnswer{
		Id:       randString(6),
		Question: question,
		Answer:   answer,
	}, nil
}

func TestCaptcha_Improve_Cover(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()
	l := verified.NewVerifiedCaptcha(
		new(TestVerifiedCaptchaProvider),
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
	l.Name(defaultKind)
}

func TestCaptcha_Unsupported_Driver(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)

	l := verified.NewVerifiedCaptcha(
		new(TestVerifiedCaptchaProvider),
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
	mr.Close()

	_, _, err = l.Generate(unsupportedKind)
	assert.Error(t, err)
}

func TestCaptcha_RedisUnavailable(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)

	l := verified.NewVerifiedCaptcha(
		new(TestVerifiedCaptchaProvider),
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
	mr.Close()

	_, _, err = l.Generate(defaultKind)
	assert.Error(t, err)
}

func TestCaptcha_OneTime(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	l := verified.NewVerifiedCaptcha(
		new(TestVerifiedCaptchaProvider),
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
		verified.WithKeyPrefix("verified:captcha:"),
		verified.WithKeyExpires(time.Minute*3),
	)

	id, _, err := l.Generate(defaultKind, verified.WithGenerateKeyExpires(time.Minute*5))
	assert.NoError(t, err)

	b := l.Verify(defaultKind, id, answer)
	require.True(t, b)

	b = l.Verify(defaultKind, id, answer)
	require.False(t, b)
}

func TestCaptcha_In_Quota(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	l := verified.NewVerifiedCaptcha(
		new(TestVerifiedCaptchaProvider),
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
		verified.WithKeyPrefix("verified:captcha:"),
		verified.WithKeyExpires(time.Minute*3),
		verified.WithMaxErrQuota(3),
	)

	id, _, err := l.Generate(defaultKind,
		verified.WithGenerateKeyExpires(time.Minute*5),
	)
	assert.NoError(t, err)

	b := l.Verify(defaultKind, id, badAnswer)
	require.False(t, b)
	b = l.Verify(defaultKind, id, badAnswer)
	require.False(t, b)
	b = l.Verify(defaultKind, id, answer)
	require.True(t, b)
}

func TestCaptcha_Over_Quota(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	l := verified.NewVerifiedCaptcha(
		new(TestVerifiedCaptchaProvider),
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
		verified.WithKeyPrefix("verified:captcha:"),
		verified.WithKeyExpires(time.Minute*3),
		verified.WithMaxErrQuota(3),
	)

	id, _, err := l.Generate(defaultKind,
		verified.WithGenerateKeyExpires(time.Minute*5),
		verified.WithGenerateMaxErrQuota(6),
	)
	assert.NoError(t, err)

	for i := 0; i < 6; i++ {
		b := l.Verify(defaultKind, id, badAnswer)
		require.False(t, b)
	}
	b := l.Verify(defaultKind, id, answer)
	require.False(t, b)
}

// TODO: success in redis, but failed in miniredis
func TestCaptcha_Onetime_Timeout(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	l := verified.NewVerifiedCaptcha(
		new(TestVerifiedCaptchaProvider),
		// NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
		NewRedisStore(redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "123456", DB: 0})),
	)
	id, _, err := l.Generate(defaultKind, verified.WithGenerateKeyExpires(time.Second*1))
	assert.NoError(t, err)

	time.Sleep(time.Second * 2)

	b := l.Verify(defaultKind, id, "2")
	require.False(t, b)
}
