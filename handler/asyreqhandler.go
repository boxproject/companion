package handler

import (
	"math/big"
	"os"

	"github.com/boxproject/companion/comm"
	"github.com/boxproject/companion/config"
	"github.com/boxproject/companion/contract"
	"github.com/boxproject/companion/db"
	"github.com/boxproject/companion/util"
	logger "github.com/alecthomas/log4go"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

//异步处理
type PriAsyEthHandler struct {
	ethCfg      config.EthCfg
	quitChannel chan int
	client      *ethclient.Client
	sinkAddress common.Address
	ldb         *db.Ldb
}

func NewPriAsyEthHandler(cfg *config.Config, db *db.Ldb) *PriAsyEthHandler {
	return &PriAsyEthHandler{ethCfg: cfg.PriEthCfg, sinkAddress: common.HexToAddress(cfg.SinkAddress), ldb: db,quitChannel:make(chan int,1)}
}

//上私链操作
func (this *PriAsyEthHandler) Start() {
	logger.Info("PriAsyEthHandler start...")
	priClient, err := ethclient.Dial(this.ethCfg.GethAPI) //断线重连待处理
	if err != nil {
		logger.Error("Dial to the geth node failed. cause: %s", err)
		return
	}
	this.client = priClient
	loop := true
	for loop {
		select {
		case <-this.quitChannel:
			logger.Info("PriEthHandler::SendMessage thread exitCh!")
			loop = false
		case data, ok := <-comm.ReqChan:
			if ok {
				switch data.ReqType {
				case comm.REQ_HASH_ADD:
					this.addHash(data)
				case comm.REQ_HASH_ENABLE:
					this.enableHash(data)
				case comm.REQ_HASH_DISABLE:
					this.disableHash(data)
				case comm.REQ_OUT_APPROVE:
					this.approve(data)
				default:
					logger.Info("unknow asy req: %s", data.ReqType)
				}
			} else {
				logger.Error("PriAsyEthHandler read from channel failed")
			}
		}
	}
}

//关闭私链操作处理
func (this *PriAsyEthHandler) Close() {
	close(this.quitChannel)
	logger.Info("PriAsyEthHandler closed")
}

//hash上链
func (this *PriAsyEthHandler) addHash(req *comm.RequestModel) error {
	logger.Info("PriAsyEthHandler addHash....")

	//if err := this.ldb.PutStrWithPrifix(comm.HASH_ADD_CONTENT_PREFIX, req.Hash, req.Content); err != nil { //content 内容存入db，供私链申请同意后查询使用
	//	logger.Error("land to db failed: %s", err)
	//	return err
	//}

	opts, err := this.createTransactor(this.ethCfg.CreatorKeystorePath, this.ethCfg.CreatorPassphrase)
	if err != nil {
		logger.Error("Create options failed: %s", err)
		return err
	}

	nonce, err := util.ReadNumberFromFile(this.ethCfg.NonceFilePath) //nonce file
	if err != nil {
		logger.Error("read block info err :%s", err)
		return err
	}

	opts.Nonce = nonce
	logger.Debug("current nonce :%d", nonce.Int64())

	sink, err := contract.NewSink(this.sinkAddress, this.client)
	if err != nil {
		logger.Error("NewSink error: %s", err)
		return err
	}
	hash32 := util.Byte2Byte32(common.FromHex(req.Hash))
	if tx, err := sink.AddHash(opts, hash32); err != nil {
		logger.Error("add hash err: %s", err)
		return err
	} else {
		logger.Info("PriAsyEthHandler addHash tx: %s", tx.Hash().Hex())
		nonce = nonce.Add(nonce, big.NewInt(comm.NONCE_PLUS))
		util.WriteNumberToFile(this.ethCfg.NonceFilePath, nonce)
	}
	return nil
}

//hash 确认
func (this *PriAsyEthHandler) enableHash(req *comm.RequestModel) error {
	logger.Debug("PriAsyEthHandler enableHash....")
	opts, err := this.createTransactor(this.ethCfg.CreatorKeystorePath, this.ethCfg.CreatorPassphrase)
	if err != nil {
		logger.Info("Create options failed: %s", err)
		return err
	}

	nonce, err := util.ReadNumberFromFile(this.ethCfg.NonceFilePath) //nonce file
	if err != nil {
		logger.Error("read block info err :%s", err)
		return err
	}

	opts.Nonce = nonce
	logger.Debug("current nonce :%d", nonce.Int64())

	sink, err := contract.NewSink(this.sinkAddress, this.client)
	if err != nil {
		logger.Info("NewSink error: %s", err)
	}
	hash32 := util.Byte2Byte32(common.FromHex(req.Hash))
	if tx, err := sink.Enable(opts, hash32); err != nil {
		logger.Info("enable hash err: %s",err)
		return err
	} else {
		logger.Info("PriAsyEthHandler enableHash: %s", tx.Hash().Hex())
		nonce = nonce.Add(nonce, big.NewInt(comm.NONCE_PLUS))
		util.WriteNumberToFile(this.ethCfg.NonceFilePath, nonce)
	}

	return nil
}

//hash 禁用
func (this *PriAsyEthHandler) disableHash(req *comm.RequestModel) error {
	logger.Info("PriAsyEthHandler disableHash....")
	opts, err := this.createTransactor(this.ethCfg.CreatorKeystorePath, this.ethCfg.CreatorPassphrase)
	if err != nil {
		logger.Info("Create options failed: %s", err)
		return err
	}

	nonce, err := util.ReadNumberFromFile(this.ethCfg.NonceFilePath) //nonce file
	if err != nil {
		logger.Error("read block info err :%s", err)
		return err
	}

	opts.Nonce = nonce
	logger.Debug("current nonce :%d", nonce.Int64())

	sink, err := contract.NewSink(this.sinkAddress, this.client)
	if err != nil {
		logger.Info("NewSink error: %s", err)
	}

	hash32 := util.Byte2Byte32(common.FromHex(req.Hash))
	if tx, err := sink.Disable(opts, hash32); err != nil {
		logger.Info("disable hash err: %s",err)
		return err
	} else {
		logger.Info("Transaction disableHash: %s\n", tx.Hash().Hex())
		nonce = nonce.Add(nonce, big.NewInt(comm.NONCE_PLUS))
		util.WriteNumberToFile(this.ethCfg.NonceFilePath, nonce)
	}
	return nil
}

//out apply
func (this *PriAsyEthHandler) approve(req *comm.RequestModel) error {
	logger.Debug("PriAsyEthHandler approve....")

	if err := this.ldb.PutStrWithPrifix(comm.APPROVE_RECADDR_PREFIX, req.WdHash, req.RecAddress); err != nil { //recaddress 内容存入db，供私链申请同意后查询使用
		logger.Error("land to db failed: %s", err)
		return err
	}

	opts, err := this.createTransactor(this.ethCfg.CreatorKeystorePath, this.ethCfg.CreatorPassphrase)
	if err != nil {
		logger.Info("Create options failed: %s", err)
		return err
	}

	nonce, err := util.ReadNumberFromFile(this.ethCfg.NonceFilePath) //nonce file
	if err != nil {
		logger.Error("read block info err :%s", err)
		return err
	}

	opts.Nonce = nonce
	logger.Debug("current nonce :%d", nonce.Int64())

	sink, err := contract.NewSink(this.sinkAddress, this.client)
	if err != nil {
		logger.Info("NewSink error: %s", err)
	}

	hash32 := util.Byte2Byte32(common.FromHex(req.Hash))
	wdHash32 := util.Byte2Byte32(common.FromHex(req.WdHash))
	amount := new(big.Int)
	amount.SetString(req.Amount, 10)
	//amount := big.NewInt()

	fee := new(big.Int)
	fee.SetString(req.Fee, 10)
	category := big.NewInt(req.Category)
	recAddress, err := util.GetRecAddress(*req)
	if err != nil {
		logger.Error("getRecAddress err:", err)
		return err
	}

	logger.Debug("recAddress...", recAddress.Hex())
	if tx, err := sink.Approve(opts, wdHash32, amount, fee, recAddress, hash32, category); err != nil {
		logger.Error("approve tx err: %s", err)
		return err
	} else {
		logger.Info("Transaction hash: %s\n", tx.Hash().Hex())
		nonce = nonce.Add(nonce, big.NewInt(comm.NONCE_PLUS))
		util.WriteNumberToFile(this.ethCfg.NonceFilePath, nonce)
	}

	return nil
}

func (this *PriAsyEthHandler) createTransactor(filePath, passphrase string) (*bind.TransactOpts, error) {
	keyFile, err := openKey(filePath)
	if err != nil {
		return nil, err
	}
	defer keyFile.Close()
	if transactor, err := bind.NewTransactor(keyFile, passphrase); err != nil {
		return nil, err
	} else {
		//transactor.GasLimit = uint64(this.ethCfg.GasLimit)
		//transactor.GasPrice = big.NewInt(this.ethCfg.GasPrice)
		return transactor, nil
	}
}

func openKey(filePath string) (*os.File, error) {
	return os.OpenFile(filePath, os.O_RDONLY, 0600)
}
