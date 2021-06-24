package orm

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"time"
	"unsafe"
)

var Redis RedisConn

type RedisConf struct {
	Host        string        `env:"REDIS_HOST"`
	Port        int64         `env:"REDIS_PORT"`
	Password    string        `env:"REDIS_PASSWD"`
	Db          int64         `env:"REDIS_DB"`
	Prefix      string        `env:"REDIS_PREFIX"`       // redis 前缀 + 模块
	MaxActive   int64         `env:"REDIS_MAX_ACTIVE"`   // 最大连接数，即最多的tcp连接数
	MaxIdle     int64         `env:"REDIS_MAX_IDLE"`     // 最大空闲连接数，即会有这么多个连接提前等待着，但过了超时时间也会关闭。
	IdleTimeout time.Duration `env:"REDIS_IDLE_TIMEOUT"` // 超时时间
}

type RedisConn struct {
	conf   RedisConf
	conn   *redis.Pool
	client redis.PubSubConn
	cbMap  map[string]SubscribeCallback
}

func (c *RedisConn) Connect(conf RedisConf) error {
	// 需要公司编码, 否则报错
	log.Println("connect redis")
	c.conf = conf
	// 建立连接池
	c.conn = &redis.Pool{
		MaxIdle:     int(c.conf.MaxIdle),
		MaxActive:   int(c.conf.MaxActive),
		IdleTimeout: c.conf.IdleTimeout,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			con, err := redis.Dial(
				"tcp",
				fmt.Sprintf("%s:%v", c.conf.Host, c.conf.Port),
				redis.DialPassword(c.conf.Password),
				redis.DialDatabase(int(c.conf.Db)),
				redis.DialConnectTimeout(c.conf.IdleTimeout),
				redis.DialReadTimeout(c.conf.IdleTimeout),
				redis.DialWriteTimeout(c.conf.IdleTimeout),
			)
			if err != nil {
				return nil, err
			}
			return con, nil
		},
	}

	return c.Status()
}
func (c *RedisConn) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
	if c.client.Conn != nil {
		c.client.Close()
	}
}
func (c *RedisConn) GetConn() (redis.Conn, error) {
	rc := c.conn.Get()
	if err := rc.Err(); err != nil {
		return nil, err
	}
	return rc, nil
}
func (c *RedisConn) Status() error {
	rc, err := c.GetConn()
	if err != nil {
		return err
	}
	defer rc.Close()
	_, err = redis.String(rc.Do("PING"))
	return err
}

func (c *RedisConn) exec(cmd string, key string, args ...interface{}) (interface{}, error) {
	rc, err := c.GetConn()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	parmas := make([]interface{}, 0)
	parmas = append(parmas, c.conf.Prefix+key)
	if len(args) > 0 {
		for _, v := range args {
			parmas = append(parmas, v)
		}
	}

	return rc.Do(cmd, parmas...)
}
func (c RedisConn) Set(key string, rdVal string, rdExTime int64) error {
	replay, err := c.exec("set", key, rdVal)
	if err != nil {
		return err
	}
	if replay.(string) != "OK" {
		return errors.New("设置key值失败")
	}
	if rdExTime > 0 {
		return c.Expire(key, rdExTime)
	}
	return nil
}
func (c RedisConn) Get(key string) (string, error) {
	// 增加前缀
	exit_bool, err := c.Exists(key)
	if err != nil {
		return "", err
	}
	if !exit_bool {
		return "", errors.New("无该key值")
	}
	return redis.String(c.exec("GET", key))
}
func (c RedisConn) Exists(key string) (bool, error) {
	return redis.Bool(c.exec("EXISTS", key))
}
func (c RedisConn) Del(key string) error {
	_, err := c.exec("DEL", key)
	return err
}
func (c RedisConn) Ttl(key string) (int64, error) {
	return redis.Int64(c.exec("TTL", key))
}
func (c RedisConn) Expire(key string, rdExTime int64) error {
	replay, err := redis.Int64(c.exec("Expire", key, rdExTime))
	if err != nil {
		return err
	}
	if replay != 1 {
		return errors.New("key 不存在#" + key)
	}
	return nil
}

//返回所有元素
func (c RedisConn) Time() ([]string, error) {
	rc, err := c.GetConn()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return redis.Strings(rc.Do("TIME"))
}

/*

	队列处理

*/

func (c RedisConn) LPush(key string, val string) (int64, error) {
	return redis.Int64(c.exec("lPush", key, val))
}
func (c RedisConn) LPop(key string) (string, error) {
	return redis.String(c.exec("LPOP", key))
}
func (c RedisConn) RPush(key string, val string) (int64, error) {
	return redis.Int64(c.exec("rPush", key, val))
}
func (c RedisConn) RPop(key string) (string, error) {
	return redis.String(c.exec("rPOP", key))
}
func (c RedisConn) LLen(key string) (int64, error) {
	return redis.Int64(c.exec("LLEN", key))
}

