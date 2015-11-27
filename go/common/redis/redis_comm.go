package redis

import (
	"strings"
	"time"

	"github.com/vmihailenco/redis/v2"
	"sirendaou.com/duserver/common/errors"
	"sirendaou.com/duserver/common/syslog"
)

type RedisItem struct {
	Client    *redis.Client
	RedisAddr string
}

func CreateRedisItem(addr string) *RedisItem {
	client := redis.NewTCPClient(&redis.Options{
		Addr: addr,
	})

	if err := client.Set("test", "test").Err(); err != nil {
		panic(err)
	}

	return &RedisItem{client, addr}
}

func ReconnectRedis(pclient *RedisItem) {
	pclient.Client.Close()
	pclient.Client = CreateRedisItem(pclient.RedisAddr).Client
	return
}

type RedisManager struct {
	RedisCh chan *RedisItem
	count   int
}

var g_redis *RedisManager = nil

func Init(addrs string, connCountPerAddr int) {
	addrSlice := strings.Split(addrs, ",")
	poolsize := connCountPerAddr * len(addrSlice)
	AliasRediPool := make(chan *RedisItem, poolsize)

	for _, addr := range addrSlice {
		for i := 0; i < connCountPerAddr; i++ {
			it := CreateRedisItem(addr)
			AliasRediPool <- it
		}
	}

	g_redis = &RedisManager{AliasRediPool, poolsize}
}

func Deinit() {
	for i := 0; i < g_redis.count; i++ {
		c := <-g_redis.RedisCh
		c.Client.Close()
	}
}

func get() *RedisItem {
	client := <-g_redis.RedisCh
	return client
}
func put(ri *RedisItem) {
	g_redis.RedisCh <- ri
}
func RedisSet(key string, value string) error {
	rClient := get()
	defer put(rClient)

	if _, err := rClient.Client.Set(key, value).Result(); err != nil {
		return errors.As(err, "redis set failed:", key, value)
	}
	return nil
}

func Set(key string, value string) error {
	rClient := get()
	defer put(rClient)

	if _, err := rClient.Client.Set(key, value).Result(); err != nil {
		return errors.As(err, "redis set failed:", key, value)
	}
	return nil
}
func RedisDel(key string) error {
	rClient := get()
	defer put(rClient)

	if _, err := rClient.Client.Del(key).Result(); err != nil {
		return errors.As(err, "redis del key failed:", key)
	}

	return nil
}

func Del(key string) error {
	rClient := get()
	defer put(rClient)

	if _, err := rClient.Client.Del(key).Result(); err != nil {
		return errors.As(err, "redis del key failed:", key)
	}

	return nil
}

func RedisSetEx(key string, dur time.Duration, value string) error {
	rClient := get()
	defer put(rClient)

	if _, err := rClient.Client.SetEx(key, dur, value).Result(); err != nil {
		return errors.As(err, "redis setex failed:", key, dur, value)
	}

	return nil
}

func SetEx(key string, dur time.Duration, value string) error {
	rClient := get()
	defer put(rClient)

	if _, err := rClient.Client.SetEx(key, dur, value).Result(); err != nil {
		return errors.As(err, "redis setex failed:", key, dur, value)
	}

	return nil
}

func Get(key string) (string, error) {
	rClient := get()
	defer put(rClient)

	v, err := rClient.Client.Get(key).Result()
	if err != nil && err != redis.Nil {
		return v, errors.As(err, "redis get failed:", key)
	}
	if err == redis.Nil {
		return v, errors.ERR_NO_DATA
	}
	return v, nil
}
func RedisGet(key string) (string, error) {
	rClient := get()
	defer put(rClient)

	v, err := rClient.Client.Get(key).Result()
	if err != nil && err != redis.Nil {
		return v, errors.As(err, "redis set failed:", key)
	}
	if err == redis.Nil {
		return v, errors.ERR_NO_DATA
	}
	return v, nil
}

func RedisHGet(key string, field string) (string, error) {
	rClient := get()
	defer put(rClient)

	v, err := rClient.Client.HGet(key, field).Result()
	if err != nil && err != redis.Nil {
		return v, errors.As(err, "redis hget failed:", key, field)
	}

	return v, nil
}

func RedisMGet(keys []string) ([]interface{}, error) {
	rClient := get()
	defer put(rClient)

	v, err := rClient.Client.MGet(keys...).Result()
	if err != nil && err != redis.Nil {
		return v, errors.As(err, "redis mget failed:", keys)
	}

	return v, nil
}

