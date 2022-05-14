FROM jrei/systemd-centos:7

COPY sphinx-plugin /usr/local/bin/sphinx-plugin

ENV PROXY='' \
    CONTRACT='' \
    LOG=/var/log \
    LEVEL=debug

ENTRYPOINT ["sphinx-plugin"]
CMD ["help"]
