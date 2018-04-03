package grpcserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	//"io"
	"io/ioutil"
	"time"

	"github.com/boxproject/companion/comm"
	"github.com/boxproject/companion/config"
	pb "github.com/boxproject/companion/pb"
	"github.com/boxproject/companion/util"
	"github.com/boxproject/companion/watcher"

	log "github.com/alecthomas/log4go"
	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"math/big"
)

//
type replyServer struct {
	routerInfo config.RouterInfo
	conn       *grpc.ClientConn
	watcher    *watcher.EthEventLogWatcher
	isRouther  bool
}

func loadCredential(cfg *config.Config) (credentials.TransportCredentials, error) {
	//加载证书
	cert, err := tls.LoadX509KeyPair(cfg.ClientCert, cfg.ClientKey)
	if err != nil {
		return nil, err
	}

	certBytes, err := ioutil.ReadFile(cfg.ClientCert)
	if err != nil {
		return nil, err
	}

	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(certBytes)
	if !ok {
		return nil, err
	}

	config := &tls.Config{
		RootCAs:            clientCertPool,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
	}
	return credentials.NewTLS(config), nil
}

func InitConn(cfg *config.Config, watcher *watcher.EthEventLogWatcher) error {
	log.Debug("init rpc client ....")

	//重新发送失败GRPC
	watcher.ReSendGrpcStream()

	cred, err := loadCredential(cfg)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}
	conn, err := grpc.Dial(cfg.GrpcSerHost, grpc.WithTransportCredentials(cred))
	if err != nil {
		log.Error("connect to the remote server failed. cause: %v", err)
		return err
	}
	replyServer := &replyServer{conn: conn, watcher: watcher, routerInfo: cfg.RouterInfo}

	go streamRecv(replyServer)

	return nil
}

