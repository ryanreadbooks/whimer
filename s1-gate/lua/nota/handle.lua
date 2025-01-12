local ctx = require('nota.ctx')
local httpc = require('resty.http').new()
local canonicalhd = require('common.header')
local common = require('common.resp')
local httpmethod = require('http.method')
local aws = require('aws.v4_sign')
local env = require('common.env')

local oss_host = env.get_oss_endpoint_host()
local oss_port = env.get_oss_endpoint_port()
local oss_location = env.get_oss_endpoint_location()
local obj_key = ngx.var.uri

-- generate auth header
local function get_oss_auth_header(method)
  return aws.aws_signed_headers(
    method,
    oss_host,
    obj_key,
    oss_location,
    's3',
    ''
  )
end


-- generate auth header without signed payload
local function get_oss_auth_header_unsigned_payload(method)
  return aws.aws_signed_headers_unsigned_payload(
    method,
    oss_host,
    obj_key,
    oss_location,
    's3'
  )
end

-- read minio response data
local function read_oss_resp(res, bufsize)
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


local function copy_req_header_except(key)
  local copied = canonicalhd.canonical_header()
  for k, v in pairs(ngx.req.get_headers()) do
    if k:lower() ~= key:lower() then -- ignore this header cause we will replace it
      copied[k] = v
    end
  end

  return copied
end


if ctx.is_upload_request() then                         -- the request is a upload request
  local body = ngx.ctx.requests[ctx.REQUESTS_BODY_DATA] -- kv table
  local body_arr = { body.header, body.rest }
  if not body then
    ngx.log(ngx.ERROR, 'ctx content body data is nil' .. ngx.var.uri)
    common.make_400_err('empty body data')
    return
  end

  -- calculate aws v4 signature headers for heading object
  local head_headers = get_oss_auth_header(httpmethod.HEAD)

  local ok, err, session = httpc:connect({
    scheme = 'http',
    host = oss_host,
    port = oss_port,
  })

  if not ok then
    ngx.log(ngx.ERROR,
      string.format('connection to oss server at %s:%d failed, err: %s',
        oss_host,
        oss_port,
        err)
    )

    common.make_500_err('oss server abnormal')
    return
  end

  -- we should check object existence first to prevent duplication
  local head_res
  head_res, err = httpc:request({
    path = obj_key,
    method = httpmethod.HEAD,
    headers = head_headers,
  })
  if err ~= nil then
    common.make_500_err('can not connect to oss server head ' .. err)
    return
  end

  if head_res.status == 200 then
    -- object already exists
    common.make_403_err('object already exists')
    return
  end

  -- calculate aws v4 headers for putting object
  local put_headers = get_oss_auth_header_unsigned_payload(httpmethod.PUT)
  local merged_headers = copy_req_header_except('authorization')
  -- append aws auth headers
  for k, v in pairs(put_headers) do
    merged_headers[k] = v
  end

  -- override
  merged_headers['Content-Type'] = ngx.ctx.requests[ctx.REQUESTS_CONTENT_TYPE]
  merged_headers['Content-Length'] = ngx.ctx.requests[ctx.REQUESTS_CONTENT_LENGTH]

  -- perform http PUT request to upload
  local res
  res, err = httpc:request({
    path = obj_key,
    method = httpmethod.PUT,
    body = body_arr,
    headers = merged_headers
  })

  if err ~= nil then
    common.make_500_err('can not connect to oss server ' .. err)
    return
  end

  if res.status ~= 200 then
    common.make_status_resp(res.status, res.reason)
    ngx.log(ngx.WARN,
      string.format('do oss request returns non 200, status=%d, reason=%s, body=%s',
        res.status, res.reason, read_oss_resp(res)
      )
    )
    return
  end
else -- the request is not a upload request
  -- for get method, we just add header and proxy pass
  local auth_headers = get_oss_auth_header(httpmethod.GET)
  ngx.log(ngx.INFO, 'auth => ')
  for k, v in pairs(auth_headers) do
    ngx.log(ngx.INFO, 'auth: ', k, ' = ', v)
  end
  for k, v in pairs(auth_headers) do
    ngx.req.set_header(k:lower(), v)
  end

  ngx.log(ngx.INFO, 'raw => ')
  for k, v in pairs(ngx.req.get_headers()) do
    ngx.log(ngx.INFO, 'raw: ', k, ' = ', v)
  end

  ngx.log(ngx.INFO, 'obj key is ', obj_key)
  
  -- proxy pass
  local res = ngx.location.capture('/minio-server' .. obj_key, {
    method = ngx.HTTP_GET,
    ctx = ngx.ctx
  })

  ngx.status = res.status
  for k, v in pairs(res.header) do
    ngx.header[k] = v
  end
  ngx.print(res.body)

  -- -- if using resty-http to finish proxy pass
  -- local ok, err, session = httpc:connect({
  --   scheme = 'http',
  --   host = oss_host,
  --   port = oss_port,
  -- })

  -- if not ok then
  --   ngx.log(ngx.ERROR,
  --     string.format('connection to oss server at %s:%d failed, err: %s',
  --       oss_host,
  --       oss_port,
  --       err)
  --   )

  --   common.make_500_err('oss server abnormal')
  --   return
  -- end
  -- local res
  -- res, err = httpc:request({
  --   path = obj_key,
  --   method = httpmethod.GET,
  --   headers = auth_headers
  -- })
  -- if err ~= nil then
  --   common.make_500_err('internal server error ' .. err)
  --   return
  -- end

  -- -- response
  -- ngx.status = res.status
  -- for k, v in pairs(res.headers) do
  --   ngx.header[k] = v
  -- end
  -- ngx.print(read_oss_resp(res, 512 * 1024))
end
