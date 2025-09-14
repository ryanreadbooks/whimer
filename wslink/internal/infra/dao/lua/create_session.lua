local sid = KEYS[1]
local uid_key = KEYS[2]

local function is_pcall_err(res)
  return type(res) == 'table' and res['err'] ~= nil
end

-- step1. set session hash
local r1 = redis.pcall('HSET', sid, unpack(ARGV))
if is_pcall_err(r1) then
  -- return immediately
  return redis.error_reply('failed to set session: ' .. sid)
end

-- step2. assign sess id to uid session
local r2 = redis.pcall('SADD', uid_key, sid)
if is_pcall_err(r2) then
  -- try best effort to rollback the first redis call
  redis.pcall('DEL', sid)
  return redis.error_reply('failed to set uid session: ' .. uid_key)
end
