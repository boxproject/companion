package handler

import (


	"github.com/boxproject/companion/comm"
	"github.com/boxproject/companion/config"
	"github.com/boxproject/companion/contract"
	"github.com/boxproject/companion/util"
	logger "github.com/alecthomas/log4go"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var PriSynEth *PriSynEthHandler

//同步处理
type PriSynEthHandler struct {
	ethCfg   config.EthCfg
	client   *ethclient.Client
	sinkAddr common.Address
}

func InitPriSynEthHandler(sinkAddrStr string, cfg config.EthCfg) (*PriSynEthHandler, error) {
	priClient, err := ethclient.Dial(cfg.GethAPI) //断线重连待处理
	if err != nil {
		logger.Error("Dial to the geth node failed. cause: %v", err)
		return nil, err
	}
	sinkAddr := common.HexToAddress(sinkAddrStr)
	return &PriSynEthHandler{ethCfg: cfg, client: priClient, sinkAddr: sinkAddr}, nil
}

//hash是否有效
func (p *PriSynEthHandler) HashAvailable(hashStr string) (bool, error) {
	logger.Info("PriEthHandler availableHash....")
	opts, err := p.createCallOpts()
	if err != nil {
		logger.Info("Create callopts failed", err)
		return false, err
	}
	sink, err := contract.NewSink(p.sinkAddr, p.client)
	if err != nil {
		logger.Info("NewSink error:", err)
	}

	hash := common.FromHex(hashStr)
	hash32 := util.Byte2Byte32(hash)
	tx, b, err := sink.Available(opts, hash32)

	if err != nil {
		logger.Info(err)
		return false, err
	}
	hh := make([]byte, comm.HASH_ENABLE_LENGTH)
	copy(hh[0:comm.HASH_ENABLE_LENGTH], tx[:])
	logger.Info("Transaction hash: %s", common.Bytes2Hex(hh))
	logger.Info("Transaction b: %s", b)
	return b, nil
}

//申请是否成功
func (p *PriSynEthHandler) TxExists(hashStr, txHashStr string) (bool, error) {
	logger.Info("PriEthHandler availableHash....")
	opts, err := p.createCallOpts()
	if err != nil {
		logger.Info("Create callopts failed", err)
		return false, err
	}

	sink, err := contract.NewSink(p.sinkAddr, p.client)
	if err != nil {
		logger.Info("NewSink error:", err)
	}
	hash := common.FromHex(hashStr)
	txHash := common.FromHex(txHashStr)

	hash32 := util.Byte2Byte32(hash)
	txHash32 := util.Byte2Byte32(txHash)
	result, err := sink.TxExists(opts, hash32, txHash32)

	if err != nil {
		logger.Info(err)
		return false, err
	}
	logger.Info("Transaction result: %s", result)
	return result, nil
}

func (p *PriSynEthHandler) createCallOpts() (*bind.CallOpts, error) {
	return &bind.CallOpts{}, nil
}
