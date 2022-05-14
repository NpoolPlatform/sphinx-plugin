FROM jrei/systemd-centos:7

COPY sphinx-plugin /usr/local/bin/sphinx-plugin

ENV PROXY='' \
    CONTRACT='' \
    LOG=/var/log \
    LEVEL=debug \
    ENV_COIN_NET='' \
    ENV_COIN_TYPE='' \
    ENV_COIN_API='' \
    ENV_COIN_TOKEN='' \
    ENV_COIN_USER='' \
    ENV_COIN_PASS=''

ENTRYPOINT ["sphinx-plugin"]
CMD ["help"]
