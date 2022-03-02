# Npool go service app template

[![Test](https://github.com/NpoolPlatform/sphinx-plugin/actions/workflows/main.yml/badge.svg?branch=master)](https://github.com/NpoolPlatform/sphinx-plugin/actions/workflows/main.yml)

[目录](#目录)
- [Npool go service app template](#npool-go-service-app-template)
    - [功能](#功能)
    - [命令](#命令)
    - [最佳实践](#最佳实践)
    - [环境变量](#环境变量)
    - [部署](#部署)
    - [升级说明](#升级说明)
    - [推荐](#推荐)
    - [说明](#说明)

-----------
### 功能
* [x] 将服务部署到k8s集群
* [x] 将服务api通过traefik-internet ingress代理，供外部应用调用(视服务功能决定是否需要)

### 命令
* make init ```初始化仓库，创建go.mod```
* make verify ```验证开发环境与构建环境，检查code conduct```
* make verify-build ```编译目标```
* make test ```单元测试```
* make generate-docker-images ```生成docker镜像```
* make sphinx-plugin ```单独编译服务```
* make sphinx-plugin-image ```单独生成服务镜像```
* make deploy-to-k8s-cluster ```部署到k8s集群```

### 最佳实践
* 每个服务只提供单一可执行文件，有利于docker镜像打包与k8s部署管理
* 每个服务提供http调试接口，通过curl获取调试信息
* 集群内服务间direct call调用通过服务发现获取目标地址进行调用

### 环境变量

|  币种  | 变量名称       |       支持的值        | 说明  |
| :----: | :------------- | :-------------------: | :---: |
| Common | ENV_COIN_NET   |     main or test      |       |
|        | ENV_COIN_TYPE  | FIL BTC ETH SpaceMesh |       |
|        | ENV_COIN_API   |        ip:port        |       |
|  FIL   | ENV_COIN_TOKEN |                       |       |
|  BTC   | ENV_COIN_USER  |                       |       |
|        | ENV_COIN_PASS  |                       |       |

1. **ENV_COIN_API** 钱包服务的 **ipv4** 或者 **ipv6** 地址
2. **ENV_COIN_TOKEN** 钱包服务的 **token**

### 部署

```
# mkdir -p /etc/SphinxPlugin
# cp cmd/sphinx-plugin/SphinxPlugin.viper.yaml /etc/SphinxPlugin/
# cat /etc/SphinxPlugin/SphinxPlugin.viper.yaml
---
config:
  hostname: "sphinx-plugin.npool.top"
  http_port: 50170
  grpc_port: 50171
  prometheus_port: 50172
  appid: "89089012783789789719823798127398"
  logdir: "/var/log"
  apolloAccessKey: "0147fb70b815403790e8634b899fac07"
  sphinx_proxy_addr: "sphinx.proxy.api.npool.top:8080,sphinx.proxy.api.xpool.top:8080"


# cat /etc/systemd/system/sphinx-plugin.service
[Unit]
Description=Sphinx Plugin
After=network.target

[Service]
# FIL
Environment="ENV_COIN_NET=test"
Environment="ENV_COIN_TYPE=FIL"
Environment="ENV_COIN_API=$wallet-ip:1234"
# BTC test
# Environment="ENV_COIN_API=127.0.0.1:18443"
Environment="ENV_COIN_TOKEN=$wallet-api"
ExecStart=/opt/sphinx-plugin/sphinx-plugin run
ExecStop=/bin/kill -s QUIT $MAINPID
# resource limit
CPUQuota=200%
MemoryMax=128M
Restart=always
RestartSec=30
TimeoutSec=infinity
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

### 升级说明

+ **需要关闭用户购买商品的入口**
+ **失败可以重试, 成功操作不可重试**
+ **注意 SQL 只更新了 FIL 和 BTC 币种，其余可参考 FIL 和 BTC，必须要等 tfilecoin 和 tbitcoin 上报完成才可以执行**

| 条件    | 升级 SQL                     |
| :------ | :--------------------------- |
| mainnet | DO NOTHING                   |
| testnet | [upgrade](./sql/upgrade.sql) |

### 推荐
BTC 钱包节点的配置文件中, **rpcclienttimeout=30** 需要配置

### 说明

+ 不支持 **Windows**
