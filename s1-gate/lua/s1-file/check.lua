local resplib = require('common.resp')
local httpstatus = require('http.status')
local httpmethod = require('http.method')

-- requests
local req_method = ngx.req.get_method():upper()
if req_method ~= httpmethod.GET and
    req_method ~= httpmethod.HEAD and
    req_method ~= httpmethod.OPTIONS then
  resplib.make_status_resp(httpstatus.HTTP_METHOD_NOT_ALLOWED, 'method not allowed')
  return
end
