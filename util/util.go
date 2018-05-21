package util

import (
	"bytes"

	"github.com/boxproject/companion/comm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"net"
	"math/big"
	"os"
	"sync"
	log "github.com/alecthomas/log4go"
	"io/ioutil"
)

//byte[] --> byte[32] 不校验长度，已校验过
func Byte2Byte32(src []byte) [32]byte {
	var obj [32]byte
	copy(obj[:], src[:32])
	return obj
}

//检测转账类型
func CheckCategory(category int64) bool {
	return category >= comm.MIN_CATEGORY && category <= comm.MAX_CATEGORY
}

func AddressEquals(a, b common.Address) bool {
	return bytes.Equal(a.Bytes(), b.Bytes())
}

func String2Hash(hashStr string) common.Hash {
	return common.BytesToHash([]byte(hashStr))
}

func HexToHash(hashStr string) common.Hash {
	return common.HexToHash(hashStr)
}

func GetRecAddress(req comm.RequestModel)(common.Address, error){
	if req.Category == comm.CATEGORY_BTC {
		//btc提取20byte的hash
		if btcPubAddr, err := btcutil.DecodeAddress(req.RecAddress,&chaincfg.MainNetParams); err != nil {
			return  common.Address{},err
		}else {
			return common.BytesToAddress(btcPubAddr.ScriptAddress()),nil
		}

	}else {
		return common.HexToAddress(req.RecAddress),nil
	}
}

func GetCurrentIp() string {
	addrSlice, err := net.InterfaceAddrs()
	if nil != err {
		//logger.Error("Get local IP addr failed!!!")
		return "localhost"
	}
	for _, addr := range addrSlice {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if nil != ipnet.IP.To4() {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}

//file
var noRWMutex sync.RWMutex
func WriteNumberToFile(filePath string, blkNumber *big.Int) error {
	noRWMutex.Lock()
	defer noRWMutex.Unlock()
	return ioutil.WriteFile(filePath, []byte(blkNumber.String()), 0755)
}

func ReadNumberFromFile(filePath string) (*big.Int, error) {
	noRWMutex.Lock()
	defer noRWMutex.Unlock()
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Debug("file not found, %v", err)
		return big.NewInt(0), nil
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	data = bytes.TrimSpace(data)
	delta, _ := big.NewInt(0).SetString(string(data), 10)
	return delta, nil
	}