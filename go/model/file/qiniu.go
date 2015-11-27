package file

import (
	"fmt"
	"time"

	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/api/conf"
	"github.com/qiniu/api/rs"
	"github.com/satori/uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"sirendaou.com/duserver/common"
	"sirendaou.com/duserver/common/errors"
	"sirendaou.com/duserver/common/sync"
	"sirendaou.com/duserver/common/syslog"
)

var (
	DOWNLOAD_URL = "7xlnqu.com2.z0.glb.qiniucdn.com"
	//g_upload_url    = "http://up.qiniu.com"
	g_access_key    = ""
	g_secret_key    = ""
	CALLBACK_URL    = ""
	g_callback_body = ""
	g_return_url    = ""
	g_return_body   = ""
	g_bucket_name   = "duserver"

	g_file_mgr *FileMgr = nil

	DEFAULT_EXPIRED_TIME = 30 * 24 * 3600
)

func init() {
	conf.ACCESS_KEY = "9cY7epZMkbljif2XZMuphvMADeAhdZnP5T3yYbxC"
	conf.SECRET_KEY = "4VcM0t3uz0dNfLMfsn6D4Z-CrB2Co5ksnNO9uTnm"
	g_access_key = conf.ACCESS_KEY
	g_secret_key = conf.SECRET_KEY
}

type FileMgr struct {
	fileCh       chan *File
	expiredTimer <-chan time.Time
	waitGroup    *sync.WaitGroup
	rsClient     rs.Client
}

type File struct {
	Filename    string
	ExpiredTime int64
	CreateTime  string
}

func QiniuInit() error {
	c := common.MongoCollection(FILE_DB, FILE_QINIU_FILE_TABLE)

	index := mgo.Index{
		Key:      []string{"filename"},
		Unique:   true,
		DropDups: true,
	}

	g_file_mgr = &FileMgr{
		fileCh:       make(chan *File, 100),
		expiredTimer: time.Tick(time.Hour),
		waitGroup:    sync.NewWaitGroup(),
		rsClient:     rs.New(&digest.Mac{AccessKey: g_access_key, SecretKey: []byte(g_secret_key)}),
	}

	go g_file_mgr.run()

	return errors.As(c.EnsureIndex(index))
}

func addFile(file *File) error {
	c := common.MongoCollection(FILE_DB, FILE_QINIU_FILE_TABLE)
	sel := bson.M{"filename": file.Filename}
	set := bson.M{
		"$set":         bson.M{"expiredtime": file.ExpiredTime},
		"$setOnInsert": bson.M{"createtime": file.CreateTime},
	}
	_, err := c.Upsert(sel, set)
	return errors.As(err, *file)
}

func deleteFile(filelist []string) error {
	c := common.MongoCollection(FILE_DB, FILE_QINIU_FILE_TABLE)
	_, err := c.RemoveAll(bson.M{"filename": bson.M{"$in": filelist}})
	if err != nil && err != mgo.ErrNotFound {
		return errors.As(err, filelist)
	}
	return nil
}

func GetToken() (string, string) {
	key := uuid.NewV4().String()
	putPolicy := rs.PutPolicy{
		Scope:        g_bucket_name + ":" + key,
		CallbackUrl:  CALLBACK_URL,
		CallbackBody: g_callback_body,
		ReturnUrl:    g_return_url,
		ReturnBody:   g_return_body,
	}

	return putPolicy.Token(nil), key
}

func GetTokenByArgs(fmtStr string, args ...interface{}) (string, string) {
	body := fmt.Sprintf(fmtStr, args...)
	g_callback_body = "hash=$(etag)&key=$(key)"
	if len(body) > 0 {
		g_callback_body += "&" + body
	}

	return GetToken()
}

func GetUrl(key string) string {
	baseUrl := rs.MakeBaseUrl(DOWNLOAD_URL, key)
	getPolicy := rs.GetPolicy{}
	return getPolicy.MakeRequest(baseUrl, nil)
}

func SetFileExpiredTime(key string, expiredTime int64) error {
	file := &File{
		Filename:    key,
		ExpiredTime: time.Now().Unix() + expiredTime,
		CreateTime:  time.Now().Format(common.DATETIME_FMT),
	}
	g_file_mgr.fileCh <- file
	return nil
}

func QiniuDeleteFiles(keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	entryPathes := []rs.EntryPath{}
	for _, key := range keys {
		entryPathes = append(entryPathes, rs.EntryPath{Bucket: g_bucket_name, Key: key})
	}
	_, err := g_file_mgr.rsClient.BatchDelete(nil, entryPathes)
	if err != nil {
		syslog.Info(err, keys)
	}
	deleteFile(keys)
	return nil
}

func (this *FileMgr) run() {
	exitNotify := this.waitGroup.ExitNotify()
	for {
		select {
		case <-exitNotify:
			return
		case <-this.expiredTimer:
			if err := this.RemoveExpiredFile(); err != nil {
				syslog.Warn(err)
			}
		case file := <-this.fileCh:
			if err := addFile(file); err != nil {
				syslog.Warn(err, *file)
			}

		}
	}
}

func (this *FileMgr) RemoveExpiredFile() error {
	c := common.MongoCollection(FILE_DB, FILE_QINIU_FILE_TABLE)

	expiredTime := time.Now().Unix()
	iter := c.Find(bson.M{"expiredtime": bson.M{"$lt": expiredTime}}).Iter()
	defer iter.Close()
	file := &File{}
	keys := []string{}
	for iter.Next(file) {
		keys = append(keys, file.Filename)
	}
	QiniuDeleteFiles(keys)
	return nil
}
