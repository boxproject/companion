package watcher

import (
	"context"
	"encoding/json"
	"math/big"

	logger "github.com/alecthomas/log4go"
	"github.com/boxproject/companion/comm"
	"github.com/boxproject/companion/config"
	"github.com/boxproject/companion/db"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"strings"
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
	logger.Debug("Block file: %s", logW.blkFile)

	cursorBlkNumber, err := ReadBlockNumberFromFile(logW.blkFile)
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

	// 获取向前推N个区块的big.Int值
	logW.checkBefore = big.NewInt(logW.appCfg.CheckBlockBefore)
	logger.Debug("before:: max blkNumber: %s, cursor blkNumber: %s", maxBlkNumber.String(), cursorBlkNumber.String())

	// -------|-------------------|
	//    current                max
	//  max - current >= checkBefore(30) 检查向前推的区块
	diff := new(big.Int).Sub(maxBlkNumber, cursorBlkNumber)
	for diff.Cmp(logW.checkBefore) != -1 {
		if err = logW.checkLogs(cursorBlkNumber); err != nil {
			return err
		}

		cursorBlkNumber = new(big.Int).Add(cursorBlkNumber, big.NewInt(1))
		diff = new(big.Int).Sub(maxBlkNumber, cursorBlkNumber)
	}
	// 记录下当前的 blocknumber 供恢复用
	//-----
	logger.Info("logW.blkFile---------", logW.blkFile)

	WriteCheckpointBlockNumberToFile(logW.blkFile, new(big.Int).Sub(cursorBlkNumber, big.NewInt(1)))

	logger.Debug("after:: cursor blkNumber: %s", cursorBlkNumber.String())
	return nil
}

func (logW *EthEventLogWatcher) Listen() {
	ch := make(chan *types.Header)
	sid, err := logW.client.SubscribeNewHead(context.Background(), ch)
	if err != nil {
		logger.Error("Sub to the block failed. cause: %v\n", err)
		return
	}

	defer sid.Unsubscribe()

	if err = logW.recv(sid, ch); err != nil {
		logger.Error("Resubscribe to the geth. Receive header failed. cause: %v\n", err)
		// reconnect
		go logW.Listen()
	}
}

func (logW *EthEventLogWatcher) Stop() {
	close(logW.quitSignal)
	logger.Info("ETH Event log Watcher stopped!")
}

func (logW *EthEventLogWatcher) recv(sid ethereum.Subscription, ch <-chan *types.Header) error {
	var err error
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
			if head.Number == nil {
				continue
			}

			if err = logW.checkLogs(head.Number); err != nil {
				return err
			}
		}
	}
}

func (logW *EthEventLogWatcher) checkLogs(blkNumber *big.Int) error {
	checkPoint, err := ReadBlockNumberFromFile(logW.blkFile) //文件中记录的点
	if err != nil {
		logger.Error("read block number err: %s", err)
		return err
	}

	if checkPoint.Cmp(blkNumber) > 0 { // 文件中记录点有误
		checkPoint = blkNumber
	} else if checkPoint.Cmp(big.NewInt(0)) <= 0 {
		checkPoint = new(big.Int).Sub(blkNumber, logW.checkBefore)
	} else {
		checkPoint = checkPoint.Add(checkPoint, big.NewInt(1))
	}

	logger.Debug("[HEADER] FromBlock : %s, blkNumber: %s", checkPoint.String(), blkNumber.String())

	//logger.Debug("[HEADER] blkNumber: %s， blkNumber checkpoint: %s", blkNumber.String(), checkPoint.String())
	if logs, err := logW.client.FilterLogs(
		context.Background(),
		ethereum.FilterQuery{
			FromBlock: checkPoint,
			ToBlock:   blkNumber,
		}); err != nil {
		logger.Debug("FilterLogs :%s", err)
		return err
	} else {
		if len(logs) == 0 {
			return nil
		}
		for _, log := range logs {
			handler, ok := logW.eventHandlerMap[log.Topics[0]]

			//WriteCheckpointBlockNumberToFile(logW.blkFile, big.NewInt(int64(log.BlockNumber)))
			if !ok {
				logger.Info("false No ==> %s", log.Topics[0].Hex())
				continue
			}
			logger.Info("true No ==> %s", log.Topics[0].Hex())
			if err = handler(logW, &log); err != nil {
				logger.Error("log handler err: %s", err)
				WriteCheckpointBlockNumberToFile(logW.blkFile, big.NewInt(int64(log.BlockNumber)))
				return err
			}
		}
		WriteCheckpointBlockNumberToFile(logW.blkFile, blkNumber)
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
					grpcStream.CreateTime = time.Now()
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
