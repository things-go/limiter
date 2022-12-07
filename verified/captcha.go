package verified

import (
	"context"
	"errors"
	"time"
)

// QuestionAnswer question and answer for CaptchaDriver driver
type QuestionAnswer struct {
	Id       string
	Question string
	Answer   string
}

// CaptchaDriver the captcha driver
type CaptchaDriver interface {
	Name() string
	GenerateQuestionAnswer() (*QuestionAnswer, error)
}

// CaptchaProvider the captcha provider
type CaptchaProvider interface {
	AcquireDriver(kind string) CaptchaDriver
}

// Captcha verified captcha limit
type Captcha struct {
	p              CaptchaProvider // CaptchaProvider generate captcha provider
	store          Storage         // store client
	disableOneTime bool            // 禁用一次性验证
	keyPrefix      string          // store 存验证码key的前缀, 默认 verified:captcha:
	keyExpires     time.Duration   // store 存验证码key的过期时间, 默认: 3 分种
	maxErrQuota    int             // store 验证码验证最大错误次数限制, 默认: 1
}

// NewVerifiedCaptcha
func NewVerifiedCaptcha(p CaptchaProvider, store Storage, opts ...Option) *Captcha {
	v := &Captcha{
		p,
		store,
		false,
		"verified:captcha:",
		time.Minute * 3,
		1,
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// Name the provider name
func (v *Captcha) Name(kind string) string { return v.p.AcquireDriver(kind).Name() }

// Generate generate id, question.
func (v *Captcha) Generate(ctx context.Context, kind string, opts ...GenerateOption) (id, question string, err error) {
	genOpt := generateOption{
		keyExpires:  v.keyExpires,
		maxErrQuota: v.maxErrQuota,
	}
	for _, f := range opts {
		f(&genOpt)
	}

	q, err := v.p.AcquireDriver(kind).GenerateQuestionAnswer()
	if err != nil {
		return "", "", err
	}
	err = v.store.Store(ctx, &StoreArgs{
		v.disableOneTime,
		v.keyPrefix + kind + ":" + q.Id,
		genOpt.keyExpires,
		genOpt.maxErrQuota,
		q.Answer,
	})
	if err != nil {
		return "", "", err
	}
	return q.Id, q.Question, nil
}

// Verify the answer.
func (v *Captcha) Verify(ctx context.Context, kind, id, answer string) bool {
	return v.store.Verify(ctx,
		&VerifyArgs{
			v.disableOneTime,
			v.keyPrefix + kind + ":" + id,
			answer,
		},
	)
}

func (v *Captcha) setKeyPrefix(k string)         { v.keyPrefix = k }
func (v *Captcha) setKeyExpires(t time.Duration) { v.keyExpires = t }
func (v *Captcha) setMaxErrQuota(quota int) {
	v.disableOneTime = true
	v.maxErrQuota = quota
}

type UnsupportedVerifiedCaptchaDriver struct{}

func (x UnsupportedVerifiedCaptchaDriver) Name() string {
	return "Unsupported verified captcha driver"
}
func (x UnsupportedVerifiedCaptchaDriver) GenerateQuestionAnswer() (*QuestionAnswer, error) {
	return nil, errors.New(x.Name())
}
