local httpmethod = require('http.method')

local _M = {}

_M.REQUESTS_CONTENT_LENGTH = 'content-length'
_M.REQUESTS_CONTENT_TYPE = 'content-type'
_M.REQUESTS_BODY_DATA = 'body-data'

function _M.is_upload_request()
  local method_name = ngx.req.get_method():upper()
  return method_name == httpmethod.PUT or method_name == httpmethod.POST
end

return _M
