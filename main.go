package main

import (
	"gmqtt/api"
	"gmqtt/orm"
	"gmqtt/router"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := orm.Config.ReadEnv(); err != nil {
		log.Fatalln("读取配置失败, 请设置 .env 文件, ", err.Error())
	}

	if err := orm.RabbitMQ.Init(orm.Config.RabbitMQ); err != nil {
		log.Fatalln("RabbitMQ", err.Error())
	}
	if err := api.Subscribe(); err != nil {
		log.Fatalln("RabbitMQ", "subscribe", err.Error())
	}

	switch orm.Config.Auth {
	case "redis":
		if err := orm.Redis.Connect(orm.Config.Redis); err != nil {
			log.Fatalln("redis ", err.Error())
		}
	case "mysql":
		if err := orm.MySql.Connect(orm.Config.MySQL); err != nil {
			log.Fatalln("MySql ", err.Error())
		}
	case "wjt":
		orm.Token.Init(orm.Config.WJT)
	default:
		//开放权限
	}

	//进程停止时候运行
	ch := make(chan os.Signal, 1)
	signal.Notify(
		ch,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	)
	go func() {
		s := <-ch
		log.Println("[gin] 停止服务", s)

		orm.Redis.Close()
		orm.RabbitMQ.Close()
		orm.MySql.Close()

		os.Exit(1)
	}()

	router.HttpStart()
	return
}
