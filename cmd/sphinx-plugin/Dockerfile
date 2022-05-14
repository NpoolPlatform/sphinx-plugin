FROM jrei/systemd-centos:7

COPY sphinx-plugin /usr/local/bin/sphinx-plugin

ENV PROXY
ENV CONTRACT
ENV LOG /var/log
ENV LEVEL debug

ENTRYPOINT ["sphinx-plugin"]
CMD ["help"]
