#!lua name=libmisc

local function is_pcall_err(res)
  if type(res) == 'table' and res.err ~= nil then
    return true, res.err
  end
  return false, nil
end

redis.register_function('is_pcall_err', is_pcall_err)