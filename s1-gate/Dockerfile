FROM openresty/openresty:bionic

# system dependencies
RUN mkdir -p /app

# luarocks and opm dependencies
COPY install.sh /app/install.sh
COPY Makefile /app/Makefile
COPY conf /usr/local/openresty/nginx/conf
COPY lua /usr/local/openresty/nginx/lua

RUN cd /app && make install-dev

CMD ["/usr/local/openresty/bin/openresty", "-g", "daemon off;"]

STOPSIGNAL SIGQUIT
