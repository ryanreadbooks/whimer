local _M = {}

local function is_leap_year(year)
  return (year % 4 == 0 and year % 100 ~= 0) or (year % 400 == 0)
end

local function get_max_day(year, month)
  local max_days = { 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31 }
  if month == 2 and is_leap_year(year) then
    return 29
  else
    return max_days[month]
  end
end
function _M.get_datetime(timestamp)
  return tostring(os.date('!%Y%m%dT%H%M%SZ', timestamp))
end

function _M.get_date(timestamp)
  return tostring(os.date('!%Y%m%d', timestamp))
end

-- format: 20231001T235959Z
function _M.is_valid_datetime(datetime_str)
  if #datetime_str ~= 16 then
    return false
  end

  -- YYYYMMDDTHHMMSSZ
  if not datetime_str:match("^%d%d%d%d%d%d%d%dT%d%d%d%d%d%dZ$") then
    return false
  end

  local year = tonumber(datetime_str:sub(1, 4))
  local month = tonumber(datetime_str:sub(5, 6))
  local day = tonumber(datetime_str:sub(7, 8))
  local hour = tonumber(datetime_str:sub(10, 11))
  local minute = tonumber(datetime_str:sub(12, 13))
  local second = tonumber(datetime_str:sub(14, 15))

  if month < 1 or month > 12 then
    return false
  end

  local max_days = { 31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31 }
  if day < 1 or day > get_max_day(year, month) then
    return false
  end

  if hour > 23 then
    return false
  end

  if minute > 59 or second > 59 then
    return false
  end

  return true
end

return _M
