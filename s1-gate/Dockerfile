FROM openresty/openresty:bionic

# system dependencies
RUN mkdir -p /app

COPY --from=hairyhenderson/gomplate:stable /gomplate /bin/gomplate

# luarocks and opm dependencies
COPY install.sh /app/install.sh
COPY Makefile /app/Makefile
COPY conf /usr/local/openresty/nginx/conf
COPY lua /usr/local/openresty/nginx/lua
COPY entrypoint.sh /app/entrypoint.sh

RUN cd /app && make install-dev && chmod +x /app/entrypoint.sh

ENTRYPOINT [ "/app/entrypoint.sh" ]
CMD ["/usr/local/openresty/bin/openresty", "-g", "daemon off;"]

STOPSIGNAL SIGQUIT
