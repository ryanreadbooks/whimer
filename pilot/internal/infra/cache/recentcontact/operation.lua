#!lua name=libpilot_recentcontact

local function is_pcall_err(res)
  if type(res) == 'table' and res.err ~= nil then
    return true, res.err
  end
  return false, nil
end

local function make_redis_err(text, key, e)
  return redis.error_reply(text .. ': ' .. key .. ', err: ' .. e)
end

local function recent_contact_cleanup(keys, args)
  local key = keys[1]
  local threshold = tonumber(args[1])
  local start = tonumber(args[2])
  local stop = tonumber(args[3])

  local r1 = redis.pcall('ZCARD', key)
  local is_err, err = is_pcall_err(r1)
  if is_err then
    return make_redis_err('zcard failed', key, err)
  end

  if threshold <= r1 then
    -- clean
    return redis.pcall('ZREMRANGEBYSCORE', key, start, stop)
  end

  return 0
end

local function recent_contact_append(keys, args)
  local key = keys[1]
  local threshold = tonumber(args[1])
  local expireSec = tonumber(args[2])
  -- starting from index 3 is zadd member args
  local zadd_args = { select(3, unpack(args)) }

  if threshold == nil then
    threshold = 50
  end
  if expireSec == nil then
    expireSec = 604800
  end

  if #zadd_args % 2 ~= 0 then
    return make_redis_err('invalid count for zadd', key, 'got ' .. #zadd_args)
  end

  -- check if we need to pop to main threshold count in key
  local r1 = redis.pcall('ZCARD', key)
  local is_err, err = is_pcall_err(r1)
  if is_err then
    return make_redis_err('zcard failed', key, err)
  end

  local new_member_cnt = #zadd_args / 2
  local new_total_cnt = r1 + new_member_cnt
  local pop_count = new_total_cnt - threshold

  if pop_count > 0 then
    -- need popping from key
    local r2 = redis.pcall('ZPOPMIN', key, pop_count)
    is_err, err = is_pcall_err(r2)
    if is_err then
      return make_redis_err('zpopmin failed', key, err)
    end
  end

  -- now we can add to sorted set
  local r3 = redis.pcall('ZADD', key, unpack(zadd_args))
  -- set expire for key
  redis.pcall('EXPIRE', key, expireSec)

  return r3
end

-- register redis functions
redis.register_function('recent_contact_cleanup', recent_contact_cleanup)
redis.register_function('recent_contact_append', recent_contact_append)