func RedisRPop(key string) (string, error) {
	rClient := get()
	defer put(rClient)

	v, err := rClient.Client.RPop(key).Result()
	if err != nil && err != redis.Nil {
		return v, errors.As(err, "redis rpop failed:", key)
	}

	return v, nil
}

func RedisLPush(key, val string) {
	rClient := get()
	defer put(rClient)

	rClient.Client.LPush(key, val)
	return
}

func RedisSAdd(key, val string) {
	rClient := get()
	defer put(rClient)

	rClient.Client.SAdd(key, val)
	return
}

func RedisSDel(key, val string) {
	rClient := get()
	defer put(rClient)

	rClient.Client.SRem(key, val)
	return
}

func PipelineGetString(keys []string) []string {
	rClient := get()
	defer put(rClient)

	pipeline := rClient.Client.Pipeline()

	keysNum := 0
	for _, key := range keys {
		if len(key) > 2 {
			syslog.Debug("pipeline key ", key)
			pipeline.Get(key)
			keysNum++
		}
	}

	cmds, err := pipeline.Exec()

	syslog.Debug("pipe result:", cmds, err)

	result := ""

	valList := make([]string, keysNum)
	n := 0
	if err != nil && err != redis.Nil {
		syslog.Debug("redisClient pipeline err %s", err.Error())

		ReconnectRedis(rClient)
	} else {
		syslog.Debug("redisClient pipeline ok")

		for _, cmd := range cmds {
			syslog.Debug(cmd, " ret result:", cmd.(*redis.StringCmd).Val())
			if cmd.(*redis.StringCmd).Err() != nil {
				syslog.Error(cmd.(*redis.StringCmd).Err())
			} else {
				result = cmd.(*redis.StringCmd).Val()
			}
			if len(result) > 0 {
				syslog.Debug(n, result)
				valList[n] = result
				n++
			}
		}
	}

	return valList
}

func RedisZRange(key string) ([]string, error) {
	rClient := get()
	defer put(rClient)

	val, err := rClient.Client.ZRange(key, 0, -1).Result()
	if err != nil && err != redis.Nil {
		return val, errors.As(err, "redis zrange failed:", key)
	}

	return val, nil
}

func RedisZRange2(key string, cnt int) ([]string, error) {
	if cnt < 1 {
		return []string{}, nil
	}

	rClient := get()
	defer put(rClient)

	val, err := rClient.Client.ZRange(key, 0, int64(cnt-1)).Result()
	if err != nil && err != redis.Nil {
		return val, errors.As(err, "redis zrange failed:", key)
	}

	return val, nil
}

func RedisZRem(key, val string) error {
	rClient := get()
	defer put(rClient)

	if _, err := rClient.Client.ZRem(key, val).Result(); err != nil && err != redis.Nil {
		return errors.As(err, "redis zrem failed:", key, val)
	}

	return nil
}

func SIsMember(key, val string) *redis.BoolCmd {
	rClient := get()
	defer put(rClient)

	return rClient.Client.SIsMember(key, val)
}

func SMembers(key string) *redis.StringSliceCmd {
	rClient := get()
	defer put(rClient)

	return rClient.Client.SMembers(key)
}

func SRem(key string, members ...string) *redis.IntCmd {
	rClient := get()
	defer put(rClient)

	return rClient.Client.SRem(key, members...)
}

func SCard(key string) *redis.IntCmd {
	rClient := get()
	defer put(rClient)

	return rClient.Client.SCard(key)
}

func ZCard(key string) *redis.IntCmd {
	rClient := get()
	defer put(rClient)

	return rClient.Client.ZCard(key)
}

type Z struct {
	Score  float64
	Member string
}

func ZAdd(key string, members ...Z) *redis.IntCmd {
	rClient := get()
	defer put(rClient)

	zs := make([]redis.Z, len(members))
	for i, m := range members {
		zs[i].Score = m.Score
		zs[i].Member = m.Member
	}
	return rClient.Client.ZAdd(key, zs...)
}

type ZRangeByScoreT struct {
	Min, Max      string
	Offset, Count int64
}

func ZRangeByScore(key string, opt ZRangeByScoreT) *redis.StringSliceCmd {
	rClient := get()
	defer put(rClient)

	return rClient.Client.ZRangeByScore(key, redis.ZRangeByScore{
		Min:    opt.Min,
		Max:    opt.Max,
		Offset: opt.Offset,
		Count:  opt.Count,
	})
}

func ZRemRangeByRank(key string, start, stop int64) *redis.IntCmd {
	rClient := get()
	defer put(rClient)

	return rClient.Client.ZRemRangeByRank(key, start, stop)
}
