# Companion

[![Build Status](https://travis-ci.org/boxproject/companion.svg?branch=master)](https://travis-ci.org/boxproject/companion) [![Hex.pm](https://img.shields.io/hexpm/l/plug.svg)](https://www.apache.org/licenses/LICENSE-2.0) [![language](https://img.shields.io/badge/golang-%5E1.10-blue.svg)]()

本程序主要用途有三个：

1. 与代理服务通信
2、多个节点一定比例（例如2／3，3个审核节点中2个节点）确认，才认可本次提交的数据正确性。
3. 监控以太坊私链event log，确认审批流操作，以及转账审批等


**使用步骤：**

1. 初始化。创建好oracle、sink智能合约。
2. 准备好合约授权人的keystore，并向其中充值一定以太用以支付创建合约的费用。
3. 分别对每个节点授权人进行授权（对oracle智能合约进行操作）
4. 用本程序加密keystore密码，将第一步和本步骤产生的输出写入到config.json配置文件中对应的参数中。注意配置文件中保存的是加密过后的keystore密码！系统启动时会要求操作者输入密码来解密keystore密码！
5. 将连接代理的地址、端口以及ssl公钥以及证书
6. 启动本程序。



**命令行用法：**

```bash
➜ ./companion help
NAME:
   Blockchain companion - The blockchain companion command line interface

USAGE:
   companion [global options] command [command options] [arguments...]

VERSION:
   dev-20180424

DESCRIPTION:
   companion handler

AUTHOR:
   box Group <support@2se.com>

COMMANDS:
     start        start the manager
     stop         stop the manager
     help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version

COPYRIGHT:
   Copyright 2017-2018 The box Authors
```



**功能说明：**


＊ 故障恢复

＊ 根据配置设定出块数据确认，以防止因分叉导致确认数目出错。

＊ 审批流创建及确认监控

＊ 转账审批监控