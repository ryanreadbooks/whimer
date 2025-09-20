#!lua name=librelation_dao

local function is_pcall_err(res)
  if type(res) == 'table' and res.err ~= nil then
    return true, res.err
  end
  return false, nil
end

local function make_redis_err(text, key, e)
  return redis.error_reply(text .. ': ' .. key .. ', err: ' .. e)
end

-- follow user
local function relation_do_follow(keys, args)
  local link_key = keys[1]
  local following_zset_key = keys[2]
  local fan_zset_key = keys[3]

  local link_value = args[1]
  local link_value_expire_sec = args[2]
  local uid = args[3]
  local followee = args[4]
  local follow_time = tonumber(args[5])
  local max_fan_zset_count = tonumber(args[6])

  -- 1. set link cache
  local r1 = redis.pcall('SET', link_key, link_value, 'EX', link_value_expire_sec)
  local is_err, err = is_pcall_err(r1)
  if is_err then
    return make_redis_err('set failed', link_key, err)
  end

  -- 2. set following zset
  local r2 = redis.pcall('ZADD', following_zset_key, follow_time, followee)
  is_err, err = is_pcall_err(r2)
  if is_err then
    return make_redis_err('zadd failed', following_zset_key, err)
  end

  -- 3. set fan zset
  local r3 = redis.pcall('ZCARD', fan_zset_key)
  is_err, err = is_pcall_err(r3)
  if is_err then
    return make_redis_err('zcard failed', fan_zset_key, err)
  end

  -- step3.1. check if we need to evit some members with smallest scores
  if r3 > max_fan_zset_count then
    -- pop min from key to spare space
    local r4 = redis.pcall('ZPOPMIN', fan_zset_key, evit_number)
    is_err, err = is_pcall_err(r4)
    if is_err then
      return make_redis_err('zpopmin failed', fan_zset_key, err)
    end
  end

  -- 3.2. zadd members
  return redis.pcall('ZADD', fan_zset_key, follow_time, uid)
end

-- unfollow user
local function relation_do_unfollow(keys, args)
  local link_key = keys[1]
  local following_zset_key = keys[2]
  local fan_zset_key = keys[3]

  local uid = args[1]
  local followee = args[2]

  -- 1. del link cache?
  redis.pcall('DEL', link_key)

  -- 2. zrem following zset
  redis.pcall('ZREM', following_zset_key, followee)

  -- 3. zrem fan zset
  redis.pcall('ZREM', fan_zset_key, uid)
end

-- register redis functions
redis.register_function('relation_do_follow', relation_do_follow)
redis.register_function('relation_do_unfollow', relation_do_unfollow)
