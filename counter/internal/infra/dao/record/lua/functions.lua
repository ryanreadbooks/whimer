#!lua name=libcounter

local function is_pcall_err(res)
  if type(res) == 'table' and res.err ~= nil then
    return true, res.err
  end
  return false, nil
end

local function counter_sizelimit_batchadd(keys, args)
  local key = keys[1]
  local arg_max_limit = args[1]
  local arg_evit_number = args[2]
  -- starting from index 3 is zadd member args
  local arg_members = { select(3, unpack(args)) }

  local max_limit = tonumber(arg_max_limit)
  if max_limit == nil then
    return redis.error_reply('ARGV[1] max_limit is not a number')
  end
  local evit_number = tonumber(arg_evit_number)
  if evit_number == nil then
    return redis.error_reply('ARGV[2] evit_number is not a number')
  end

  -- step1. zcard check max_limit
  local r1 = redis.pcall('ZCARD', key)
  local yes, err = is_pcall_err(r1)
  if yes then
    return redis.error_reply('zcard failed: ' .. key .. ', err: ' .. err)
  end

  -- step2. check if we need to evit some members with smallest scores
  if r1 > max_limit then
    -- pop min from key to spare space
    local pop_result = redis.pcall('ZPOPMIN', key, evit_number)
    yes, err = is_pcall_err(pop_result)
    if yes then
      -- do not abort even if error occurs
      redis.log(redis.LOG_WARNING, 'zpopmin failed in sizelimit_batch_add: ' .. err)
    end
  end

  -- step3. zdd members
  return redis.pcall('ZADD', key, unpack(arg_members))
end

-- register redis functions
redis.register_function('counter_sizelimit_batchadd', counter_sizelimit_batchadd)
