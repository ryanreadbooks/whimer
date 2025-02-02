--[[
Copyright 2025 ryanreadbooks
Copyright 2024 Adrian Astley (RichieSams)
Copyright 2018 JobTeaser

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
--]]

local resty_hmac = require('resty.hmac')
local resty_sha256 = require('resty.sha256')
local str = require('resty.string')
local env = require('common.env')

local _M = {}


local reserved_object_names = "^[a-zA-Z0-9-_.~/]+$"

local function is_reserved_object_name(path)
  return string.match(path, reserved_object_names) ~= nil
end

local function encode_char(ch)
  return string.format("%%%02X", string.byte(ch))
end

local function encode_path(path)
  if is_reserved_object_name(path) then
    return path
  end

  local encoded_path = {}
  for i = 1, #path do
    local char = path:sub(i, i)
    -- §2.3 Unreserved characters (mark)
    if string.match(char, "[A-Za-z0-9]") or
        string.match(char, "[-_.~/]") then
      table.insert(encoded_path, char)
    else
      table.insert(encoded_path, encode_char(char))
    end
  end

  return table.concat(encoded_path)
end

---------------------------------
--- Helper functions
---------------------------------

---@alias Credentials {access_key:string, secret_key:string}

---Fetches the AWS credentials from ENV variables and returns them in a table
---@return Credentials
local function get_credentials()
  local access_key = env.get_aws_access_key_id()
  local secret_key = env.get_aws_secret_access_key()

  return {
    access_key = access_key,
    secret_key = secret_key
  }
end

---Formats the given timestamp in ISO 8601
---@param timestamp integer The Unix timestamp to format
---@return string
local function get_iso8601_basic(timestamp)
  return tostring(os.date('!%Y%m%dT%H%M%SZ', timestamp))
end

---Formats the given timestamp in ISO 8601 shortened
---@param  timestamp integer The timestamp in Unix Epoch seconds to format
---@return           string
local function get_iso8601_basic_short(timestamp)
  return tostring(os.date('!%Y%m%d', timestamp))
end

---Calculates and returns the Derived Signing Key
---@param  keys      Credentials Credentials returned from get_credentials()
---@param  timestamp integer     The time as a Unix timestamp
---@param  region    string      The AWS region the request will use
---@param  service   string      The AWS service the request will use
---@return           string
local function get_derived_signing_key(keys, timestamp, region, service)
  local h_date = resty_hmac:new('AWS4' .. keys['secret_key'], resty_hmac.ALGOS.SHA256)
  h_date:update(get_iso8601_basic_short(timestamp))
  local k_date = h_date:final()

  local h_region = resty_hmac:new(k_date, resty_hmac.ALGOS.SHA256)
  h_region:update(region)
  local k_region = h_region:final()

  local h_service = resty_hmac:new(k_region, resty_hmac.ALGOS.SHA256)
  h_service:update(service)
  local k_service = h_service:final()

  local h = resty_hmac:new(k_service, resty_hmac.ALGOS.SHA256)
  h:update('aws4_request')
  return h:final()
end

---Calculates and returns the credential scope
---
---See: https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_sigv-create-signed-request.html#create-string-to-sign
---@param  timestamp integer The time as a Unix timestamp
---@param  region    string  The AWS region the request will use
---@param  service   string  The AWS service the request will use
---@return           string
local function get_cred_scope(timestamp, region, service)
  return get_iso8601_basic_short(timestamp)
      .. '/' .. region
      .. '/' .. service
      .. '/aws4_request'
end

---Returns the list of headers that we are signing, concatenated with semicolons
---
---See: https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_sigv-create-signed-request.html#create-canonical-request
---@return string
local function get_signed_headers()
  return 'host;x-amz-content-sha256;x-amz-date'
end

---Returns The SHA256 hex-encoded digest of the request body of the input
---@param  s string The input to hash
---@return   string
local function get_sha256_digest(s)
  local h = resty_sha256:new()
  h:update(s or '')
  return str.to_hex(h:final())
end

---Calculates the SHA256 hashed canonical request.
---
---See: https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_sigv-create-signed-request.html#create-canonical-request
---@param  method      string  The http method
---@param  timestamp   integer The current time in Unix Epoch seconds
---@param  host        string  The upstream host
---@param  uri         string  The path portion of the request URI
---@param  body_digest string  The SHA256 hex-encoded digest of the request body
---@return             string
local function get_hashed_canonical_request(method, timestamp, host, uri, body_digest)
  local canonical_request = method .. '\n'
      .. encode_path(uri) .. '\n'
      .. '\n'
      .. 'host:' .. host .. '\n'
      .. 'x-amz-content-sha256:' .. body_digest .. '\n'
      .. 'x-amz-date:' .. get_iso8601_basic(timestamp) .. '\n'
      .. '\n'
      .. get_signed_headers() .. '\n'
      .. body_digest
  return get_sha256_digest(canonical_request)
end

---Calculates and returns the "string to sign"
---
---See: https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_sigv-create-signed-request.html#create-string-to-sign
---@param  method      string  The http method
---@param  timestamp   integer The current time in Unix Epoch seconds
---@param  region      string  The AWS region the request will use
---@param  service     string  The AWS service the request will use
---@param  host        string  The upstream host
---@param  uri         string  The path portion of the request URI
---@param  body_digest string  The SHA256 hex-encoded digest of the request body
---@return             string
local function get_string_to_sign(method, timestamp, region, service, host, uri, body_digest)
  return 'AWS4-HMAC-SHA256\n'
      .. get_iso8601_basic(timestamp) .. '\n'
      .. get_cred_scope(timestamp, region, service) .. '\n'
      .. get_hashed_canonical_request(method, timestamp, host, uri, body_digest)
