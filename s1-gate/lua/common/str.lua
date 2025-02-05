local _M = {}

function _M.split(s, sep)
  local t = {}
  local i = 1
  for str in string.gmatch(s, "([^" .. sep .. "]+)") do
    t[i] = str
    i = i + 1
  end
  return t
end

function _M:trim(s)
  local from = s:match "^%s*()"
  return s:sub(from, #s - (s:reverse():match("^%s*()") - 1))
end

return _M
