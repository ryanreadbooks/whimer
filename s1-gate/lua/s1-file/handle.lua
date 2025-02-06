-- we just proxy pass to internal path
local res = ngx.location.capture('/imgproxyserver' .. ngx.var.uri, {
  ctx = ngx.ctx
})

ngx.status = res.status
-- add header
for k, v in pairs(res.header) do
  ngx.header[k] = v
end

ngx.print(res.body)   -- send response body back
