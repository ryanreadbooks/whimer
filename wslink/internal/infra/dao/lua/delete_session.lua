local sid = KEYS[1]
local uid_key = KEYS[2]

local function is_pcall_err(res)
  return type(res) == 'table' and res['err'] ~= nil
end

-- step1. delete session
local r1 = redis.pcall('DEL', sid)
if is_pcall_err(r1) then
  return redis.error_reply('ERR del session: ' .. sid)
end

-- step2. delete uid session
local r2 = redis.pcall('SREM', uid_key, sid)
if is_pcall_err(r2) then
  return redis.error_reply('ERR del uid session: ' .. uid_key)
end
