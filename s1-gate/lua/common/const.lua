local _M = {}

-- constant
_M.MAX_BODY_BYTES_ALLOWED = 10 * 1024 * 1024 -- 10M
_M.ALLOWED_CONTENT_TYPE = {
  'image/jpeg',
  'image/png',
  'image/webp'
}
_M.SOCK_TIMEOUT = 30 * 1000 -- 30s

return _M
