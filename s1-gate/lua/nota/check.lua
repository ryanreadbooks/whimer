local common = require('common.resp')
local httpstatus = require('http.status')
local httpmethod = require('http.method')
local imgsniff = require('mime.imgsniff')
local ctx = require('nota.ctx')

-- constant
local MAX_BODY_BYTES_ALLOWED = 10 * 1024 * 1024 -- 10M
local ALLOWED_CONTENT_TYPE = {
  'image/jpeg',
  'image/png',
  'image/webp'
}
local SOCK_TIMEOUT = 30 * 1000 -- 30s

-- requests
local req_method = string.upper(ngx.req.get_method())
if req_method ~= httpmethod.PUT and req_method ~= httpmethod.GET then
  common.make_status_resp(httpstatus.HTTP_METHOD_NOT_ALLOWED, 'method not allowed')
  return
end

local req_headers = ngx.req.get_headers()

-- this is considered as a upload request
-- 1. check request method
if req_method == httpmethod.PUT then
  -- 1.1 content-length header is required
  local content_length = tonumber(req_headers['Content-Length']) or 0
  if content_length == 0 then
    common.make_status_resp(httpstatus.HTTP_LENGTH_REQUIRED, 'content-length is required')
    return
  elseif content_length > MAX_BODY_BYTES_ALLOWED then
    common.make_status_resp(httpstatus.HTTP_REQUEST_ENTITY_TOO_LARGE, 'payload is too large')
    return
  end

  -- 1.2 make sure authorization is in header
  local authorization = req_headers['Authorization'] or ''
  if #authorization == 0 then
    common.make_403_err('authorization is required')
    return
  end

  -- TODO 1.3 furthur checking for authorization

  -- 1.4 we need to check payload mime type
  local sock, err = ngx.req.socket()
  if err ~= nil then
    ngx.log(ngx.ERROR, 'ngx.req.socket err: ' .. err)
    common.make_500_err('socket internal error')
    return
  end

  sock:settimeout(SOCK_TIMEOUT) -- ms

  -- 1.4.1 read body in a streaming way
  -- In case of success, it returns the data received;
  -- in case of error, it returns nil with a string describing the error and
  -- the partial data received so far.
  local body_header, p
  body_header, err, p = sock:receive(imgsniff.MAX_SNIFF_BYTE)
  if err ~= nil then
    ngx.log(ngx.WARN, 'sock:receive err: ' .. err)
    sock:close()
    return
  end

  local magic = body_header:sub(1, imgsniff.MAX_SNIFF_BYTE)
  local detected_content_type = imgsniff.detect(magic) -- detected content type
  local is_allowed = false
  for i = 1, #ALLOWED_CONTENT_TYPE do
    if detected_content_type == ALLOWED_CONTENT_TYPE[i] then
      is_allowed = true
      break
    end
  end
  if not is_allowed then
    common.make_400_err('unsupported content-type')
    return
  end

  -- 1.5 read the rest of the body and prepare for later usage
  local body_rest = sock:receive('*a')

  -- 2. save all the related requests data into ctx
  ngx.ctx.requests = {}
  ngx.ctx.requests[ctx.REQUESTS_CONTENT_LENGTH] = content_length
  ngx.ctx.requests[ctx.REQUESTS_CONTENT_TYPE] = detected_content_type
  ngx.ctx.requests[ctx.REQUESTS_BODY_DATA] = {
    header = body_header,
    rest = body_rest
  }
end
