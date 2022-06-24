# Npool go service app template

[![Test](https://github.com/NpoolPlatform/sphinx-plugin/actions/workflows/main.yml/badge.svg?branch=master)](https://github.com/NpoolPlatform/sphinx-plugin/actions/workflows/main.yml)

[目录](#目录)

- [Npool go service app template](#npool-go-service-app-template)
  - [功能](#功能)
  - [命令](#命令)
  - [最佳实践](#最佳实践)
  - [环境变量](#环境变量)
  - [新增币种的开发步骤](#新增币种的开发步骤)
  - [ethereum 部署](#ethereum-部署)
  - [solana 部署](#solana-部署)
  - [部署](#部署)
  - [升级说明](#升级说明)
  - [推荐](#推荐)
  - [说明](#说明)
  - [优化](#优化)

-----------

## 功能

- [x] 将服务部署到k8s集群
- [x] 将服务api通过traefik-internet ingress代理，供外部应用调用(视服务功能决定是否需要)

## 命令

- make init ```初始化仓库，创建go.mod```
- make verify ```验证开发环境与构建环境，检查code conduct```
- make verify-build ```编译目标```
- make test ```单元测试```
- make generate-docker-images ```生成docker镜像```
- make sphinx-plugin ```单独编译服务```
- make sphinx-plugin-image ```单独生成服务镜像```
- make deploy-to-k8s-cluster ```部署到k8s集群```

## 最佳实践

- 每个服务只提供单一可执行文件，有利于docker镜像打包与k8s部署管理
- 每个服务提供http调试接口，通过curl获取调试信息
- 集群内服务间direct call调用通过服务发现获取目标地址进行调用

## 环境变量

| 币种                               | 变量名称                        | 支持的值                                                    | 说明                                            |
|:-------------------------------- |:--------------------------- |:------------------------------------------------------- |:--------------------------------------------- |
| common                           | ENV_COIN_NET                | main or test                                            |                                               |
|                                  | ENV_COIN_TYPE               | filecoin bitcoin ethereum/usdterc20 spacemesh usdttrc20 | 如果此**plugin**支持多币种使用,分割                       |
| ~~fil btc sol~~                  | ~~ENV_COIN_API~~            | ~~ip:port~~                                             | 已经废弃，使用ENV_COIN_LOCAL_API及ENV_COIN_PUBLIC_API |
| fil btc sol eth/erc20 tron/trc20 | ENV_COIN_LOCAL_API          | ip:port                                                 | 多个地址使用,分割                                     |
| fil btc sol eth/erc20 tron/trc20 | ENV_COIN_PUBLIC_API         | ip:port                                                 | 多个地址使用,分割                                     |
| tron/trc20                       | ENV_COIN_JSONRPC_LOCAL_API  | ip:port                                                 | 多个地址使用,分割                                     |
| tron/trc20                       | ENV_COIN_JSONRPC_PUBLIC_API | ip:port                                                 | 多个地址使用,分割                                     |
| ethereum/usdterc20               |                             |                                                         |                                               |
| filecoin                         | ENV_COIN_TOKEN              |                                                         |                                               |
| bitcoin                          | ENV_COIN_USER               |                                                         |                                               |
|                                  | ENV_COIN_PASS               |                                                         |                                               |
| usdttrc20                        | ENV_CONTRACT                |                                                         | 填写trc20的合约地址                                  |

1. **ENV_COIN_LOCAL_API/ENV_COIN_PUBLIC_API** 钱包服务的 **ipv4** 、 **ipv6** 地址或是域名
2. **ENV_COIN_TOKEN** 钱包服务的 **token**

## Plugin Features

由于币种之间差异，会造成plugin在每个币种的功能存在差异

此部分记录当前版本plugin特性对各币种的支持情况

### multiple-endpoints

每个plugin可配置多个节点地址（可提供账户余额、交易状态、链状态等相关API的地址）

相关环境变量：ENV_COIN_LOCAL_API、ENV_COIN_PUBLIC_API

eth/erc20 tron/trc20 bsc/bep20 支持配置多个地址并使用","分割

fil btc sol 只支持配置ENV_COIN_LOCAL_API环境变量，且仅支持配置单节点

### wallet-status-check

钱包状态检查，目前仅检查节点高度是否与链高度一致

eth/erc20 bsc/bep20 仅在查询balance时验证区块高度

tron/trc20 在连接节点时检测区块高度，需要节点开启json-api端口才可检测，目前仅检测ENV_COIN_LOCAL_API，信任ENV_COIN_PUBLIC_API配置的节点地址

其他币种暂无

### account-check

账户验证

tron/trc20 在获取balance时检测账户格式，与波场HTTP-API提供的wallet/validateaddress功能一致

其他币种暂无

### [新增币种的开发步骤](./newcoin.md)

1. 必须要实现的接口

2. 注册新币种

### ethereum 部署

1. 启动测试网

2. 部署智能合约

   1. 部署合约

      ```sh
      sphinx-plugin usdterc20 -addr 127.0.0.1 -port 8545
      ```

   2. 上述的命令会返回合约的**ID**,设置到配置文件**ENV_CONTRACT**的值

   3. 部署支持 ethereum/usdterc20 的 plugin

### TRC20部署

环境变量示例（例子中为波场neil的测试环境）

```sh
export ENV_COIN_NET=test  # main | test
export ENV_COIN_TYPE=usdttrc20  # usdttrc20
export ENV_CONTRACT=TXLAQ63Xg1NAzckPwKHvzw7CSEmLMEqcdj
export ENV_COIN_PUBLIC_API=grpc.nile.trongrid.io:50051,grpc.nile.trongrid.io:50051 # 提供grpc-api的地址
export ENV_COIN_LOCAL_API=47.252.19.181:50051
export ENV_COIN_JSONRPC_LOCAL_API=47.252.19.181:50545
export ENV_COIN_JSONRPC_PUBLIC_API=
export ENV_PROXY='10.107.172.251:50001'
export ENV_LOG_DIR=/var/log
export ENV_LOG_LEVEL=debug
```

主网trc20合约

TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t

资源文档（官方endpoint、水龙头、测试网信息等）

<https://cn.developers.tron.network/docs/networks>

### solana 部署

这个示例中用编译好的二进制文件直接跑，如果用docker或者systemd跑设置好相应环境变量即可

准备配置文件，配置文件可能需要配置proxy地址

```sh
mkdir -p /etc/SphinxPlugin
cp cmd/sphinx-plugin/SphinxPlugin.viper.yaml /etc/SphinxPlugin/
cat /etc/SphinxPlugin/SphinxPlugin.viper.yaml
```

设置环境变量
plugin-solana 环境变量
可以设置到systemctl的service文件中

```sh
export ENV_COIN_NET=test
export ENV_COIN_TYPE=solana
# 开发网
export ENV_COIN_LOCAL_API=https://api.devnet.solana.com
# 测试网
#export ENV_COIN_LOCAL_API=https://api.testnet.solana.com
# 主网
#export ENV_COIN_LOCAL_API=https://api.mainnet-beta.solana.com
```

运行plugin

```sh
/opt/sphinx-plugin/sphinx-plugin run
```

### 部署

```conf
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
```

### 升级说明

- **需要关闭用户购买商品的入口**
- **失败可以重试, 成功操作不可重试**
- **注意 SQL 只更新了 filecoin 和 bitcoin 币种，其余可参考 filecoin 和 bitcoin, tfilecoin 和 tbitcoin 上报完成才可以执行**

| 条件      | 升级 SQL                       |
|:------- |:---------------------------- |
| mainnet | DO NOTHING                   |
| testnet | [upgrade](./sql/upgrade.sql) |

### 推荐

bitcoin 钱包节点的配置文件中, **rpcclienttimeout=30** 需要配置

### 说明

- 不支持 **Windows**

## 优化

- 镜像多阶段构建
- 尝试关闭 **CGO_ENABLE**
