local httpmethod = require('http.method')
local envlib = require('common.env')
local resplib = require('common.resp')
local helper = require('s1-upload.handle_helper')
local ctx = require('s1-upload.ctx')

-- make sure this is a upload request
if ctx.is_upload_request() then
  local oss_host = envlib.get_oss_endpoint_host()
  local oss_port = envlib.get_oss_endpoint_port()
  local oss_location = envlib.get_oss_endpoint_location()
  local obj_key = ngx.var.uri

  -- we only check key is already exists in oss here
  local head_headers = helper.get_oss_auth_header(httpmethod.HEAD, oss_host, obj_key, oss_location)
  local httpclient = require("resty.http").new()
  local path = string.format('http://%s:%s%s', oss_host, oss_port, obj_key)
  local res, err = httpclient:request_uri(path, {
    method = httpmethod.HEAD,
    headers = head_headers,
  })

  if not res then
    ngx.log(ngx.ERR,
      string.format('connection to oss server at %s:%d failed, err: %s',
        oss_host,
        oss_port,
        err)
    )

    resplib.make_500_err('oss server abnormal')
    return
  end

  if res.status == 200 then
    -- object already exists
    -- resplib.make_403_err('object already exists')
    resplib.make_403_s3_xml_resp('KeyAlreadyExists', 'The uploading key already exists', obj_key)
    return
  else
    -- non 200 and not 404 then it is considered as an internal error
    if res.status ~= 404 then
      local log_str
      for k, v in pairs(res.headers) do
        log_str = log_str .. ' ' .. string.format('%s=%s', k, v)
      end

      log_str = string.format('head object return status = %d, headers = %s', res.status, log_str)
      ngx.log(ngx.ERR, log_str)
      resplib.make_500_err('s1-upload gate internal error')
    end
  end
end
