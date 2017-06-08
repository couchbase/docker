FROM couchbase

RUN mkdir /opt/util
RUN curl -q https://gist.githubusercontent.com/cha87de/56304bbf5f5fe68f65d35aa7a5fe00d9/raw/06009d30c2960979966df689e9f31bde7decdf0e/retry.sh > /opt/util/retry.sh

ADD ./entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
CMD ["couchbase-server"]