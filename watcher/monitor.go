package watcher

import (
	"context"
	"encoding/json"
	"math/big"

	logger "github.com/alecthomas/log4go"
	"github.com/boxproject/companion/comm"
	"github.com/boxproject/companion/config"
	"github.com/boxproject/companion/db"
	"github.com/boxproject/companion/util"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"strings"
	//"time"
	"time"
)

type EthEventLogWatcher struct {
	client          *ethclient.Client
	appCfg          *config.EthCfg
	blkFile         string
	quitSignal      chan struct{}
	eventHandlerMap map[common.Hash]EventHandler
	checkBefore     *big.Int
	ldb             *db.Ldb
}

func NewEthEventLogWatcher(c *rpc.Client, ethCfg *config.EthCfg, blkFile string, ldb *db.Ldb) (*EthEventLogWatcher, error) {
	client := ethclient.NewClient(c)
	logWatcher := &EthEventLogWatcher{
		client:     client,
		appCfg:     ethCfg,
		blkFile:    blkFile,
		quitSignal: make(chan struct{}),
		ldb:        ldb,
	}

	return logWatcher, nil
}

func (logW *EthEventLogWatcher) Initial(events map[common.Hash]EventHandler) error {
	logW.eventHandlerMap = events
	// 读取当前日志记录下的区块号
	logger.Debug("Block file:[%v]", logW.blkFile)

	lastCursorBlkNumber, err := ReadBlockNumberFromFile(logW.blkFile)
	if err != nil {
		logger.Error("Read current blkNumber from file failed, cause: %v", err)
		return err
	}

	// 获取当前节点上最大区块号
	blk, err := logW.client.BlockByNumber(context.Background(), nil)
	if err != nil {
		logger.Error("Get blkNumber from geth node failed. cause: %v", err)
		return err
	}
	maxBlkNumber := blk.Number()

	//nonce值初始化
	nonce, err := logW.client.NonceAt(context.Background(), common.HexToAddress(logW.appCfg.Creator), maxBlkNumber)
	if err != nil { //设置初始值nonce值
		logger.Error("get nonce err: %v", err)
	}
	util.WriteNumberToFile(logW.appCfg.NonceFilePath, big.NewInt(int64(nonce)))
	logger.Info("Nonce file:[%v], current nonce value:%v", logW.appCfg.NonceFilePath, nonce)

	// 获取向前推N个区块的big.Int值
	logW.checkBefore = big.NewInt(logW.appCfg.CheckBlockBefore)
	logger.Info("[BEGIN] rescan block ...")
	logger.Info("Last scan block height: %v", lastCursorBlkNumber.String())
	logger.Info("Current max block height: %v", maxBlkNumber.String())
	// -------|-------------------|
	//    current                max
	//  max - current >= checkBefore(30) 检查向前推的区块
	cursorBlkNumber := new(big.Int).Add(lastCursorBlkNumber, logW.checkBefore)
	for maxBlkNumber.Cmp(new(big.Int).Add(cursorBlkNumber, big.NewInt(1))) >= 0 {
		if err := logW.checkLogs(new(big.Int).Add(cursorBlkNumber, big.NewInt(1))); err != nil {
			continue
		}
		//add 1 to next
		cursorBlkNumber.Add(cursorBlkNumber, big.NewInt(1))
	}
	logger.Info("current scan block height: %v", new(big.Int).Sub(cursorBlkNumber, logW.checkBefore))
	logger.Info("[END] rescan block ...")

	return nil
}

func (logW *EthEventLogWatcher) Listen() {

	ch := make(chan *types.Header)
	sid, err := logW.client.SubscribeNewHead(context.Background(), ch)
	if err == nil {
		err = logW.recv(sid, ch)
		defer sid.Unsubscribe()
	}

	//retry connect
	if err != nil {
		logger.Error("[ETH CONNECT ERROR]: %v", err)
		d := util.DefaultBackoff.Duration(util.RetryCount)
		if d > 0 {
			time.Sleep(d)
			logger.Info("[RETRY ETH CONNECT][%v] sleep:%v", util.RetryCount, d)
			util.RetryCount++
			go logW.Listen()
			return
		}
	}
	logger.Debug("watcher listener stopped.")
}

func (logW *EthEventLogWatcher) Stop() {
	close(logW.quitSignal)
	logger.Info("ETH Event log Watcher stopped!")
}

func (logW *EthEventLogWatcher) recv(sid ethereum.Subscription, ch <-chan *types.Header) error {
	logger.Debug("EthHandler recv...")
	var err error
	var lastScanHeight = big.NewInt(-1)
	for {
		select {
		case <-logW.quitSignal:
			logger.Info("Monitor stopped!")
			return nil
		case err = <-sid.Err():
			if err != nil {
				logger.Error("When subscribe the header, error found. cause: %v", err)
				return err
			}
		case head := <-ch:
			if util.RetryCount != 0 {
				//发现掉线,重新扫描
				logW.Initial(PriEventMap)
				util.RetryCount = 0
				continue
			}
			if head.Number == nil {
				continue
			}
			if lastScanHeight.Cmp(head.Number) != 0 {
				if err = logW.checkLogs(head.Number); err != nil {
					return err
				}
				lastScanHeight = head.Number
			}else {
				logger.Debug("[BLOCK] Get Same Block: %v",head.Number)
			}
		}
	}
}

