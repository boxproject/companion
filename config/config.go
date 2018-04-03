package config

type Config struct {
	PriEthCfg   EthCfg     `json:"pri_eth,omitempty"`
	RouterInfo  RouterInfo `json:"router_info,omitempty"`
	LevelDbPath string     `json:"level_db_path,omitempty"`
	SinkAddress string     `json:"sink_address,omitempty"`
	ServerCert  string     `json:"server_cert,omitempty"`
	ServerKey   string     `json:"server_key,omitempty"`
	ClientCert  string     `json:"client_cert,omitempty"`
	ClientKey   string     `json:"client_key,omitempty"`
	GrpcSerHost string     `json:"grpc_ser_host,omitempty"`
	GrpcSerPort string     `json:"grpc_ser_port,omitempty"`
	AccountUrl  string         `json:"account_url,omitempty"`
	DepositUrl    string `json:"deposit_url,omitempty"`
	WithDrawUrl   string `json:"withdraw_url,omitempty"`
	WithDrawTxUrl string `json:"withdraw_tx_url,omitempty"`
}

type EthCfg struct {
	CashierSetter       string `json:"cashier_setter"`        // CashierSetter 统一设置合约的转账地址
	Creator             string `json:"creator"`               // Creator 创建者的地址
	CreatorPassphrase   string `json:"creator_passphrase"`    // Creator 创建者的keystore密钥
	CreatorKeystorePath string `json:"creator_keystore_path"` // Creator 创建者keystore 路径
	GethAPI             string `json:"geth_api"`              // GethAPI 以太坊接口地址，要支持websocket
	CheckBlockBefore    int64  `json:"check_block_before"`    // CheckBlockBefore 设置当前块向前推若干个块做校验
	CursorFilePath      string `json:"cursor_file_path"`      // CursorFilePath 设置当前块处理游标
	GasLimit            int64  `json:"gas_limit"`             //执行方法gaslimit
	GasPrice            int64  `json:"gas_price"`             //执行gasprice
	WalletGas           int    `json:"wallet_gas"`            // 部署wallet所需gas
	FactoryGas          int    `json:"factory_gas"`           // 部署wallet factory 所需gas
}

type HttpServer struct {
	HttpBind         string `json:"http_bind"`
	HttpReadTimeOut  string `json:"http_read_timeout"`
	HttpWriteTimeOut string `json:"http_write_timeout"`
}

type RouterInfo struct {
	SerVoucher          	string `json:"ser_voucher"`
	SerCompanion  			string `json:"ser_companion"`
	CompanionName  			string `json:"companion_name"`
}
