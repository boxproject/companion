package watcher

import (
	"bytes"
	"math"
	"math/big"

	"encoding/json"

	logger "github.com/alecthomas/log4go"
	"github.com/boxproject/companion/comm"
	"github.com/boxproject/companion/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
)

var (
	//公链充值
	pubCoinInEvent = signFunc("Transfer(address,address,uint256)")
	//公链提现
	pubCoinOutEvent            = signFunc("Transfer(address,address,uint256)")
	unit            *big.Float = big.NewFloat(math.Pow(10, 18))

	//私链
	priSignflowAddEvent     = signFunc("SignflowAdded(bytes32,address)")
	priSignflowEnEvent      = signFunc("SignflowEnabled(bytes32,address)")
	priSignflowDisEvent     = signFunc("SignflowDisabled(bytes32,address)")
	priWithdrawAppliedEvent = signFunc("WithdrawApplied(bytes32,bytes32,uint256,uint256,address,uint256,address)")

	PubEventMap = map[common.Hash]EventHandler{
		pubCoinInEvent:  addHashHandler,
		pubCoinOutEvent: addHashHandler,
	}
	PriEventMap = map[common.Hash]EventHandler{
		priSignflowAddEvent:     addHashHandler,
		priSignflowEnEvent:      enableHashHandler,
		priSignflowDisEvent:     disableHashHandler,
		priWithdrawAppliedEvent: withdrawApplyHandler,
	}
)

type EventHandler func(logW *EthEventLogWatcher, log *types.Log) error

func addHashHandler(logW *EthEventLogWatcher, log *types.Log) error {
	logger.Debug("addHashHandler......")

	if dataBytes := log.Data; len(dataBytes) > 0 {
		hash := common.BytesToHash(dataBytes[:32])
		lastConfirmed := common.BytesToAddress(dataBytes[32:64])
		logger.Debug("addHashHandler......db....", hash)
		if util.AddressEquals(lastConfirmed, common.HexToAddress(logW.appCfg.Creator)) { //最终确认人
			//if contentByte, err := logW.ldb.GetByte([]byte(comm.HASH_ADD_CONTENT_PREFIX + hash.Hex())); err != nil {
			//	logger.Error("load content err:%v", err)
			//} else {
			grpcStream := &comm.GrpcStream{BlockNumber: log.BlockNumber, Type: comm.GRPC_HASH_ADD_LOG, Hash: hash, Status: comm.HASH_STATUS_APPLY}
			if grpcStreamJson, err := json.Marshal(grpcStream); err != nil {
				logger.Error("EventStream marshal failed. cause:%v", err)
			} else {
				//write to db
				if err := logW.SetGrpcStreamDB(false, grpcStream.Type, hash.Hex(), grpcStreamJson); err != nil {
					logger.Error("landtodb error", err)
				}
			}
			comm.GrpcStreamChan <- grpcStream
		}
		//}
	}
	return nil
}

//确认hash
func enableHashHandler(logW *EthEventLogWatcher, log *types.Log) error {
	logger.Debug("enableHashHandler......")
	if dataBytes := log.Data; len(dataBytes) > 0 {
		hash := common.BytesToHash(dataBytes[:32])
		lastConfirmed := common.BytesToAddress(dataBytes[32:64])
		logger.Debug("enableHashHandler......db....", hash)
		if util.AddressEquals(lastConfirmed, common.HexToAddress(logW.appCfg.Creator)) { //最终确认人
			//if contentByte, err := logW.ldb.GetByte([]byte(comm.HASH_ADD_CONTENT_PREFIX + hash.Hex())); err != nil {
			//	logger.Error("load content err:%v", err)
			//} else {
				grpcStream := &comm.GrpcStream{BlockNumber: log.BlockNumber, Type: comm.GRPC_HASH_ENABLE_LOG, Hash: hash}
				if grpcStreamJson, err := json.Marshal(grpcStream); err != nil {
					logger.Error("EventStream marshal failed. cause:%v", err)
				} else {
					//write to db
					if err := logW.SetGrpcStreamDB(false, grpcStream.Type, hash.Hex(), grpcStreamJson); err != nil {
						logger.Error("landtodb error", err)
					}
				}
				comm.GrpcStreamChan <- grpcStream
			//}
		}
	}
	return nil
}

