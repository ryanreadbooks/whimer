local aws4 = require('auth.awsv4')
local resty_hmac = require('resty.hmac')
local resty_sha256 = require('resty.sha256')
local str = require('resty.string')

local _M = {}

function _M.hmac_sha256_as_hex(key, content)
  local h = resty_hmac:new(key, resty_hmac.ALGOS.SHA256)
  h:update(content)
  return h:final(nil, true)
end

function _M.hmac_sha256(key, content)
  local h = resty_hmac:new(key, resty_hmac.ALGOS.SHA256)
  h:update(content)
  return h:final(nil, false)
end

function _M.sign_request(secret)
  local req = ngx.req
  local headers = req.get_headers()
  local ctl = headers["Content-Length"] or ""
  local host = headers["Host"] or ""
  local x_date = headers["X-Date"] or ""
  local x_secu_token = headers["X-Security-Token"] or ""
  local method = ngx.req.get_method():upper()
  local uri = aws4.encode_path(ngx.var.uri)

  local canonical_request = method .. "\n" ..
      uri .. "\n" ..
      "content-length:" .. ctl .. "\n" ..
      "host:" .. host .. "\n" ..
      "x-date:" .. x_date .. "\n" ..
      "x-security-token:" .. x_secu_token .. "\n" ..
      "\n" ..
      "content-length;host;x-date;x-security-token" .. "\n"

  local hashed_canonical_request = _M.hmac_sha256_as_hex(secret, canonical_request)
  local string_to_sign = "SHA256" .. "\n" .. x_date .. "\n" .. hashed_canonical_request

  -- derived signing key
  local seck = "WHMR" .. secret
  local signing_key = _M.hmac_sha256(seck, x_date)
  signing_key = _M.hmac_sha256(signing_key, "whmr_request")

  -- 计算最终签名并转十六进制
  local signature = _M.hmac_sha256_as_hex(signing_key, string_to_sign)
  return signature
end

return _M