end

---Signs the given string using the given key with the HMAC SHA256 algorithm
---
---See: https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_sigv-create-signed-request.html#calculate-signature
---@param derived_signing_key string The signing key
---@param string_to_sign      string The string to sign
---@return                    string
local function get_signature(derived_signing_key, string_to_sign)
  local h = resty_hmac:new(derived_signing_key, resty_hmac.ALGOS.SHA256)
  h:update(string_to_sign)
  return h:final(nil, true)
end

---Calculates and returns the appropriate value for the Authorization header
---
---See: https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_sigv-create-signed-request.html#add-signature-to-request
---@param  method      string      The http method
---@param  keys        Credentials Credentials returned from get_credentials()
---@param  timestamp   integer     The current time in Unix Epoch seconds
---@param  region      string      The AWS region the request will use
---@param  service     string      The AWS service the request will use
---@param  host        string      The upstream host
---@param  uri         string      The path portion of the request URI
---@param  body_digest string      The SHA256 hex-encoded digest of the request body
---@return             string
local function get_authorization(method, keys, timestamp, region, service, host, uri, body_digest)
  local derived_signing_key = get_derived_signing_key(keys, timestamp, region, service)
  local string_to_sign = get_string_to_sign(method, timestamp, region, service, host, uri, body_digest)
  local auth = 'AWS4-HMAC-SHA256 '
      .. 'Credential=' .. keys['access_key'] .. '/' .. get_cred_scope(timestamp, region, service)
      .. ',SignedHeaders=' .. get_signed_headers()
      .. ',Signature=' .. get_signature(derived_signing_key, string_to_sign)
  return auth
end


---------------------------------
--- Exported module functions
---------------------------------

---Calculates and returns a table of request headers for an authenticated AWS request
---
---This function takes as an argument the request body so that it can hash it to include in
---the authentication signature. If you want to avoid this overhead, you can use
---aws_signed_headers_unsigned_payload()
---
---Note: This function requires the AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment
---      variables to be set. You must expose them to LUA in your nginx.conf using:
---
--- ```
--- env AWS_ACCESS_KEY_ID;
--- env AWS_SECRET_ACCESS_KEY;
--- ```
---
---See: https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-authenticating-requests.html
---@param  method  string  The http method
---@param  host    string  The upstream host
---@param  uri     string  The path portion of the request URI
---@param  region  string  The AWS region the request will use
---@param  service string  The AWS service the request will use
---@param  body    string  The request body
---@return         table   # The table of headers to apply to your request
function _M.aws_signed_headers(method, host, uri, region, service, body)
  local body_digest = get_sha256_digest(body)
  local timestamp = tonumber(ngx.time())

  return _M.aws_signed_headers_detailed(host, uri, region, service, body_digest, timestamp, method)
end

---Calculates and returns a table of request headers for an authenticated AWS request
---
---This function will skip the request body digest calculation. Which saves the cost and
---time of hashing the entire request body. If you do want the request body digest
---to be part of the signature though, you can use aws_signed_headers() instead
---
---Note: This function requires the AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment
---      variables to be set. You must expose them to LUA in your nginx.conf using:
---
--- ```
--- env AWS_ACCESS_KEY_ID;
--- env AWS_SECRET_ACCESS_KEY;
--- ```
---
---See: https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-authenticating-requests.html
---@param  method  string  The http method
---@param  host    string  The upstream host
---@param  uri     string  The path portion of the request URI
---@param  region  string  The AWS region the request will use
---@param  service string  The AWS service the request will use
---@return         table   # The table of headers to apply to your request
function _M.aws_signed_headers_unsigned_payload(method, host, uri, region, service)
  local body_digest = 'UNSIGNED-PAYLOAD'
  local timestamp = tonumber(ngx.time())

  return _M.aws_signed_headers_detailed(host, uri, region, service, body_digest, timestamp, method)
end

---Calculates and returns a table of request headers for an authenticated AWS request
---
---This function is identical to aws_signed_headers(), but it allows you to set the body_digest
---and timestamp yourself. Unless you are doing something very special, you should generally
---just use aws_signed_headers() or aws_signed_headers_unsigned_payload()
---
---Note: This function requires the AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment
---      variables to be set. You must expose them to LUA in your nginx.conf using:
---
--- ```
--- env AWS_ACCESS_KEY_ID;
--- env AWS_SECRET_ACCESS_KEY;
--- ```
---
---See: https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-authenticating-requests.html
---@param  host        string  The upstream host
---@param  uri         string  The path portion of the request URI
---@param  region      string  The AWS region the request will use
---@param  service     string  The AWS service the request will use
---@param  body_digest string  The SHA256 hex-encoded digest of the request body
---@param  timestamp   integer The current time as a Unix timestamp
---@param  method      string  The http method
---@return            table    # The table of headers to apply to your request
function _M.aws_signed_headers_detailed(host, uri, region, service, body_digest, timestamp, method)
  local creds = get_credentials()

  local auth = get_authorization(method, creds, timestamp, region, service, host, uri, body_digest)

  return {
    ["authorization"] = auth,
    ["host"] = host,
    ["x-amz-date"] = get_iso8601_basic(timestamp),
    ["x-amz-content-sha256"] = body_digest,
  }
end

return _M
