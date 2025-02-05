local headerlib = require('common.header')
local aws = require('auth.awsv4')

local helper = {}

-- generate auth header
function helper.get_oss_auth_header(method, host, key, location)
  return aws.aws_signed_headers(
    method,
    host,
    key,
    location,
    's3',
    ''
  )
end

-- generate auth header without signed payload
function helper.get_oss_auth_header_unsigned_payload(method,host, key, location)
  return aws.aws_signed_headers_unsigned_payload(
    method,
    host,
    key,
    location,
    's3'
  )
end

-- read minio response data
function helper.read_oss_resp(res, bufsize)
  local reader = res.body_reader
  local buffer_size = bufsize or 512
  local buffers = {}

  repeat
    local buffer, err = reader(buffer_size)
    if err then
      ngx.log(ngx.ERR, err)
      break
    end

    if buffer then
      table.insert(buffers, buffer)
    end
  until not buffer

  return table.concat(buffers, '')
end

function helper.copy_req_header_except(key)
  local copied = headerlib.canonical_header()
  for k, v in pairs(ngx.req.get_headers()) do
    if k:lower() ~= key:lower() then -- ignore this header cause we will replace it
      copied[k] = v
    end
  end

  return copied
end

return helper
