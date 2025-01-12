local _M = {}

local req_headers_mt = {
  __index = function(tb, key)
    if type(key) == 'string' then
      key = key:lower()
    end
    return rawget(tb, key)
  end,

  __newindex = function(tb, key, val)
    if type(key) == 'string' then
      key = key:lower()
    end
    return rawset(tb, key, val)
  end,

  size = function(tbl)
    return #tbl
  end
}

-- try to get #header length is not working
function _M.canonical_header()
  local headers = {}
  headers = setmetatable(headers, req_headers_mt)
  return headers
end

return _M
