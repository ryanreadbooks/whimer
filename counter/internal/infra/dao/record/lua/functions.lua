#!lua name=libcounter

local function is_pcall_err(res)
  if type(res) == 'table' and res.err ~= nil then
    return true, res.err
  end
  return false, nil
end

-- add members to zset with size limit
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
      redis.log(redis.LOG_WARNING, 'zpopmin failed in counter_sizelimit_batchadd: ' .. err)
    end
  end

  -- step3. zdd members
  return redis.pcall('ZADD', key, unpack(arg_members))
end

-- check if someone has acted ActDo record on specific bizcode and oid
local function counter_check_actdo_record(keys, args)
  local counter_list_key = keys[1]
  local counter_record_key = keys[2]
  local counter_list_member = args[1]
  
  -- step1. check counter_list_key first
  local result = redis.pcall('ZSCORE', counter_list_key, counter_list_member)
  local is_err, err = is_pcall_err(result)
  if is_err then
    redis.log(redis.LOG_WARNING, 'zscore failed in counter_check_actdo_record: ' .. err)
  else
    result = tonumber(result)
    if result ~= nil and result > 0 then
      return 1
    end
  end

  -- step2. if no record is found in counter_list_key, we try to find it in counter_record_key
  

  return 0
end

-- register redis functions
redis.register_function('counter_sizelimit_batchadd', counter_sizelimit_batchadd)
redis.register_function('counter_check_actdo_record', counter_check_actdo_record)
