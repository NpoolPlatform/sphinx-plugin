# Npool go service app template

[![Test](https://github.com/NpoolPlatform/sphinx-plugin/actions/workflows/main.yml/badge.svg?branch=master)](https://github.com/NpoolPlatform/sphinx-plugin/actions/workflows/main.yml)

[目录](#目录)

- [Npool go service app template](#npool-go-service-app-template)
  - [新增币种](#新增币种)
    - [新增功能](#新增功能)
  - [功能](#功能)
  - [命令](#命令)
  - [最佳实践](#最佳实践)
  - [环境变量](#环境变量)
    - [wallet-status-check](#wallet-status-check)
    - [account-check](#account-check)
    - [ethereum 部署](#ethereum-部署)
    - [升级说明](#升级说明)
    - [推荐](#推荐)
    - [说明](#说明)
  - [优化](#优化)

-----------

## [新增币种](./newcoin.md)

### 新增功能

- [x] 自定义调度周期
- [x] 自定义错误处理
- [ ] 链路追踪
- [ ] 监控
- [ ] 上报meta信息到proxy
- [ ] 优化配置
- [ ] 现在相同地址的并发处理
- [ ] payload 同步到 redis
- [ ] 动态调整 **gas fee**
- [ ] 支持多 **pod** 部署

新币种的支持步骤

1. 配置新币种单位和名称
2. 必须要实现的接口
3. 注册新币种
4. 设置默认SyncTime

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

| 币种  | 变量名称            | 支持的值     | 说明                             |
|:------|:--------------------|:-------------|:---------------------------------|
| comm  | ENV_COIN_NET        | main or test |                                  |
|       | ENV_COIN_TYPE       |              |                                  |
|       | ENV_SYNC_INTERVAL   |              | optional,交易状态同步间隔周期(s) |
|       | ENV_WAN_IP          |              | plugin的wan-ip                   |
|       | ENV_POSITION        |              | plugin的位置信息(如NewYork_NO2)  |
|       | ENV_COIN_LOCAL_API  |              | 多个地址使用,分割                |
|       | ENV_COIN_PUBLIC_API |              | 多个地址使用,分割                |
| erc20 | ENV_CONTRACT        |              | 合约币种的合约地址               |

配置说明

对于合约地址配置说明

钱包地址配置格式:
  **url|auth,url|auth,url|auth**

- 不需要认证

  ````conf
    auth 格式
    示例: https://127.0.0.1:8080|
  ````

- 账号密码体系

  ````conf
    auth 格式
    user@password
    示例: https://127.0.0.1:8080|root@3306
  ```

- token 体系

  ````conf
    auth 格式
    token
    示例: https://127.0.0.1:8080|token
  ```

交易上链状态查询默认周期

|          币种          | 默认值 | 出块时间 |
|:----------------------:|:------:|:--------:|
|        filecoin        |  20s   |   30s    |
|        bitcoin         |  7min  |  10min   |
|         solana         |   1s   |   0.4s   |
|   ethereum/usdterc20   |  12s   |  10~20s  |
| binancecoin/binanceusd |   4s   |    5s    |
|     tron/usdttrc20     |   2s   |    3s    |

### wallet-status-check

钱包状态检查，目前仅检查节点高度是否与链高度一致

eth/erc20 bsc/bep20 仅在查询balance时验证区块高度

tron/trc20 在连接节点时检测区块高度，需要节点开启json-api端口才可检测，目前仅检测ENV_COIN_LOCAL_API，信任ENV_COIN_PUBLIC_API配置的节点地址

其他币种暂无

### account-check

账户验证

tron/trc20 在获取balance时检测账户格式，与波场HTTP-API提供的wallet/validateaddress功能一致

其他币种暂无

### ethereum 部署

1. 启动测试网

2. 部署智能合约

   1. 部署合约

      ```sh
      sphinx-plugin usdterc20 -addr 127.0.0.1 -port 8545
      ```

   2. 上述的命令会返回合约的**ID**,设置到配置文件**ENV_CONTRACT**的值

   3. 部署支持 ethereum/usdterc20 的 plugin

### 升级说明

- **需要关闭用户购买商品的入口**
- **失败可以重试, 成功操作不可重试**
- **注意 SQL 只更新了 filecoin 和 bitcoin 币种，其余可参考 filecoin 和 bitcoin, tfilecoin 和 tbitcoin 上报完成才可以执行**

| 条件    | 升级 SQL                     |
|:--------|:-----------------------------|
| mainnet | DO NOTHING                   |
| testnet | [upgrade](./sql/upgrade.sql) |

### 推荐

bitcoin 钱包节点的配置文件中, **rpcclienttimeout=30** 需要配置

### 说明

- 不支持 **Windows**

## 优化

- 镜像多阶段构建
- 尝试关闭 **CGO_ENABLE**
