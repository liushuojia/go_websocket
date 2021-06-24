package orm

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"log"
)

var MySql MySqlConn

type MySqlConf struct {
	Host              string `env:"MYSQL_HOST"`
	Port              int64  `env:"MYSQL_PORT"`
	User              string `env:"MYSQL_USER"`
	Password          string `env:"MYSQL_PASSWD"`
	Database          string `env:"MYSQL_DATABASE"`
	MaxId             int64  `env:"MYSQL_MAX_ID"`              // 空闲最大
	MaxOpen           int64  `env:"MYSQL_MAX_OPEN"`            // 最多连接池
	AuthTable         string `env:"MYSQL_AUTH_TABLE"`          // 验证表
	AuthFieldUsername string `env:"MYSQL_AUTH_FIELD_USERNAME"` // 验证字段username
	AuthFieldPassword string `env:"MYSQL_AUTH_FIELD_PASSWORD"` // 验证字段password
}

type MySqlConn struct {
	Conf MySqlConf
	Conn *gorm.DB
}

func (obj *MySqlConn) Connect(conf MySqlConf) error {
	log.Println("connect mysql")

	obj.Conf = conf
	connArgs := fmt.Sprintf("%s:%s@(%s:%v)/%s?charset=utf8&parseTime=True&loc=Local&timeout=100ms",
		obj.Conf.User, obj.Conf.Password, obj.Conf.Host, obj.Conf.Port, obj.Conf.Database)

	var err error
	obj.Conn, err = gorm.Open("mysql", connArgs)
	if err != nil {
		return err
	}
	if obj.Conn.Error != nil {
		return obj.Conn.Error
	}

	if obj.Conf.MaxId > 0 {
		obj.Conn.DB().SetMaxIdleConns(int(obj.Conf.MaxId))
	}
	if obj.Conf.MaxOpen > 0 {
		obj.Conn.DB().SetMaxOpenConns(int(obj.Conf.MaxOpen))
	}

	obj.Conn.LogMode(true)
	obj.Conn.SingularTable(true)
	return nil
}
func (obj *MySqlConn) Close() {
	if obj.Conn != nil {
		obj.Conn.Close()
	}
}
func (obj *MySqlConn) Check(username, password string) (bool, error) {
	db := obj.Conn.Table(obj.Conf.AuthTable).Select("count(1) as num")
	db = db.Where(obj.Conf.AuthFieldUsername+"=?", username)
	db = db.Where(obj.Conf.AuthFieldPassword+"=?", password)

	var num int = 0
	if err := db.Count(&num).Error; err != nil {
		return false, err
	}
	if num <= 0 {
		return false, nil
	}
	return true, nil
}
