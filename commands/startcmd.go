package commands

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/rpc"

	logger "github.com/alecthomas/log4go"
	//"github.com/astaxie/beego"
	"github.com/boxproject/companion/comm"
	"github.com/boxproject/companion/config"
	//"github.com/boxproject/companion/controllers"
	"github.com/boxproject/companion/db"
	"github.com/boxproject/companion/grpcserver"
	"github.com/boxproject/companion/handler"
	"github.com/boxproject/companion/watcher"
	"gopkg.in/urfave/cli.v1"
)

func StartCmd(c *cli.Context) error {
	logger.Debug("Starting companion service...")
	cfg, err := LoadConfig(c.String("c"), "config.json")
	if err != nil {
		logger.Error("Load config failed. cause: %v", err)
		return err
	}
	logger.Info("Load config.  %v", cfg)

	//init db
	db, err := initDb(cfg.LevelDbPath)
	if err != nil {
		logger.Error("Init Db failed . cause: %v", err)
		return err
	}
	comm.Ldb = db

	//prichain conn
	priLogWatcher, err := connPriChain(c, cfg, db)
	if err != nil {
		logger.Error("Dial to the pri geth node failed. cause: %v", err)
		return err
	}

	//init grpc
	go initGrpcSer(cfg, priLogWatcher)

	// monitor log
	go priLogWatcher.Listen()

	//sink合约同步处理
	handler.PriSynEth, err = handler.InitPriSynEthHandler(cfg.SinkAddress, cfg.PriEthCfg)
	if err != nil {
		logger.Error("Init Syn EthHandler failed . cause: %v", err)
		return err
	}
	asyEthHandler := handler.NewPriAsyEthHandler(cfg, db)
	go asyEthHandler.Start()
	//提供http服务
	//go httpServer()

	////上报程序
	//repCli := httpcli.NewRepCli(cfg)
	//repCli.Start()

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh,
		syscall.SIGINT, syscall.SIGTERM,
		syscall.SIGHUP, syscall.SIGKILL,
		syscall.SIGUSR1, syscall.SIGUSR2)
	<-signalCh

	asyEthHandler.Close()
	priLogWatcher.Stop()
	//repCli.Stop()

	logger.Info("companion has already been shutdown...")
	return nil
}

//connect private chain
func connPriChain(c *cli.Context, cfg *config.Config, ldb *db.Ldb) (*watcher.EthEventLogWatcher, error) {
	logger.Info("conn pri eth start........")
	priClient, err := rpc.Dial(cfg.PriEthCfg.GethAPI)
	if err != nil {
		logger.Error("Dial to the geth node failed, cause: %v", err)
		return nil, err
	}

	cursorPath := c.String("b") //priority
	if cursorPath == "" {
		cursorPath = cfg.PriEthCfg.CursorFilePath
	}

	blkFile := GetConfigFilePath(cursorPath, comm.DEF_CURSOR_FILE_PATH)
	logger.Debug("Blockfile: %s", blkFile)

	logWatcher, err := watcher.NewEthEventLogWatcher(priClient, &cfg.PriEthCfg, blkFile, ldb)
	if err != nil {
		logger.Error("New ETH Event log watcher failed. cause: %v", err)
		return nil, err
	}

	if err = logWatcher.Initial(watcher.PriEventMap); err != nil {
		logger.Error("initial block infomation failed. cause: %v", err)
		return nil, err
	}
	return logWatcher, nil

}

//init db
func initDb(path string) (*db.Ldb, error) {
	return db.InitDb(path)
}

//init grpc
func initGrpcSer(cfg *config.Config, watcher *watcher.EthEventLogWatcher) error {
	return grpcserver.InitConn(cfg, watcher)
}

//http
func httpServer() {
	//beego.Router(ServiceName_HASH, &controllers.HashController{}, "get,post:Hash")
	//beego.Router(ServiceName_APPLY, &controllers.ApplyController{}, "get,post:Apply")
	//
	//beego.Run()
}
