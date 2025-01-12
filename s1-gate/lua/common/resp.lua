local cjson = require('cjson')


local _M = {}


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

return _M