//禁用hash
func disableHashHandler(logW *EthEventLogWatcher, log *types.Log) error {
	logger.Debug("disableHashHandler......")
	if dataBytes := log.Data; len(dataBytes) > 0 {
		hash := common.BytesToHash(dataBytes[:32])
		lastConfirmed := common.BytesToAddress(dataBytes[32:64])
		logger.Debug("disableHashHandler......db....", hash)

		if util.AddressEquals(lastConfirmed, common.HexToAddress(logW.appCfg.Creator)) { //最终确认人
			//if contentByte, err := logW.ldb.GetByte([]byte(comm.HASH_ADD_CONTENT_PREFIX + hash.Hex())); err != nil {
			//	logger.Error("load content err:%v", err)
			//} else {
				grpcStream := &comm.GrpcStream{BlockNumber: log.BlockNumber, Type: comm.GRPC_HASH_DISABLE_LOG, Hash: hash}
				if grpcStreamJson, err := json.Marshal(grpcStream); err != nil {
					logger.Error("EventStream marshal failed. cause:%v", err)
				} else {
					//write to db
					if err := logW.SetGrpcStreamDB(false, grpcStream.Type, hash.Hex(), grpcStreamJson); err != nil {
						logger.Error("landtodb error", err)
					}
				}
				comm.GrpcStreamChan <- grpcStream
			//}
		}
	}
	return nil
}

//提现申请
func withdrawApplyHandler(logW *EthEventLogWatcher, log *types.Log) error {
	logger.Debug("withdrawAplyHandler......")

	if dataBytes := log.Data; len(dataBytes) > 0 {
		hash := log.Topics[1]
		wdHash := log.Topics[2]
		amount := common.BytesToHash(dataBytes[:32]).Big()
		fee := common.BytesToHash(dataBytes[32:64]).Big()
		category := common.BytesToHash(dataBytes[96:128]).Big()
		var to string = ""
		if category.Int64() == comm.CATEGORY_BTC {
			//获取db中的地址数据
			if recAddrByte, err := logW.ldb.GetByte([]byte(comm.APPROVE_RECADDR_PREFIX + wdHash.Hex())); err != nil {
				logger.Error("load recAddress err:%v", err)
			} else {
				to = string(recAddrByte)
			}
		} else {
			to = common.BytesToAddress(dataBytes[64:96]).Hex()
		}
		logger.Debug("withdrawAplyHandler......db....")
		lastConfirmed := common.BytesToAddress(dataBytes[128:160])
		if util.AddressEquals(lastConfirmed, common.HexToAddress(logW.appCfg.Creator)) { //最终确认人
			grpcStream := &comm.GrpcStream{BlockNumber: log.BlockNumber, Type: comm.GRPC_WITHDRAW_LOG, Hash: hash, WdHash: wdHash, Amount: amount, Fee: fee, To: to, Category: category}
			if grpcStreamJson, err := json.Marshal(grpcStream); err != nil {
				logger.Error("EventStream marshal failed. cause:%v", err)
			} else {
				//write to db
				if err := logW.SetGrpcStreamDB(false, grpcStream.Type, wdHash.Hex(), grpcStreamJson); err != nil {
					logger.Error("landtodb error", err)
				}
			}
			comm.GrpcStreamChan <- grpcStream
		}
	}
	return nil
}

//解析地址
func parseAddress(data []byte) common.Address {
	addr := bytes.TrimLeftFunc(data[:32], func(r rune) bool {
		return r == 0
	})
	return common.BytesToAddress(addr)
}

// 解析充值事件
func parseDeposit(data []byte) (string, float64) {
	return parseUUID(data[:32]), parseAmount(data[32:])
}

//解析金额
func parseAmount(data []byte) float64 {
	logger.Debug("parseAmount-> data: %v", data)
	value := new(big.Float).SetInt(new(big.Int).SetBytes(data))
	result := new(big.Float).Quo(value, unit)
	v, a := result.Float64()
	logger.Info("big Float: %v, %f, flag: %v", value, v, a)
	return v
}

//解析UUID
func parseUUID(data []byte) string {
	var ret uuid.UUID
	copy(ret[:], data[:16])
	return ret.String()
}

func signFunc(f string) common.Hash {
	data := crypto.Keccak256([]byte(f))
	return common.BytesToHash(data)
}
