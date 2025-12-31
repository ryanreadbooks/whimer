local cjson = require('cjson')


local _M = {}

local xml_err_tmpl =
'<?xml version="1.0" encoding="UTF-8"?><Error><Code>%s</Code><Message>%s</Message><Resource>%s</Resource></Error>'

function _M.make_err_resp(msg)
  local resp = {
    code = -1,
    msg = msg
  }
  return resp
end

function _M.make_err_json_resp(msg)
  local resp = {
    code = -1,
    msg = msg
  }

  return cjson.encode(resp)
end

function _M.make_status_resp(status, msg)
  ngx.status = status
  ngx.header['Content-Type'] = 'application/json'
  ngx.say(_M.make_err_json_resp(msg))
end

function _M.make_403_err(msg)
  _M.make_status_resp(ngx.HTTP_FORBIDDEN, msg)
end

function _M.make_400_err(msg)
  _M.make_status_resp(ngx.HTTP_BAD_REQUEST, msg)
end

function _M.make_500_err(msg)
  _M.make_status_resp(ngx.HTTP_INTERNAL_SERVER_ERROR, msg)
end

function _M.make_err_xml_resp(code, msg, key)
  return string.format(xml_err_tmpl, code, msg, key)
end

function _M.make_s3_xml_resp(status, code, msg, key)
  ngx.status = status
  ngx.header['Content-Type'] = 'application/xml'
  ngx.say(_M.make_err_xml_resp(code, msg, key))
end

function _M.make_403_s3_xml_resp(code, msg, key)
  _M.make_s3_xml_resp(ngx.HTTP_FORBIDDEN, code, msg, key)
end

return _M
