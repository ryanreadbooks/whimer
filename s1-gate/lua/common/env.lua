local _M = {}

local globals = ngx.shared.globals

function _M.get_oss_endpoint_host()
  local host, _ = globals:get('ENV_OSS_ENDPOINT_HOST')
  if host == nil then
    host = os.getenv('ENV_OSS_ENDPOINT_HOST') or '127.0.0.1'
    globals:set('ENV_OSS_ENDPOINT_HOST', host)
  end

  return host
end

function _M.get_oss_endpoint_port()
  local port, _ = globals:get('ENV_OSS_ENDPOINT_PORT')
  if port == nil then
    port = os.getenv('ENV_OSS_ENDPOINT_PORT') or 9000
    globals:set('ENV_OSS_ENDPOINT_PORT', port)
  end

  return port
end

function _M.get_oss_endpoint_location()
  local location, _ = globals:get('ENV_OSS_ENDPOINT_LOCATION')
  if location == nil then
    location = os.getenv('ENV_OSS_ENDPOINT_LOCATION') or 'local'
    globals:set('ENV_OSS_ENDPOINT_LOCATION', location)
  end

  return location
end

function _M.get_aws_access_key_id()
  local ak, _ = globals:get('AWS_ACCESS_KEY_ID')
  if ak == nil then
    ak = os.getenv('AWS_ACCESS_KEY_ID') or ''
    globals:set('AWS_ACCESS_KEY_ID', ak)
  end

  return ak
end

function _M.get_aws_secret_access_key()
  local sk, _ = globals:get('AWS_SECRET_ACCESS_KEY')
  if sk == nil then
    sk = os.getenv('AWS_SECRET_ACCESS_KEY') or ''
    globals:set('AWS_SECRET_ACCESS_KEY', sk)
  end

  return sk
end

return _M
