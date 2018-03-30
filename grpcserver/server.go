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
)
//
type replyServer struct {
	conn *grpc.ClientConn
	watcher *watcher.EthEventLogWatcher
}
//
//func (s *replyServer) Heart(ctx context.Context, req *pb.HeartRequest) (*pb.HeartResponse, error) {
//	log.Debug("Heart.....hash:", req)
//	response := &pb.HeartResponse{}
//	return response, nil
//}
//
//func (s *replyServer) Router(ctx context.Context, req *pb.RouterRequest) (*pb.RouterResponse, error) {
//	log.Debug("Router.....hash:", req)
//	response := &pb.RouterResponse{}
//	return response, nil
//}
////
//////TODO blockNum
//func (s *replyServer) Listen(stream pb.Synchronizer_ListenServer) error {
//	defer log.Info("grpc server listen end ......")
//	log.Info("grpc server listen start......")
//
//	listReq, err := stream.Recv()
//	log.Debug("listReq: %s", listReq)
//	quitCh := make(chan bool)
//
//	go func() { //监控连接情况
//		_, err = stream.Recv()
//		if err == io.EOF {
//			log.Debug("err EOF...", err)
//			quitCh <- false
//		}
//
//		if err != nil {
//			log.Error("[LISTEN ERR] %v\n", err)
//			quitCh <- false
//		}
//	}()
//
//	for {
//		select {
//		case data, ok := <-comm.GrpcStreamChan:
//			if ok {
//				if msgJson, err := json.Marshal(data); err != nil {
//					log.Error("json marshal error:%v", err)
//				} else {
//					log.Debug("grpc send...", data)
//					stream.Send(&pb.StreamRsp{Msg: msgJson})
//				}
//			} else {
//				log.Error("read from grpc channel failed")
//			}
//		case <-quitCh:
//			{
//				return nil
//			}
//		}
//	}
//	return nil
//}

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
	replyServer := &replyServer{conn:conn,watcher:watcher}

	go streamRecv(replyServer)
	go heart(replyServer)
	go router(replyServer)

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
				stream.Send(&pb.ListenReq{ServerName: "grpc", Name: "companion", Ip: util.GetCurrentIp()})
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
				<-waitc
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
			log.Info("try heart...%d", timeCount)
			client := pb.NewSynchronizerClient(n.conn)
			if rsp, err := client.Heart(context.TODO(), &pb.HeartRequest{RouterType: "grpc", ServerName: "grpc", Name: "companion", Ip: util.GetCurrentIp()}); err != nil {
				log.Error("heart req failed %s\n", err)
			} else {
				log.Debug("heart response", rsp)
			}
			timeCount++
		}
	}
}

func router(n *replyServer){
	timerTest := time.NewTicker(time.Second * 3)
	for {
		select {
		case data, ok := <-comm.GrpcStreamChan:
			if ok {
				if msgJson, err := json.Marshal(data); err != nil {
					log.Error("json marshal error:%v", err)
				} else {
					log.Debug("grpc send...", data)
					//发送标志
					var isSendOK bool = true
					client := pb.NewSynchronizerClient(n.conn)
					if rsp, err := client.Router(context.TODO(), &pb.RouterRequest{RouterType: "grpc", RouterName: "grpc", Msg: msgJson}); err != nil {
						log.Error("heart req failed %s\n", err)
						isSendOK = false
					} else {
						log.Debug("heart response", rsp)
					}
					//update grpc db
					switch data.Type {
					case comm.GRPC_ACCOUNT_USE:
						//重新写入数据
						if err := n.watcher.SetGrpcStreamDB(isSendOK,data.Type,data.Account,msgJson); err != nil {
							log.Error("landtodb error", err)
						}
						break
					case comm.GRPC_SIGN_ADD,
						comm.GRPC_SIGN_ENABLE,
						comm.GRPC_SIGN_DISABLE:
						//重新写入数据
						if err := n.watcher.SetGrpcStreamDB(isSendOK, data.Type, data.Hash.Hex(), msgJson); err != nil {
							log.Error("landtodb error", err)
						}
						break
					case comm.GRPC_APPROVE:
						//重新写入数据
						if err := n.watcher.SetGrpcStreamDB(isSendOK, data.Type, data.WdHash.Hex(), msgJson); err != nil {
							log.Error("landtodb error", err)
						}
					default:
						log.Info("no grpc type :",data.Type)
					}
				}
			} else {
				log.Error("read from grpc channel failed")
			}
		case <-timerTest.C:
			data := &comm.GrpcStream{Type:"BTC",BlockNumber:12039402}
			if msgJson, err := json.Marshal(data); err != nil {
				log.Error("json marshal error:%v", err)
			} else {
				//log.Debug("grpc send...", data)

				client := pb.NewSynchronizerClient(n.conn)
				if rsp, err := client.Router(context.TODO(), &pb.RouterRequest{RouterType: "grpc", RouterName: "grpc", Msg: msgJson}); err != nil {
					log.Error("Router req failed %s\n", err)
				} else {
					log.Debug("Router response", rsp)
				}
			}
		}
	}
}

//处理流
func handleStream(streamRsp *pb.StreamRsp) {
	streamModel := &comm.GrpcStream{}
	if err := json.Unmarshal(streamRsp.Msg, streamModel); err != nil {
		log.Error("json marshal error:%v", err)
		return
	}
	switch streamModel.Type {
	case comm.GRPC_SIGN_ENABLE: //同意
		hash := streamModel.Hash.Hex()
		if !common.HasHexPrefix(hash) || len(common.FromHex(hash)) != comm.HASH_ENABLE_LENGTH {
			log.Error("allow err")
		}else {
			comm.ReqChan <- &comm.RequestModel{Hash: hash, ReqType: comm.REQ_HASH_ENABLE}
		}
	case comm.GRPC_SIGN_DISABLE: //禁用
		hash := streamModel.Hash.Hex()
		if !common.HasHexPrefix(hash) || len(common.FromHex(hash)) != comm.HASH_ENABLE_LENGTH {
			log.Error("disallow err")
		}else {
			comm.ReqChan <- &comm.RequestModel{Hash: hash, ReqType: comm.REQ_HASH_DISABLE}
		}

	default:
		log.Info("no type,streamModel:",streamModel)
	}
}
