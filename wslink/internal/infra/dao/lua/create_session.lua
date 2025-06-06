-- step1. set session key-value pair
local sid = KEYS[1]
local sval = ARGV[1]
local uid_key = KEYS[2]

local function is_pcall_err(res)
  return type(res) == 'table' and res['err'] ~= nil
end

local r1 = redis.pcall('SET', sid, sval)
if is_pcall_err(r1) then
  -- return immediately
  return redis.error_reply('ERR set session: ' .. sid)
end

-- step2. assign sess id to uid session
local r2 = redis.pcall('SADD', uid_key, sid)
if is_pcall_err(r2) then
  -- try best effort to rollback the first redis call
  redis.pcall('DEL', sid)
  return redis.error_reply('ERR set uid session: ' .. uid_key)
end
