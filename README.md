# Blockchain Monitor

本程序主要用途有三个：

1. 自动创建部署以太坊智能合约
2. 监控以太坊event log，确认充值、提现以及ERC20代币充值和提现，确认合约部署情况
3. 打款



**使用步骤：**

1. 初始化。准备好若干个以太坊账号，分别标定它们的角色。本系统中将存在两种角色——合约创建人和资产提取人。合约创建人负责创建合约，资产提取人负责收取用户充值的数字货币。
2. 准备好合约创建人的keystore，并向其中充值一定以太用以支付创建合约的费用。
3. 用本程序加密keystore密码，将第一步和本步骤产生的输出写入到config.json配置文件中对应的参数中。注意配置文件中保存的是加密过后的keystore密码！系统启动时会要求操作者输入密码来解密keystore密码！
4. 在以太坊上使用合约创建人的账号手动来创建钱包工厂合约。该钱包工厂合约将用于自动化创建和部署用户智能合约。
5. 初始化系统运作所需数据库表。
6. 用本程序预生成一批智能合约地址。这些地址将用来提供给申请充值地址的交易所用户。
7. 启动本程序。



**命令行用法：**

```bash
➜ ./bcmonitor help
NAME:
   Blockchain monitor - The blockchain monitor command line interface

USAGE:
   bcmonitor [global options] command [command options] [arguments...]

VERSION:
   dev-20170816

DESCRIPTION:
   blockchain monitor

AUTHOR:
   BitSE Group <support@2se.com>

COMMANDS:
     encrypt      encrypt the wallet creator`s passphrase
     decrypt      decrypt the wallet creator`s passphrase
     pregenerate  pre generate the contract address
     deploy       deploy wallet factory
     start        start the manager
     stop         stop the manager
     help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version

COPYRIGHT:
   Copyright 2017-2018 The exchange Authors
```



**功能说明：**

＊ 加密keystore密钥

＊ 预生成智能合约地址

＊ 部署钱包工厂智能合约（也可手工在以太坊节点上部署）

＊ 故障恢复

＊ 根据配置设定出块数据确认，以防止因分叉导致确认数目出错。

＊ 充值提现监控

＊ 智能合约创建成功监控