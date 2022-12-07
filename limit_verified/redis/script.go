package redis

type LimitVerifiedState int

const (
	// inner lua send/verify code statue value
	InnerLimitVerifiedSuccess = 0
	// inner lua send code value
	InnerLimitVerifiedOfSendCodeReachMaxSendPerDay  = 1
	InnerLimitVerifiedOfSendCodeResendTooFrequently = 2
	// inner lua verify code value
	InnerLimitVerifiedOfVerifyCodeRequiredOrExpired   = 1
	InnerLimitVerifiedOfVerifyCodeReachMaxError       = 2
	InnerLimitVerifiedOfVerifyCodeVerificationFailure = 3
)

const (
	LimitVerifiedSendCodeScript = `
local keyPrefix = KEYS[1] -- keyPrefix
local kind = KEYS[2] -- kind 
local target = KEYS[3] -- target
local code = ARGV[1] -- 验证码
local maxSendPerDay = tonumber(ARGV[2]) -- 限制一天最大发送次数
local codeMaxSendPerDay = tonumber(ARGV[3]) -- 限制一天最大发送次数
local codeMaxErrorQuota = tonumber(ARGV[4]) -- 验证码最大验证失败次数
local codeAvailWindowSecond = tonumber(ARGV[5]) -- 验证码有效窗口时间, 单位: 秒
local codeResendIntervalSecond = tonumber(ARGV[6]) -- 验证码重发间隔时间
local now = tonumber(ARGV[7]) -- 当前时间, 单位秒
local expires = tonumber(ARGV[8]) -- global key 过期时间, 单位: 秒

local globalKey = keyPrefix .. target
local codeKey = keyPrefix .. target .. ":_entry_:{" .. kind .. "}"

local sendCnt = redis.call("HINCRBY", globalKey, "sendCnt", 1)
local codeCnt = redis.call("HINCRBY", globalKey, "codeCnt", 1)
if sendCnt == 1 then
    redis.call("EXPIRE", globalKey, expires)
end
if sendCnt > maxSendPerDay or codeCnt > codeMaxSendPerDay then
	redis.call("HINCRBY", globalKey, "sendCnt", -1)
	redis.call("HINCRBY", globalKey, "codeCnt", -1)
    return 1 -- 超过每天发送限制次数
end

if (redis.call("EXISTS", codeKey) == 1) then
    local lastedAt = tonumber(redis.call("HGET", codeKey, "lasted"))
    if lastedAt + codeResendIntervalSecond > now then
		redis.call("HINCRBY", globalKey, "sendCnt", -1)
		redis.call("HINCRBY", globalKey, "codeCnt", -1)
        return 2 -- 发送过于频繁, 即还在重发限制窗口
    end
end

redis.call("HMSET", codeKey, "code", code, "quota", codeMaxErrorQuota, "err", 0, "lasted", now)
redis.call("EXPIRE", codeKey, codeAvailWindowSecond)

return 0 -- 成功
`
	LimitVerifiedRollbackSendCntAndCodeCntScript = `
local keyPrefix = KEYS[1] -- keyPrefix
local kind = KEYS[2] -- kind 
local target = KEYS[3] -- target
local code = ARGV[1] -- 验证码
local now = tonumber(ARGV[2]) -- 当前时间, 单位秒

local globalKey = keyPrefix .. target
local codeKey = keyPrefix .. target .. ":_entry_:{" .. kind .. "}"

redis.call("HINCRBY", globalKey, "sendCnt", -1)
redis.call("HINCRBY", globalKey, "codeCnt", -1)
if (redis.call("EXISTS", codeKey) == 1) then
	local currentCode = redis.call("HGET", codeKey, "code")
	local lastedAt = tonumber(redis.call("HGET", codeKey, "lasted"))
	if currentCode == code and lastedAt == now then 
        redis.call("DEL", codeKey) -- 的确是你发送的, 删除 code key
	end
end
`
	LimitVerifiedVerifyCodeScript = `
local keyPrefix = KEYS[1] -- keyPrefix
local kind = KEYS[2] -- kind 
local target = KEYS[3] -- target
local code = ARGV[1] -- 验证码
local now = tonumber(ARGV[2]) -- 当前时间, 单位秒

local globalKey = keyPrefix .. target
local codeKey = keyPrefix .. target .. ":_entry_:{" .. kind .. "}" 

if redis.call("EXISTS", codeKey) == 0 then
    return 1  -- 未发送短信验证码 或 验证码已过期
end

local errCnt = tonumber(redis.call('HGET', codeKey, "err"))
local codeMaxErrorQuota = tonumber(redis.call('HGET', codeKey, "quota"))
local currentCode = redis.call('HGET', codeKey, "code")
if errCnt >= codeMaxErrorQuota then
    return 2  -- 验证码错误次数超过限制
end
if currentCode == code then
    redis.call("DEL", codeKey) -- 删除 code key
    return 0 -- 成功
else
    redis.call('HINCRBY', codeKey, "err", 1)
    return 3 -- 验证码错误
end
`
	LimitVerifiedIncrSendCntScript = `
local keyPrefix = KEYS[1] -- keyPrefix
local target = KEYS[2] -- target
local maxSendPerDay = tonumber(ARGV[1]) -- 限制一天最大发送次数
local expires = tonumber(ARGV[2]) -- global key 过期时间, 单位: 秒

local globalKey = keyPrefix .. target

local sendCnt = redis.call("HINCRBY", globalKey, "sendCnt", 1)
if sendCnt == 1 then
    redis.call("EXPIRE", globalKey, expires)
end
if sendCnt > maxSendPerDay then
	redis.call("HINCRBY", globalKey, "sendCnt", -1)
    return 1 -- 超过每天发送限制次数
end
return 0 -- 成功
`
	LimitVerifiedDecrSendCntScript = `
local keyPrefix = KEYS[1] -- keyPrefix
local target = KEYS[2] -- target

local globalKey = keyPrefix .. target

local sendCnt = redis.call("HINCRBY", globalKey, "sendCnt", -1)
if sendCnt < 0 then
    redis.call("DEL", globalKey)
end
return 0
`
)
