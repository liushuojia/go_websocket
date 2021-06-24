package orm

import (
	"context"
	"errors"
	"github.com/go-ini/ini"
	"github.com/sethvargo/go-envconfig"
	"reflect"
	"strconv"
	"time"
)

var Config config

type config struct {
	Name  string `env:"HOST_NAME" default:"gmqtt"`
	Port  int64  `env:"HOST_PORT" default:"8000"`
	Debug bool   `env:"HOST_DEBUG" default:"true"`
	Auth  string `env:"HOST_AUTH"`

	MySQL    MySqlConf
	Redis    RedisConf
	RabbitMQ RabbitMQConf
	WJT      WJTConf
}

func (c *config) ReadEnv() error {
	if err := c.ReadLinuxEnv(); err != nil {
		return err
	}
	if err := c.ReadFileEnv(); err != nil {
		return err
	}

	c.Redis.IdleTimeout = c.Redis.IdleTimeout * time.Second

	return nil
}

func (c *config) ReadFileEnv() error {
	envFile := ".env"
	cfg, err := ini.Load(envFile)
	if err != nil {
		return err
	}
	EnvMap := cfg.Section("").KeysHash()

	c.setValueForMap(c, EnvMap)
	c.setValueForMap(&c.MySQL, EnvMap)
	c.setValueForMap(&c.Redis, EnvMap)
	c.setValueForMap(&c.RabbitMQ, EnvMap)
	c.setValueForMap(&c.WJT, EnvMap)
	return nil
}

func (c *config) setValueForMap(obj interface{}, EnvMap map[string]string) error {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() != reflect.Ptr {
		return errors.New("请传入指针变量")
	}
	t = t.Elem()
	v = v.Elem()
	if t.Kind() != reflect.Struct {
		return errors.New("请传入指针变量")
	}

	for i := 0; i < t.NumField(); i++ {
		if !v.Field(i).IsZero() || !v.Field(i).IsValid() {
			continue
		}
		envKey := t.Field(i).Tag.Get("env")
		if envKey == "" {
			continue
		}
		value, ok := EnvMap[envKey]
		if !ok || value == "" {
			continue
		}
		switch v.Field(i).Kind() {
		case reflect.String:
			v.Field(i).SetString(value)
		case reflect.Int64:
			if num, err := strconv.Atoi(value); err == nil && num > 0 {
				v.Field(i).SetInt(int64(num))
			}
		case reflect.Int:
			if num, err := strconv.Atoi(value); err == nil && num > 0 {
				v.Field(i).SetInt(int64(num))
			}
		case reflect.Bool:
			if flag, err := strconv.ParseBool(value); err == nil {
				v.Field(i).SetBool(flag)
			}
		}
	}
	return nil
}

func (c *config) ReadLinuxEnv() error {
	ctx := context.Background()
	if err := envconfig.Process(ctx, c); err != nil {
		return err
	}
	return nil
}
