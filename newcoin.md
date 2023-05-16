# 币种交易

## 交易流程标准

基础检查：

- 节点可用性（检查高度、节点健康状态）
- From To 地址合法性
- Value 合法性（精度、能否解析）

交易流程：

- Gas + Value < Balance
- 链上估计gasPrice

## 其他功能

estimateGas 目前只有eth链上的支持