package comm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"time"
	"github.com/boxproject/companion/db"
)

//default
const (
	DEF_CURSOR_FILE_PATH = "cursor.txt"
)

//web api
const (
	REQ_HASH_ADD       = "1" //申请
	REQ_HASH_ENABLE    = "2" //确认
	REQ_HASH_DISABLE   = "3" //禁用
	REQ_HASH_AVAILABLE = "4" //是否生效

	REQ_OUT_APPROVE = "5" //提现申请
	REQ_OUT_EXISTS  = "6" //提现是否成功
)

const (
	FAILED  = "0" //失败
	SUCCESS = "1" //成功
)

const (
	HASH_PRIFIX        = "0x"
	HASH_ENABLE_LENGTH = common.HashLength
)

const (
	Err_OK                = "0"   //正确
	Err_UNKNOW_REQ_TYPE   = "10"  //未知请求类型
	Err_ETH               = "100" //链处理失败
	Err_UNENABLE_PREFIX   = "101" //非法hash前缀
	Err_UNENABLE_LENGTH   = "102" //非法hash值长度
	Err_UNENABLE_AMOUNT   = "103" //非法金额
	Err_HASH_EXSITS       = "104" //hash已确认
	Err_UNENABLE_CATEGORY = "105" //非法转账类型
)

//db key
const (
	HASH_ADD_PREFIX         = "ha_"
	HASH_ADD_CONTENT_PREFIX = "hac_"
	APPROVE_RECADDR_PREFIX  = "apr_" //recAddress
	HASH_ENABLE_PREFIX      = "he_"
	HASH_DISABLE_PREFIX     = "hd_"
	WITHDRAW_APPLY_PREFIX   = "wa_"
)

//转账类型区间
const (
	MIN_CATEGORY int64 = 0
	MAX_CATEGORY int64 = 500
)

const (
	CHAN_MAX_SIZE = 100000
)

const (
	HASH_STATUS_APPLY   = "1" //申请
	HASH_STATUS_ENABLE  = "2" //确认
	HASH_STATUS_DISABLE = "3" //禁用
)

//
const (
	CATEGORY_BTC int64 = 0
	CATEGORY_ETH int64 = 1
	CATEGORY_BOX int64 = 2
)

//grpc流类型
const (
	GRPC_SIGN_ADD     = "1"	//新加hash
	GRPC_SIGN_ENABLE  = "2"	//同意
	GRPC_SIGN_DISABLE = "3"	//禁用
	GRPC_ACCOUNT_USE  = "4"	//账户使用
	GRPC_DEPOSIT  	   = "5"	//充值上报
	GRPC_APPROVE      = "6"	//提现
	GRPC_WITHDRAW  	   = "7"	//提现上报
	GRPC_WITHDRAW_TX  = "8"	//提现tx上报
)
const (
	//grpc_0_TYPE_hash 发送失败
	//grpc_1_TYPE_hash 发送成功
	GRPC_DB_PREFIX = "grpc_"
)
//上报类型
const (
	REQ_ACCOUNT_ADD = "1" //账户上报
	REQ_DEPOSIT     = "2" //充值上报
	REQ_WITHDRAW    = "3" //提现上报
	REQ_WITHDRAW_TX = "4" //提现tx上报
)

//请求包装类
type RequestModel struct {
	Hash       string
	Approver   string //审批人
	ReqType    string
	WdHash     string
	Amount     string
	Fee        string
	Category   int64
	RecAddress string
	Content    string //hash内容
}

type GrpcStream struct {
	Type        string
	BlockNumber uint64 //区块号
	Approver    string //审批人
	Hash        common.Hash
	WdHash      common.Hash
	Amount      *big.Int
	Fee         *big.Int
	Account     string
	To          string
	Category    *big.Int
	Content     string
	Status      string
	CreateTime  time.Time //创建时间
}

//请求数据
type VReq struct {
	ReqType  string
	Account  string
	From     string
	To       string
	Category int64
	Amount   string
	WdHash   string
	TxHash   string
}

type VRsp struct {
	Code    int
	Message string
}

//request chan
var ReqChan chan *RequestModel = make(chan *RequestModel, CHAN_MAX_SIZE)

//grpc 流数据
var GrpcStreamChan chan *GrpcStream = make(chan *GrpcStream, CHAN_MAX_SIZE)

//请求channel
var VReqChan chan *VReq = make(chan *VReq, CHAN_MAX_SIZE)

var Ldb *db.Ldb