package api

import (
	"encoding/json"
	"gmqtt/orm"
	"reflect"
	"sync"
)

func Subscribe() error {
	orm.RabbitMQ.Subscribe(
		orm.Config.RabbitMQ.Subscribe,
		"channel",
		func(data string) {
			var onelineMessage orm.OnelineMessage
			if err := json.Unmarshal([]byte(data), &onelineMessage); err != nil {
				return
			}
			cbMapTmp, ok := CallBackMap.Load(onelineMessage.Topic)
			if !ok || cbMapTmp == nil {
				return
			}
			if reflect.TypeOf(cbMapTmp).String() != "map[string]orm.SubscribeCallback" {
				return
			}
			cbMap := cbMapTmp.(map[string]orm.SubscribeCallback)
			if onelineMessage.ClientId == "*" {
				for _, cb := range cbMap {
					cb(data)
				}
			} else {
				if cb, ok := cbMap[onelineMessage.ClientId]; ok {
					cb(data)
				}
			}
		},
	)
	return orm.RabbitMQ.SubscribeRun()
}
func Publish(body string) error {
	return orm.RabbitMQ.Publish(orm.Publish{
		Name: orm.Config.RabbitMQ.Publish,
		Kind: "channel",
		Key:  "",
		Body: body,
	})
}

var CallBackMap sync.Map // map[topic] map[clientId]orm.SubscribeCallback
func SubscribeAdd(onelineMessage orm.OnelineMessage, cb orm.SubscribeCallback) {
	cbMapTmp, ok := CallBackMap.Load(onelineMessage.Topic)
	cbMap := make(map[string]orm.SubscribeCallback)
	if ok && reflect.TypeOf(cbMapTmp).String() == "map[string]orm.SubscribeCallback" {
		cbMap = cbMapTmp.(map[string]orm.SubscribeCallback)
	}
	cbMap[onelineMessage.ClientId] = cb
	CallBackMap.Store(onelineMessage.Topic, cbMap)
}
func SubscribeDelete(onelineMessage orm.OnelineMessage) {
	cbMapTmp, ok := CallBackMap.Load(onelineMessage.Topic)
	if !ok {
		return
	}
	if reflect.TypeOf(cbMapTmp).String() != "map[string]orm.SubscribeCallback" {
		return
	}
	cbMap := cbMapTmp.(map[string]orm.SubscribeCallback)
	delete(cbMap, onelineMessage.ClientId)
	CallBackMap.Store(onelineMessage.Topic, cbMap)
}
