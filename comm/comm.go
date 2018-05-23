package comm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"time"
	"github.com/boxproject/companion/db"
)

const HASH_PREFIX = "0x"

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

//grpc接口
const (
	GRPC_HASH_ADD_REQ     = "1"  //hash add申请
	GRPC_HASH_ADD_LOG     = "2"  //hans add 私链log
	GRPC_HASH_ENABLE_REQ  = "3"  //hash enable 申请
	GRPC_HASH_ENABLE_LOG  = "4"  //hash enable 私链log
	GRPC_HASH_DISABLE_REQ = "5"  //hash disable 申请
	GRPC_HASH_DISABLE_LOG = "6"  //hash disable 私链log
	GRPC_WITHDRAW_REQ     = "7"  //提现 申请
	GRPC_WITHDRAW_LOG     = "8"  //提现 私链log
	GRPC_DEPOSIT_WEB      = "9"  //充值上报
	GRPC_WITHDRAW_TX_WEB  = "10" //提现tx上报
	GRPC_WITHDRAW_WEB     = "11" //提现结果上报
	GRPC_VOUCHER_OPR_REQ  = "12" //签名机操作处理
	//GRPC_HASH_LIST_REQ    = "13" //审批流查询
	//GRPC_HASH_LIST_WEB    = "14" //审批流上报
	GRPC_TOKEN_LIST_WEB   = "15" //token上报
	GRPC_COIN_LIST_WEB    = "16" //coin上报
	GRPC_HASH_ENABLE_WEB  = "17" //hash enable 公链log
	GRPC_HASH_DISABLE_WEB = "18" //hash enable 公链log
)

const (
	//grpc_0_TYPE_hash 发送失败
	//grpc_1_TYPE_hash 发送成功
	GRPC_DB_PREFIX = "grpc_"
)

const (
	DEF_NONCE  = 0
	NONCE_PLUS = 1
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
	Type           string
	BlockNumber    uint64 //区块号
	AppId          string //申请人
	Hash           common.Hash
	WdHash         common.Hash
	TxHash         string
	Amount         *big.Int
	Fee            *big.Int
	Account        string
	From           string
	To             string
	Category       *big.Int
	Flow           string //原始内容
	Sign           string //签名信息
	Status         string
	VoucherOperate *Operate
	ApplyTime      time.Time //申请时间
	TokenList      []*TokenInfo
	SignInfos      []*SignInfo
}

//私钥-签名机操作
type Operate struct {
	Type         string
	AppId        string //appid
	AppName      string //app别名
	Hash         string
	Password     string
	ReqIpPort    string
	Code         string
	PublicKey    string
	TokenName    string
	Decimals     int64
	ContractAddr string
	CoinCategory int64 //币种分类
	CoinUsed     bool  //币种使用
}

type TokenInfo struct {
	TokenName    string
	Decimals     int64
	ContractAddr string
	Category     int64
}

type SignInfo struct {
	AppId string
	Sign  string
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

type GrpcStreamModel struct {
	Msg []byte
}

//request chan
var ReqChan chan *RequestModel = make(chan *RequestModel, CHAN_MAX_SIZE)

//grpc 流数据
var GrpcStreamChan chan *GrpcStream = make(chan *GrpcStream, CHAN_MAX_SIZE)

//请求channel
var VReqChan chan *VReq = make(chan *VReq, CHAN_MAX_SIZE)

var Ldb *db.Ldb