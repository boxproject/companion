package controllers

import (
	"strings"

	logger "github.com/alecthomas/log4go"
	"github.com/boxproject/companion/comm"
	"github.com/boxproject/companion/handler"
	"github.com/boxproject/companion/util"
	"github.com/ethereum/go-ethereum/common"
)

type HashController struct {
	baseController
}

//Hash模型
type HashResultModel struct {
	RspNo  string //0-成功 其他-失败
	Result bool   //结果
}

//err json pkg
func (h *HashController) retErrJSON(hash, errNo string) {
	h.Data["json"] = &HashResultModel{RspNo: errNo}
	h.ServeJSON()
}

//Hash处理
func (h *HashController) Hash() {
	hashStr := h.GetString("hash")
	if !common.IsHexAddress(hashStr) {
		h.retErrJSON(hashStr, comm.Err_UNENABLE_PREFIX)
		return
	}

	if len(common.FromHex(hashStr)) != comm.HASH_ENABLE_LENGTH {
		h.retErrJSON(hashStr, comm.Err_UNENABLE_LENGTH)
		return
	}
	reqtype := h.GetString("reqtype")
	switch reqtype {
	case comm.REQ_HASH_ADD:
		h.addHash()
	case comm.REQ_HASH_AVAILABLE:
		h.availableHash()
	default:
		logger.Error("unknow request type: %s", reqtype)
		h.retErrJSON(hashStr, comm.Err_UNKNOW_REQ_TYPE)
		return
	}
}

//add hash asy
func (h *HashController) addHash() {
	logger.Debug("HashController addHash...")
	hash := h.GetString("hash")
	approver := h.GetString("approver") //审批人
	content := h.GetString("content")   //内容
	logger.Debug("content....", content)
	hashModel := &HashResultModel{RspNo: comm.Err_OK, Result: true}
	if b, err := handler.PriSynEth.HashAvailable(hash); err != nil {
		logger.Error("handler failed: %s", err)
		h.retErrJSON(hash, comm.Err_ETH)
		return
	} else if b {
		h.retErrJSON(hash, comm.Err_HASH_EXSITS)
		return
	}
	h.Data["json"] = hashModel
	comm.ReqChan <- &comm.RequestModel{Hash: hash, ReqType: comm.REQ_HASH_ADD, Approver: approver, Content: content}
	h.ServeJSON()
}

//avaiable hash syn
func (h *HashController) availableHash() {
	logger.Debug("HashController availableHash...")
	hash := h.GetString("hash")
	hashModel := &HashResultModel{RspNo: comm.Err_OK, Result: false}
	b, err := handler.PriSynEth.HashAvailable(hash)
	if err != nil {
		logger.Error("handler failed: %s", err)
	} else {
		hashModel.Result = b
	}
	h.Data["json"] = hashModel
	h.ServeJSON()
}

//提现申请
type ApplyController struct {
	baseController
	synHandler *handler.PriSynEthHandler
}

//提现模型
type ApplyModel struct {
	RspNo   string //0-成功 其他-失败
	RspDesc string //说明
	Hash    string //hash值
	WdHash  string //txHash
	Result  bool   //结果
}

//err json pkg
func (a *ApplyController) retErrJSON(hash, txHash, errNo string) {
	a.Data["json"] = &ApplyModel{RspNo: errNo, Hash: hash, WdHash: txHash}
	a.ServeJSON()
}

//提现申请
func (a *ApplyController) Apply() {
	hashStr := a.GetString("hash")
	wdHashStr := a.GetString("wdhash")
	if !strings.HasPrefix(hashStr, comm.HASH_PRIFIX) || !strings.HasPrefix(wdHashStr, comm.HASH_PRIFIX) {
		a.retErrJSON(hashStr, wdHashStr, comm.Err_UNENABLE_PREFIX)
		return
	}
	if len(common.FromHex(hashStr)) != comm.HASH_ENABLE_LENGTH || len(common.FromHex(wdHashStr)) != comm.HASH_ENABLE_LENGTH {
		a.retErrJSON(hashStr, wdHashStr, comm.Err_UNENABLE_LENGTH)
		return
	}
	reqtype := a.GetString("reqtype")
	switch reqtype {
	case comm.REQ_OUT_APPROVE:
		a.approve()
	case comm.REQ_OUT_EXISTS:
		a.txExists()
	default:
		logger.Error("unknow apply type: %s", reqtype)
		a.retErrJSON(hashStr, wdHashStr, comm.Err_UNKNOW_REQ_TYPE)
	}
}

//approve asy
func (a *ApplyController) approve() {
	logger.Debug("ApplyController approve...")
	hash := a.GetString("hash")
	wdHash := a.GetString("wdhash")
	recAddress := a.GetString("recaddress")
	amount := a.GetString("amount")

	fee := a.GetString("fee")

	category, err := a.GetInt64("category")
	if err != nil {
		logger.Debug("category[%d] illegal", category)
		a.retErrJSON(hash, wdHash, comm.Err_UNENABLE_AMOUNT)
		return
	} else if !util.CheckCategory(category) {
		a.retErrJSON(hash, wdHash, comm.Err_UNENABLE_CATEGORY)
		return
	}
	logger.Debug("ApplyController.approve:---", "hash:", hash, " wdhash:", wdHash, " recAddress:", recAddress, " amount:", amount, " fee:", fee, " category:", category)

	a.Data["json"] = &ApplyModel{RspNo: comm.Err_OK, Hash: hash, WdHash: wdHash}
	comm.ReqChan <- &comm.RequestModel{Hash: hash, ReqType: comm.REQ_OUT_APPROVE, WdHash: wdHash, RecAddress: recAddress, Amount: amount, Fee: fee, Category: category}
	a.ServeJSON()
}

//approve是否生效
func (a *ApplyController) txExists() {
	logger.Debug("ApplyController txExists...")
	hash := a.GetString("hash")
	wdHash := a.GetString("wdhash")
	applyModel := &ApplyModel{RspNo: comm.Err_OK, Hash: hash, WdHash: wdHash, Result: false}
	if b, err := a.synHandler.TxExists(hash, wdHash); err != nil {
		logger.Error("handler failed:%s", err)
	} else if b {
		applyModel.Result = b
	}
	a.Data["json"] = applyModel
	a.ServeJSON()
}


//account used
type AccountModel struct {
	RspNo   string //0-成功 其他-失败
	RspDesc string //说明
}
