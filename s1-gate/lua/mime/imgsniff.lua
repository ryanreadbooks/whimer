local bit = require('bit')

local _M = {}

local MAX_SNIFF_BYTE = 32

_M.MAX_SNIFF_BYTE = MAX_SNIFF_BYTE

local ISniffSig = {}

function ISniffSig:match(data)
  error("not implemented")
end

function ISniffSig:new()
  local instance = {}
  setmetatable(instance, self)
  self.__index = self
  return instance
end

local jpegSniffSig = ISniffSig:new()
function jpegSniffSig.match(data)
  if not data then
    return ''
  end

  local header = '\xFF\xD8\xFF'
  if #data >= #header and data:sub(1, #header) == header then
    return "image/jpeg"
  else
    return ''
  end
end

local pngSniffSig = ISniffSig:new()
function pngSniffSig.match(data)
  if not data then
    return ''
  end

  local header = '\x89\x50\x4E\x47\x0D\x0A\x1A\x0A'
  if #data >= #header and data:sub(1, #header) == header then
    return "image/png"
  else
    return ''
  end
end

local webpSniffSig = ISniffSig:new()
function webpSniffSig.match(data)
  if not data then
    return ''
  end

  local mask = '\xFF\xFF\xFF\xFF\x00\x00\x00\x00\xFF\xFF\xFF\xFF\xFF\xFF'
  local pattern = 'RIFF\x00\x00\x00\x00WEBPVP'
  if #data < #pattern then
    return ''
  end

  for i = 1, #pattern do
    local t1 = data:byte(i)
    local t2 = mask:byte(i)
    local masked = bit.band(t1, t2)
    if masked ~= pattern:byte(i) then
      return ''
    end
  end

  return 'image/webp'
end

local sniffSignatures = {
  jpegSniffSig,
  pngSniffSig,
  webpSniffSig,
}

-- sniff image format image/jpeg, image/png, image/webp
function _M.detect(data)
  if data == nil then
    return ''
  end

  if #data > MAX_SNIFF_BYTE then
    data = data:sub(1, MAX_SNIFF_BYTE)
  end

  -- do sniff
  for i = 1, #sniffSignatures do
    local ct = sniffSignatures[i].match(data)
    if ct ~= '' then
      return ct
    end
  end

  return ''
end

return _M
