package db

import (
	logger "github.com/alecthomas/log4go"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/pkg/errors"
)

type Ldb struct {
	*leveldb.DB
}

//init
func InitDb(dbFilePath string) (*Ldb, error) {
	logger.Info("initDb start... path:%v", dbFilePath)
	db, err := leveldb.OpenFile(dbFilePath, &opt.Options{
		OpenFilesCacheCapacity: 16,
		BlockCacheCapacity:     16 / 2 * opt.MiB,
		WriteBuffer:            16 / 4 * opt.MiB,
		Filter:                 filter.NewBloomFilter(10),
	})
	if err != nil {
		return nil, err
	}
	logger.Info("initDb end...")
	return &Ldb{db}, nil
}

func (this *Ldb) GetDb() *leveldb.DB {
	return this.DB
}

func (this *Ldb) PutByte(key, value []byte) error {
	return this.Put(key, value, nil)
}

func (this *Ldb) GetByte(key []byte) ([]byte, error) {
	return this.Get(key, nil)
}

func (this *Ldb) PutStrWithPrifix(keyPri, key, value string) error {
	return this.PutByte([]byte(keyPri+key), []byte(value))
}

//查询前缀
func (this *Ldb) GetPrifix(keyPrefix []byte) (map[string]string, error) {
	var resMap map[string]string = make(map[string]string)
	iter := this.NewIterator(util.BytesPrefix(keyPrefix), nil)
	if iter.Error() == leveldb.ErrNotFound {
		return nil, errors.New("no data")
	}
	if iter.Error() != nil {
		logger.Error("get prifix error")
		return nil, iter.Error()
	}
	for iter.Next() {
		resMap[string(iter.Key())] = string(iter.Value())
	}

	iter.Release()
	return resMap, nil
}

//del key
func (this *Ldb) DelKey(key []byte) error {
	if err := this.Delete(key, nil); err != nil {
		return err
	}
	return nil
}
