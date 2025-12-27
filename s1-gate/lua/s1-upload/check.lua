local resplib = require('common.resp')
local httpstatus = require('http.status')
local httpmethod = require('http.method')
local imgsniff = require('mime.imgsniff')
local ctx = require('s1-upload.ctx')
local env = require('common.env')
local const = require('common.const')
local jwt = require('resty.jwt')
local whmrauth = require('auth.whmr')
local iso8601 = require('time.iso8601')

ngx.log(ngx.DEBUG, 's1-upload.check is working...')
ngx.header['Access-Control-Allow-Origin'] = env.get_cors_allowed_origin()
ngx.header['Access-Control-Allow-Credentials'] = 'true'

-- requests
local req_method = ngx.req.get_method():upper()
if req_method ~= httpmethod.PUT and
    req_method ~= httpmethod.HEAD and
    req_method ~= httpmethod.OPTIONS then
  resplib.make_status_resp(httpstatus.HTTP_METHOD_NOT_ALLOWED, 'method not allowed')
  return
end

local req_headers = ngx.req.get_headers()

-- 1. check upload request
if req_method == httpmethod.PUT then
  -- 1. make sure required headers are present
  local content_length = tonumber(req_headers['Content-Length']) or 0
  if content_length == 0 then
    resplib.make_status_resp(httpstatus.HTTP_LENGTH_REQUIRED, 'content-length is required in header')
    return
  elseif content_length > const.MAX_BODY_BYTES_ALLOWED then
    resplib.make_status_resp(httpstatus.HTTP_REQUEST_ENTITY_TOO_LARGE, 'payload is too large')
    return
  end

  local host = req_headers['Host'] or ''
  if #host == 0 then
    resplib.make_403_err('host is required in header')
    return
  end

  local token = req_headers['X-Security-Token'] or ''
  if #token == 0 then
    resplib.make_403_err('x-security-token is required in header')
    return
  end

  local jwt_obj = jwt:verify(env.get_aws_secret_access_key(), token, {
    require_exp_claim = true,
    valid_issuers = { env.get_jwt_valid_issuer() }
  })
  if not jwt_obj['verified'] then
    resplib.make_403_err('invalid x-security-token ' .. jwt_obj['reason'])
    return
  end
  if jwt_obj['payload']['sub'] ~= env.get_jwt_valid_subject() then
    resplib.make_403_err('invalid x-security-token subject')
    return
  end

  local access_key = jwt_obj['payload']['access_key'] or ''
  local date = req_headers['X-Date'] or ''
  if #date == 0 then
    resplib.make_403_err('x-date is required in header')
    return
  end
  if not iso8601.is_valid_datetime(date) then
    resplib.make_403_err('x-date invalid format')
    return
  end

  local authorization = req_headers['Authorization'] or ''
  if #authorization == 0 then
    resplib.make_403_err('authorization is required in header')
    return
  end

  -- check authorization
  local res = whmrauth.sign_request(access_key)
  if res ~= authorization then
    resplib.make_403_err('The request signature we calculated does not match the signature you provided')
    return
  end

  -- 2. check mime type
  local sock, err = ngx.req.socket()
  if err ~= nil then
    ngx.log(ngx.ERROR, 'ngx.req.socket err: ' .. err)
    resplib.make_500_err('socket internal error')
    return
  end

  sock:settimeout(const.SOCK_TIMEOUT) -- ms

  -- read body in a streaming way
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
  for i = 1, #const.ALLOWED_CONTENT_TYPE do
    if detected_content_type == const.ALLOWED_CONTENT_TYPE[i] then
      is_allowed = true
      break
    end
  end
  if not is_allowed then
    resplib.make_400_err('unsupported content-type')
    return
  end

  -- 3. read the rest of the body and prepare for later usage
  local body_rest = sock:receive('*a')

  -- 4. save all the related requests data into ctx
  ngx.ctx.requests = {}
  ngx.ctx.requests[ctx.REQUESTS_CONTENT_LENGTH] = content_length
  ngx.ctx.requests[ctx.REQUESTS_CONTENT_TYPE] = detected_content_type
  ngx.ctx.requests[ctx.REQUESTS_BODY_DATA] = {
    header = body_header,
    rest = body_rest
  }
elseif req_method == httpmethod.OPTIONS then
  -- add headers
  ngx.header['Access-Control-Allow-Headers'] =
  'Authorization,Cache-Control,Content-Type,X-Security-Token,X-Date,Access-Control-Allow-Credentials'
  ngx.header['Access-Control-Allow-Methods'] = 'PUT,HEAD,OPTIONS'
  ngx.header['Access-Control-Expose-Headers'] = '*'
end
