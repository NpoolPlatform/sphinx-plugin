FROM jrei/systemd-centos:7

COPY sphinxplugin.viper.yaml /etc/sphinxplugin/sphinxplugin.viper.yaml
COPY sphinx-plugin /usr/local/bin/sphinx-plugin

CMD ["sphinx-plugin", "run"]
