package xratelimit

import (
	"context"
	"github.com/redis/go-redis/v9"
)

var _ Store = (*RedisStore)(nil)

type RedisStore struct {
	incrByLuaScriptSHA string
	cli                *redis.Client
}

func NewRedisStore(cli *redis.Client) *RedisStore {
	r := cli.ScriptLoad(context.TODO(), incrByLuaScript)
	if r.Err() != nil {
		panic("NewRedisStore: " + r.Err().Error())
	}
	return &RedisStore{cli: cli, incrByLuaScriptSHA: r.Val()}
}

var incrByLuaScript = `
local key = KEYS[1]
local unix_time = KEYS[2]
local _LastOpAtKey = 'last_op_at'

local count = tonumber(ARGV[1])
local max = tonumber(ARGV[2])

local qps = redis.call('HINCRBY', key, unix_time, 0)
if count < 1 or max < 1 then
    return -1 -- args is invalid
end
if qps + count > max then
	return 0 -- qps is fulled ! (quick path)
else
    -- incr qps
	local result_qps = redis.call('HINCRBY', key, unix_time, count)
	-- get last_op_at
	local last_op_at = redis.call('HINCRBY', key, _LastOpAtKey, 0)

	-- reset KEY if last_op_at is older than 3 minutes
	if unix_time - last_op_at >= 180 then
	    redis.call('DEL', key)

	    local k1,v1 = unix_time, result_qps
	    local k2,v2 = _LastOpAtKey, unix_time
        redis.call('HMSET', key, k1, v1, k2, v2)
	end

    return result_qps -- success (must be great than 0)
end
`

func (r *RedisStore) AtomicIncrBy(ctx context.Context, key, unixTime string, count, maxQPS int64) (IncrByRespCode, error) {
	res := r.cli.EvalSha(ctx, r.incrByLuaScriptSHA, []string{key, unixTime}, count, maxQPS)
	code, err := res.Int64()
	return IncrByRespCode(code), err
}