/*
	集合处理
*/

//增加值  return int64 1 创建新值  0 值已存在
func (c RedisConn) SAdd(key string, val string) (int64, error) {
	return redis.Int64(c.exec("SADD", key, val))
}

//集合长度
func (c RedisConn) SCard(key string) (int64, error) {
	return redis.Int64(c.exec("SCARD", key))
}

//集合移除元素  return int64 1 删除成功  0 无该值
func (c RedisConn) SRem(key string, val string) (int64, error) {
	return redis.Int64(c.exec("SREM", key, val))
}

//随机返回一个元素
func (c RedisConn) SRandmember(key string) (string, error) {
	return redis.String(c.exec("SRANDMEMBER", key))
}

//返回所有元素
func (c RedisConn) SMembers(key string) ([]string, error) {
	return redis.Strings(c.exec("SMEMBERS", key))
}

//判断集合是否存在元素
func (c RedisConn) SIsmember(key string, val string) (bool, error) {
	return redis.Bool(c.exec("SISMEMBER", key, val))
}

/*
	哈希处理
*/
//增加值  return int64 1 创建新值  0 值已存在,覆盖
func (c RedisConn) HSet(key string, field string, value string) (int64, error) {
	return redis.Int64(c.exec("HSET", key, field, value))
}
func (c RedisConn) HSetMap(key string, fieldsMap map[string]string) (int64, error) {
	var fields []interface{}
	for k, v := range fieldsMap {
		fields = append(fields, k, v)
	}
	return redis.Int64(c.exec("HSET", key, fields...))
}
func (c RedisConn) HSetNX(key string, field string, value string) error {
	replay, err := redis.Int64(c.exec("HSETNX", key, field, value))
	if err != nil {
		return err
	}
	if replay == 0 {
		return errors.New("添加失败,哈希中已存在字段")
	}
	return nil
}
func (c RedisConn) HGetAll(key string) (map[string]string, error) {
	return redis.StringMap(c.exec("HGETALL", key))
}
func (c RedisConn) HKeys(key string) ([]string, error) {
	return redis.Strings(c.exec("HKEYS", key))
}

//删除值  return int64 1 值存在并删除  0 值不存在
func (c RedisConn) HDel(key string, field string) (int64, error) {
	return redis.Int64(c.exec("HDEL", key, field))
}
func (c RedisConn) HExists(key string, field string) (bool, error) {
	return redis.Bool(c.exec("HEXISTS", key, field))
}
func (c RedisConn) HGet(key string, field string) (string, error) {
	return redis.String(c.exec("HGET", key, field))
}
func (c RedisConn) HLen(key string) (int64, error) {
	return redis.Int64(c.exec("HLEN", key))
}

/*
	订阅设计
*/

//发布订阅消息
func (c RedisConn) PublicMsg(channel, msg string) bool {
	conn, err := c.GetConn()
	if err != nil {
		return false
	}
	defer conn.Close()

	conn.Do("Publish", channel, msg)
	return true
}

//订阅reids
func (c *RedisConn) SubConnect() (redis.Conn, error) {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%v", c.conf.Host, c.conf.Port))
	if err != nil {
		return nil, err
	}
	if c.conf.Password != "" {
		_, err = conn.Do("AUTH", c.conf.Password)
		if err != nil {
			conn.Close()
			return nil, err
		}
	}

	_, err = conn.Do("SELECT", int(c.conf.Db))
	if err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}
func (c *RedisConn) SubInit() {
	if c.cbMap != nil {
		//进程只有一个, 如果已经初始化了, 这里不做什么
		return
	}

	conn, err := c.SubConnect()
	if err != nil {
		log.Println("[subscribe]", "redis sub", err.Error())
		return
	}

	c.client = redis.PubSubConn{conn}
	c.cbMap = make(map[string]SubscribeCallback)
	go func() {
		for {
			switch res := c.client.Receive().(type) {
			case redis.Message:
				channel := (*string)(unsafe.Pointer(&res.Channel))
				message := (*string)(unsafe.Pointer(&res.Data))
				c.cbMap[*channel](*message)
			case redis.Subscription:
				log.Println("[subscribe]", "redis", res.Channel, ", 共有", res.Count, "个订阅")
			case error:
				log.Println("[subscribe]", "redis reconnect", res.Error())

				//redis 服务器出错, 10s试着重新连接
				time.Sleep(reconnectDelay)
				if conn, err := c.SubConnect(); err == nil {
					c.client = redis.PubSubConn{conn}
					log.Println("[subscribe]", "redis reconnect success")
				}
			}
		}
	}()
}
func (c *RedisConn) Subscribe(channel interface{}, cb SubscribeCallback) {
	err := c.client.Subscribe(channel)
	if err != nil {
		log.Println("redis Subscribe error.", err.Error())
		return
	}

	c.cbMap[channel.(string)] = cb
}
