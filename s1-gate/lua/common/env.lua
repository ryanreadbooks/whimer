local _M = {}

local globals = ngx.shared.globals

function _M.get_oss_endpoint_host()
  local host, _ = globals:get('ENV_OSS_ENDPOINT_HOST')
  if host == nil then
    host = os.getenv('ENV_OSS_ENDPOINT_HOST')
    if host == nil then
      host = '127.0.0.1'
      ngx.log(ngx.WARN, 'ENV_OSS_ENDPOINT_HOST unset')
    end
    globals:set('ENV_OSS_ENDPOINT_HOST', host)
  end

  return host
end

function _M.get_oss_endpoint_port()
  local port, _ = globals:get('ENV_OSS_ENDPOINT_PORT')
  if port == nil then
    port = os.getenv('ENV_OSS_ENDPOINT_PORT')
    if port == nil then
      port = 9000
      ngx.log(ngx.WARN, 'ENV_OSS_ENDPOINT_PORT unset')
    end
    globals:set('ENV_OSS_ENDPOINT_PORT', port)
  end

  return port
end

function _M.get_oss_endpoint_location()
  local location, _ = globals:get('ENV_OSS_ENDPOINT_LOCATION')
  if location == nil then
    location = os.getenv('ENV_OSS_ENDPOINT_LOCATION')
    if location == nil then
      location = 'local'
      ngx.log(ngx.WARN, 'ENV_OSS_ENDPOINT_LOCATION unset')
    end
    globals:set('ENV_OSS_ENDPOINT_LOCATION', location)
  end

  return location
end

function _M.get_aws_access_key_id()
  local ak, _ = globals:get('ENV_OSS_AK')
  if ak == nil then
    ak = os.getenv('ENV_OSS_AK')
    if ak == nil then
      ak = ''
      ngx.log(ngx.WARN, 'ENV_OSS_AK unset')
    end
    globals:set('ENV_OSS_AK', ak)
  end

  return ak
end

function _M.get_aws_secret_access_key()
  local sk, _ = globals:get('ENV_OSS_SK')
  if sk == nil then
    sk = os.getenv('ENV_OSS_SK')
    if sk == nil then
      sk = ''
      ngx.log(ngx.WARN, 'ENV_OSS_SK unset')
    end
    globals:set('ENV_OSS_SK', sk)
  end

  return sk
end

return _M