func (logW *EthEventLogWatcher) checkLogs(blkNumber *big.Int) error {
	checkPoint := new(big.Int).Sub(blkNumber, logW.checkBefore)

	logger.Debug("[BLOCK] GetBlock: %v, CheckBlock: %v", blkNumber, checkPoint)

	//logger.Debug("[HEADER] blkNumber: %s， blkNumber checkpoint: %s", blkNumber.String(), checkPoint.String())
	if logs, err := logW.client.FilterLogs(
		context.Background(),
		ethereum.FilterQuery{
			FromBlock: checkPoint,
			ToBlock:   checkPoint,
		}); err != nil {
		logger.Error("FilterLogs :%s", err)
		return err
	} else {
		if len(logs) != 0 {
			for _, log := range logs {
				if log.Topics == nil || len(log.Topics) == 0 {
					logger.Info("enventLog topics nil")
					continue
				}

				handler, ok := logW.eventHandlerMap[log.Topics[0]]
				if !ok {
					logger.Info("false No ==> %s", log.Topics[0].Hex())
					continue
				}
				logger.Info("true No ==> %s", log.Topics[0].Hex())
				if err = handler(logW, &log); err != nil {
					logger.Error("log handler err: %s", err)
					return err
				}
			}
		}
		WriteCheckpointBlockNumberToFile(logW.blkFile, checkPoint)
	}

	return nil
}

//待修改
func (logW *EthEventLogWatcher) CheckLogs(blkNumber *big.Int) error {
	checkPoint := new(big.Int).Sub(blkNumber, logW.checkBefore)
	logger.Debug("[HEADER] blkNumber: %s， blkNumber checkpoint: %s", blkNumber.String(), checkPoint.String())

	logs, err := logW.client.FilterLogs(
		context.Background(),
		ethereum.FilterQuery{
			FromBlock: checkPoint,
			ToBlock:   checkPoint,
		})

	if err != nil {
		return err
	}

	if len(logs) == 0 {
		return nil
	}

	for _, log := range logs {
		handler, ok := logW.eventHandlerMap[log.Topics[0]]
		if !ok {
			logger.Info("false No ==> %s", log.Topics[0].Hex())
			continue
		}
		logger.Info("true No ==> %s", log.Topics[0].Hex())
		if err = handler(logW, &log); err != nil {
			return err
		}
	}
	return nil
}

func (logW *EthEventLogWatcher) SetGrpcStreamDB(isSendOK bool, infoType string, keyIndex string, value []byte) error {
	//删除原有数据
	var keysDelFlag string
	var keysSetFlag string
	if isSendOK {
		keysDelFlag = "0_"
		keysSetFlag = "1_"
	} else {
		keysDelFlag = "1_"
		keysSetFlag = "0_"
	}

	//logW.ldb.DelKey()
	switch infoType {
	case comm.GRPC_HASH_ADD_LOG,
		comm.GRPC_HASH_ENABLE_LOG,
		comm.GRPC_HASH_DISABLE_LOG,
		comm.GRPC_WITHDRAW_LOG:
		//logger.Debug("isSendOK:",isSendOK," infoType:",infoType, " keyIndex:",keyIndex," value:",string(value))
		//先删除数据
		logW.DelKey([]byte(comm.GRPC_DB_PREFIX + keysDelFlag + infoType + "_" + keyIndex))
		logW.DelKey([]byte(comm.GRPC_DB_PREFIX + keysSetFlag + infoType + "_" + keyIndex))
		//重新写入数据
		if err := logW.PutByte([]byte(comm.GRPC_DB_PREFIX+keysSetFlag+infoType+"_"+keyIndex), value); err != nil {
			logger.Error("landtodb error", err)
		}
	default:
		logger.Info("no grpc type :", infoType)
	}

	return nil
}

//GRPC重发检测
func (logW *EthEventLogWatcher) ReSendGrpcStream() error {
	//logger.Debug("ReSendGrpcStream....")
	if mapHashAdd, err := logW.ldb.GetPrifix([]byte(comm.GRPC_DB_PREFIX + "0_")); err != nil {
		logger.Error("get db error:", err)
	} else {
		for i, value := range mapHashAdd {
			//logger.Debug("key:",i,"value:",value)
			index := strings.Split(i, "_")[:]
			switch index[2] {
			case comm.GRPC_HASH_ADD_LOG,
				comm.GRPC_HASH_ENABLE_LOG,
				comm.GRPC_HASH_DISABLE_LOG,
				comm.GRPC_WITHDRAW_LOG:
				grpcStream := &comm.GrpcStream{}
				if err := json.Unmarshal([]byte(value), grpcStream); err != nil {
					logger.Error("db unmarshal err: %v", err)
				} else {
					//update time
					//grpcStream.CreateTime = time.Now()
					comm.GrpcStreamChan <- grpcStream
					//logger.Debug("resend....",grpcStream)
					logger.Debug("grpc resend, type value:", grpcStream.Type)
				}
				break
			default:
				logger.Info("no grpc type..", index[2])
			}
		}
	}

	return nil
}

func (logW *EthEventLogWatcher) PutByte(key, value []byte) error {
	return logW.ldb.PutByte(key, value)
}
func (logW *EthEventLogWatcher) DelKey(key []byte) error {
	return logW.ldb.DelKey(key)
}
