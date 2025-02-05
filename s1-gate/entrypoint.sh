#!/bin/bash

# render nginx.conf using environment variables
cd /usr/local/openresty/nginx/conf
cat nginx.template.conf | gomplate > nginx.conf

exec "$@"
