local ctx = require('s1-upload.ctx')
local helper = require('s1-upload.handle_helper')
local resplib = require('common.resp')
local httpmethod = require('http.method')
local envlib = require('common.env')
local httpclient = require('resty.http').new()

ngx.log(ngx.DEBUG, 's1-upload.handle is working...')

local oss_host = envlib.get_oss_endpoint_host()
local oss_port = envlib.get_oss_endpoint_port()
local oss_location = envlib.get_oss_endpoint_location()
local obj_key = ngx.var.uri

if ctx.is_upload_request() then                             -- the request is a upload request
  local req_body = ngx.ctx.requests[ctx.REQUESTS_BODY_DATA] -- kv table
  if not req_body then
    resplib.make_400_err('empty body data')
    return
  end

  -- calculate aws v4 signature headers for heading object
  local head_headers = helper.get_oss_auth_header(httpmethod.HEAD, oss_host, obj_key, oss_location)

  local ok, err, session = httpclient:connect({
    scheme = 'http',
    host = oss_host,
    port = oss_port,
  })

  if not ok then
    ngx.log(ngx.ERR,
      string.format('connection to oss server at %s:%d failed, err: %s',
        oss_host,
        oss_port,
        err)
    )

    resplib.make_500_err('oss server abnormal')
    return
  end

  -- we should check object existence first to prevent duplication
  local head_res
  head_res, err = httpclient:request({
    path = obj_key,
    method = httpmethod.HEAD,
    headers = head_headers,
  })
  if err ~= nil then
    resplib.make_500_err('can not connect to oss server head ' .. err)
    return
  end

  if head_res.status == 200 then
    -- object already exists
    resplib.make_403_err('object already exists')
    return
  end

  -- calculate aws v4 headers for putting object
  local put_headers = helper.get_oss_auth_header_unsigned_payload(httpmethod.PUT, oss_host, obj_key, oss_location)
  local merged_headers = helper.copy_req_header_except('authorization')
  -- append aws auth headers
  for k, v in pairs(put_headers) do
    merged_headers[k] = v
  end

  -- override
  merged_headers['Content-Type'] = ngx.ctx.requests[ctx.REQUESTS_CONTENT_TYPE]
  merged_headers['Content-Length'] = ngx.ctx.requests[ctx.REQUESTS_CONTENT_LENGTH]

  -- perform http PUT request to upload
  local res
  local body = table.concat({ req_body.header, req_body.rest })

  res, err = httpclient:request({
    path = obj_key,
    method = httpmethod.PUT,
    body = body,
    headers = merged_headers
  })

  if err ~= nil then
    resplib.make_500_err('can not connect to oss server ' .. err)
    return
  end

  if res.status ~= 200 then
    resplib.make_status_resp(res.status, res.reason)
    ngx.log(ngx.WARN,
      string.format('do oss request returns non 200, status=%d, reason=%s, body=%s',
        res.status, res.reason, helper.read_oss_resp(res)
      )
    )
    return
  end
else
end
