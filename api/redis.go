package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gmqtt/orm"
	"net/http"
	"strconv"
)

var RedisAccount redisAccount

type redisAccount struct {
	key string
}

func (r *redisAccount) SetKeyName(username string) {
	r.key = username
}
func (r *redisAccount) Create(setMap map[string]string, expireTime int64) error {
	if _, err := orm.Redis.HSetMap(r.key, setMap); err != nil {
		return err
	}
	if expireTime <= 0 {
		return nil
	}
	return r.Expire(expireTime)
}
func (r *redisAccount) Set(setMap map[string]string) error {
	extTime, err := orm.Redis.Ttl(r.key)
	if err != nil {
		return err
	}
	if extTime <= 0 {
		return errors.New("redis已超时")
	}
	_, err = orm.Redis.HSetMap(r.key, setMap)
	return err
}
func (r *redisAccount) Exists() (bool, error) {
	return orm.Redis.Exists(r.key)
}
func (r *redisAccount) Ttl() (int64, error) {
	return orm.Redis.Ttl(r.key)
}
func (r *redisAccount) Expire(expireTime int64) error {
	return orm.Redis.Expire(r.key, expireTime)
}
func (r *redisAccount) Get() (map[string]string, error) {
	return orm.Redis.HGetAll(r.key)
}
func (r *redisAccount) Delete() error {
	return orm.Redis.Del(r.key)
}

// @Tags                account redis
// @Summary             获取redis帐号信息
// @Produce             json
// @Param 				username 		path 	string 	true 	"帐号"
// @Success             200 	{object} 	SuccessReturn
// @Failure             400 	{object} 	FailReturn
// @Router              /account/redis/{username} [get]
func RedisGet(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数传递错误",
		})
		return
	}

	RedisAccount.SetKeyName(username)
	dataMap, err := RedisAccount.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "OK",
		"data":    dataMap,
	})
	return
}

// @Tags                account redis
// @Summary             创建redis帐号， 如果帐号存在，会清理帐号后创建
// @Produce             json
// @Param 				body 			body 	string 	true 	"json map[string]string { username:帐号, password:密钥, expire_time:过期时间(单位秒), key:value, ...  }"
// @Success             200 	{object} 	SuccessReturn
// @Failure             400 	{object} 	FailReturn
// @Router              /account/redis [post]
func RedisCreate(c *gin.Context) {
	var dataMap map[string]string
	if err := c.ShouldBindJSON(&dataMap); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}
	username, ok := dataMap["username"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "帐号为空",
		})
		return
	}

	if _, ok := dataMap["password"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "密钥为空",
		})
		return
	}

	expire_time := int64(0)
	if value, ok := dataMap["expire_time"]; ok {
		if expire, err := strconv.Atoi(value); err == nil && expire > 0 {
			expire_time = int64(expire)
		}
		delete(dataMap, "expire_time")
	}

	RedisAccount.SetKeyName(username)
	flag, err := RedisAccount.Exists()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}
	if flag {
		if err := RedisAccount.Delete(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    -1,
				"message": err.Error(),
			})
			return
		}
	}

	if err := RedisAccount.Create(dataMap, expire_time); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "OK",
	})
	return
}

// @Tags                account redis
// @Summary             更新redis帐号
// @Produce             json
// @Param 				username 		path 	string 	true 	"帐号"
// @Param 				body 			body 	body 	true 	"json map[string]string { password:密钥, expire_time:过期时间(单位秒), key:value, ...  }"
// @Success             200 	{object} 	SuccessReturn
// @Failure             400 	{object} 	FailReturn
// @Router              /account/redis/{username} [put]
func RedisChange(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数传递错误",
		})
		return
	}

	RedisAccount.SetKeyName(username)

	var dataMap map[string]string
	if err := c.ShouldBindJSON(&dataMap); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	expire_time := int64(0)
	if value, ok := dataMap["expire_time"]; ok {
		if expire, err := strconv.Atoi(value); err == nil && expire > 0 {
			expire_time = int64(expire)
		}
		delete(dataMap, "expire_time")
	}

	flag, err := RedisAccount.Exists()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}
	if !flag {
		c.JSON(http.StatusGone, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	if len(dataMap) > 0 {
		if err := RedisAccount.Set(dataMap); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    -1,
				"message": err.Error(),
			})
			return
		}
	}
	if expire_time > 0 {
		if err := RedisAccount.Expire(expire_time); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    -1,
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "OK",
	})
	return
}

// @Tags                account redis
// @Summary             删除redis帐号
// @Produce             json
// @Param 				username 		path 	string 	true 	"帐号"
// @Success             200 	{object} 	SuccessReturn
// @Failure             400 	{object} 	FailReturn
// @Router              /account/redis/{username} [delete]
func RedisDelete(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数传递错误",
		})
		return
	}

	RedisAccount.SetKeyName(username)

	flag, err := RedisAccount.Exists()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}
	if !flag {
		c.JSON(http.StatusGone, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	if err := RedisAccount.Delete(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "OK",
	})
	return
}
