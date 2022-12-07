package redis

const (
	PeriodLimitScript = `
local key = KEYS[1]
local quota = tonumber(ARGV[1])
local window = tonumber(ARGV[2])

local current = redis.call("INCRBY", key, 1)
if current == 1 then
    redis.call("EXPIRE", key, window)
end
if current < quota then
    return 0 -- allow
elseif current == quota then
    return 1 -- hit quota
else
    return 2 -- over quata
end
`
	PeriodLimitSetQuotaFullScript = `
local key = KEYS[1]
local quota = tonumber(ARGV[1])
local window = tonumber(ARGV[2])

local current = tonumber(redis.call("GET", key))
if current == nil then 
	redis.call("SETEX", key, window, quota)
elseif current < quota then 
	redis.call("SET", key, quota, "KEEPTTL")
end
return 0
`
)

const (
	// inner lua code
	InnerPeriodLimitAllowed   = 0
	InnerPeriodLimitHitQuota  = 1
	InnerPeriodLimitOverQuota = 2
)
