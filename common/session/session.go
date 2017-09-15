package session

import (
	"errors"
	"github.com/go-redis/redis"
	"github.com/lvfeiyang/proxy/common/config"
	"github.com/lvfeiyang/proxy/common/flog"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"strconv"
	"time"
)

const (
	rkIndexes     = "session_indexes"
	rkUsedIndexes = "used_indexes"
	rkRandomN     = "N"
	rkVerifyCode  = "_verify_code"
	rkMobile      = "mobile"
	rkStatus      = "status"
	rkAccount     = "account"
)

const (
	PrefixClientN = "client_n_"
)

const (
	StatusInit    = 0
	StatusNoLogin = 1
	StatusLogin   = 2
)

const CookieKey = "session"

// var client interface{}
// var (
// 	// client = redis.NewClient(&redis.Options{Addr: "172.17.0.1:6379"})
// 	r      = rand.New(rand.NewSource(time.Now().UnixNano()))
// )

func init() {
	redis.SetLogger(flog.LogFile)
}

func ConnRedis() *redis.Client {
	return redis.NewClient(&redis.Options{Addr: config.ConfigVal.RedisUrl})
}

func Init() {
	client := ConnRedis()
	defer client.Close()
	if 0 == client.Exists(rkIndexes).Val() {
		sessionIds := make([]interface{}, 1000)
		var i uint64 = 0
		for ; i < 1000; i++ {
			sessionIds[i] = i + 1
		}
		err := client.SAdd(rkIndexes, sessionIds...).Err()
		if err != nil {
			flog.LogFile.Print(err)
		}
	}
	go tickTime()
}
func tickTime() {
	c := time.Tick(3 * time.Minute)
	for tick := range c {
		client := ConnRedis()
		zr := redis.ZRangeBy{Min: "-inf", Max: strconv.FormatInt(tick.AddDate(0, -1, 0).Unix(), 10)}
		for _, id := range client.ZRangeByScore(rkUsedIndexes, zr).Val() {
			sid, _ := strconv.ParseUint(id, 10, 64)
			clearSess(sid)
		}
		//验证码过期

		client.Close()
	}
}

func updateSessTime(sid uint64) error {
	client := ConnRedis()
	defer client.Close()

	return client.ZAdd(rkUsedIndexes, redis.Z{Score: float64(time.Now().Unix()), Member: sid}).Err()
}

func clearSess(sid uint64) {
	client := ConnRedis()
	defer client.Close()

	client.Del(strconv.FormatUint(sid, 10)).Err()
	client.ZRem(rkUsedIndexes, sid).Err()
	client.SAdd(rkIndexes, sid).Err()
}

type Session struct {
	SessId     uint64
	N          uint64
	VerifyCode uint32
	Mobile     string
	AccountId  string //bson.ObjectId
}

func (s *Session) Get(sid uint64) error {
	client := ConnRedis()
	defer client.Close()

	sidStr := strconv.FormatUint(sid, 10)
	smap, err := client.HGetAll(sidStr).Result()
	if err != nil {
		return err
	}
	if len(smap) == 0 {
		return errors.New("get empty session!")
	}

	s.N, _ = strconv.ParseUint(smap[rkRandomN], 10, 64)
	// s.VerifyCode, _ = strconv.ParseUint(smap[rkVerifyCode], 10, 32)
	vc, _ := client.Get(sidStr + rkVerifyCode).Uint64()
	s.VerifyCode = uint32(vc)
	s.Mobile, _ = smap[rkMobile]
	if account, r := smap[rkAccount]; r {
		s.AccountId = account //bson.ObjectIdHex(account)
	}

	return updateSessTime(sid)
}
func (s *Session) SetVerifyCode() (uint32, error) {
	if 0 == s.VerifyCode {
		client := ConnRedis()
		defer client.Close()

		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		s.VerifyCode = uint32(r.Intn(1000000))
		if _, err := client.SetNX(strconv.FormatUint(s.SessId, 10)+rkVerifyCode, s.VerifyCode, 10*time.Minute).Result(); err != nil {
			return 0, err
		}
	}
	return s.VerifyCode, nil
}
func (s *Session) SetMobile(mobile string) error {
	client := ConnRedis()
	defer client.Close()

	s.Mobile = mobile
	if err := client.HSet(strconv.FormatUint(s.SessId, 10), rkMobile, s.Mobile).Err(); err != nil {
		return err
	}
	return nil
}
func (s *Session) SetAccount(id bson.ObjectId) error {
	client := ConnRedis()
	defer client.Close()

	s.AccountId = id.Hex()
	if err := client.HSet(strconv.FormatUint(s.SessId, 10), rkAccount, s.AccountId).Err(); err != nil { //.Hex()
		return err
	}
	return nil
}
func (s *Session) Apply() uint64 {
	client := ConnRedis()
	defer client.Close()

	id, err := client.SPop(rkIndexes).Uint64() //Result()
	if err == redis.Nil {
		flog.LogFile.Println("no empty index")
	} else if err != nil {
		flog.LogFile.Println(err)
	} else {
		s.N = rand.New(rand.NewSource(time.Now().UnixNano())).Uint64()
		err := client.HSet(strconv.FormatUint(id, 10), rkRandomN, s.N).Err()
		if err != nil {
			flog.LogFile.Println(err)
		}
		if err = updateSessTime(id); err != nil {
			flog.LogFile.Println(err)
		}
		return id
	}
	return 0
}

//redis cache
func PutCache(key, in string, expires uint) error {
	client := ConnRedis()
	defer client.Close()
	_, err := client.Set(key, in, time.Duration(expires)*time.Second).Result()
	return err
}
func GetCache(key string) (string, error) {
	client := ConnRedis()
	defer client.Close()
	return client.Get(key).Result()
}