//stream recv
func streamRecv(n *replyServer) {
	timerListen := time.NewTicker(time.Second * 5)
	timeCount := 1
	for {
		select {
		case <-timerListen.C:
			log.Info("try reveive...%d", timeCount)
			client := pb.NewSynchronizerClient(n.conn)
			stream, err := client.Listen(context.TODO())
			if err != nil {
				log.Error("[STREAM ERR] %v\n", err)
			} else {
				waitc := make(chan struct{})
				//注册服务
				stream.Send(&pb.ListenReq{ServerName: n.routerInfo.SerCompanion, Name: n.routerInfo.CompanionName, Ip: util.GetCurrentIp()})
				go func() {
					for {
						if resp, err := stream.Recv(); err != nil { //rec error
							log.Error("[STREAM ERR] %v\n", err)
							close(waitc)
							return
						} else {
							//log.Debug("stream Recv: %s\n", resp)
							handleStream(resp)
						}
					}
				}()
				//启动心跳检测
				go heart(n)
				//路由发送
				go router(n)
				<-waitc
				n.isRouther = false
				if err = stream.CloseSend(); err != nil {
					log.Error("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
				}
			}
			timeCount++
		}
	}
}

func heart(n *replyServer) {
	timerHeart := time.NewTicker(time.Second * 10)
	timeCount := 1
	for {
		select {
		case <-timerHeart.C:
			//log.Info("try heart...%d", timeCount)
			client := pb.NewSynchronizerClient(n.conn)
			if _, err := client.Heart(context.TODO(), &pb.HeartRequest{RouterType: "grpc", ServerName: n.routerInfo.SerCompanion, Name: n.routerInfo.CompanionName, Ip: util.GetCurrentIp(), Msg: []byte("heart")}); err != nil {
				log.Error("heart req failed %s\n", err)
			} else {
				//log.Debug("heart response", rsp)
			}

			timeCount++
		}
	}
}

func router(n *replyServer) {
	timerTest := time.NewTicker(time.Second * 10)
	n.isRouther = true
	for n.isRouther{
		select {
		case data, ok := <-comm.GrpcStreamChan:
			if ok {
				if msgJson, err := json.Marshal(data); err != nil {
					log.Error("json marshal error:%v", err)
				} else {
					log.Debug("grpc send:\n", data)
					//发送标志
					var isSendOK bool = true
					client := pb.NewSynchronizerClient(n.conn)
					if _, err := client.Router(context.TODO(), &pb.RouterRequest{RouterType: "grpc", RouterName: n.routerInfo.SerCompanion, Msg: msgJson}); err != nil {
						log.Error("heart req failed %s\n", err)
						isSendOK = false
					} else {
						//log.Debug("heart response", rsp)
					}
					//update grpc db
					switch data.Type {
					case comm.GRPC_HASH_ADD_LOG,
						comm.GRPC_HASH_ENABLE_LOG,
						comm.GRPC_HASH_DISABLE_LOG:
						//重新写入数据
						if err := n.watcher.SetGrpcStreamDB(isSendOK, data.Type, data.Hash.Hex(), msgJson); err != nil {
							log.Error("landtodb error", err)
						}
						break
					case comm.GRPC_WITHDRAW_LOG:
						//重新写入数据
						if err := n.watcher.SetGrpcStreamDB(isSendOK, data.Type, data.WdHash.Hex(), msgJson); err != nil {
							log.Error("landtodb error", err)
						}
					default:
						log.Info("no grpc type :", data.Type)
					}
				}
			} else {
				log.Error("read from grpc channel failed")
			}
		case <-timerTest.C:
			//routerTest(n)
		}
	}
}

var routerTestIndex int64  = 0
func routerTest(n *replyServer){
	streams:= []*comm.GrpcStream{
		{Type: comm.GRPC_HASH_ADD_LOG, 	BlockNumber: 10000000,	Hash: common.StringToHash("GRPC_HASH_ADD_LOG"), 	Content: string("GRPC_HASH_ADD_LOG"), 		Status: comm.HASH_STATUS_APPLY, CreateTime: time.Now()},
		{Type:comm.GRPC_HASH_ENABLE_LOG,	BlockNumber: 20000000, 	Hash: common.StringToHash("GRPC_HASH_ENABLE_LOG"), 	Content: string("GRPC_HASH_ENABLE_LOG"), 	Status: comm.HASH_STATUS_ENABLE, CreateTime: time.Now()},
		{Type:comm.GRPC_HASH_DISABLE_LOG,	BlockNumber: 30000000, 	Hash: common.StringToHash("GRPC_HASH_DISABLE_LOG"),	Content: string("GRPC_HASH_DISABLE_LOG"), Status: comm.HASH_STATUS_DISABLE, CreateTime: time.Now()},
		{Type:comm.GRPC_WITHDRAW_LOG,		BlockNumber: 30000000, 	Hash: common.StringToHash("GRPC_WITHDRAW_LOG"), 	WdHash: common.StringToHash("GRPC_WITHDRAW_LOG"), Amount: big.NewInt(11), Fee: big.NewInt(22), To: "to address", Category: big.NewInt(routerTestIndex%5), CreateTime: time.Now()},
		}
	comm.GrpcStreamChan <- streams[routerTestIndex%int64(len(streams))]
	routerTestIndex++
}

//处理流
func handleStream(streamRsp *pb.StreamRsp) {
	streamModel := &comm.GrpcStream{}
	if err := json.Unmarshal(streamRsp.Msg, streamModel); err != nil {
		log.Error("json marshal error:%v", err)
		return
	}
	switch streamModel.Type {
	case comm.GRPC_HASH_ENABLE_REQ: //同意
		hash := streamModel.Hash.Hex()
		if !common.HasHexPrefix(hash) || len(common.FromHex(hash)) != comm.HASH_ENABLE_LENGTH {
			log.Error("allow err")
		} else {
			comm.ReqChan <- &comm.RequestModel{Hash: hash, ReqType: comm.REQ_HASH_ENABLE}
		}
		break
	case comm.GRPC_HASH_DISABLE_REQ: //禁用
		hash := streamModel.Hash.Hex()
		if !common.HasHexPrefix(hash) || len(common.FromHex(hash)) != comm.HASH_ENABLE_LENGTH {
			log.Error("disallow err")
		} else {
			comm.ReqChan <- &comm.RequestModel{Hash: hash, ReqType: comm.REQ_HASH_DISABLE}
		}
		break
	default:
		log.Info("no type,streamModel:\n", streamModel)
	}
}
