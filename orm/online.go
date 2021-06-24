package orm

import (
	"errors"
	"sync"
)

var OnlineMap Online

type Online struct {
	onlineMap sync.Map
}

func (obj *Online) Clean() {
	obj.onlineMap.Range(func(key, value interface{}) bool {
		obj.Del(key.(string))
		return true
	})
}

func (obj *Online) Del(key string) {
	conn, ok := obj.onlineMap.Load(key)
	if !ok {
		return
	}
	if conn != nil {
		conn.(*Connection).Close()
	}
	obj.onlineMap.Delete(key)
}

func (obj *Online) Set(key string, conn *Connection) error {
	if conn == nil {
		return errors.New("长连接为nil")
	}
	if _, ok := obj.onlineMap.Load(key); ok {
		return errors.New("长连接已注册")
	}
	obj.onlineMap.Store(key, conn)
	return nil
}

func (obj *Online) Exist(key string) (err error) {
	conn, ok := obj.onlineMap.Load(key)
	if !ok {
		return errors.New("长连接已注销")
	}
	if conn == nil {
		return errors.New("长连接已注销")
	}
	return nil
}

func (obj *Online) Get(key string) (conn *Connection, err error) {
	connTmp, ok := obj.onlineMap.Load(key)
	if !ok {
		return nil, errors.New("长连接已注销")
	}
	if connTmp == nil {
		return nil, errors.New("长连接已注销")
	}
	return connTmp.(*Connection), nil
}
