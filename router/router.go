package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gmqtt/api"
	_ "gmqtt/docs"
	"gmqtt/orm"
	"net/http"
	"strconv"
)

// 中间件
func AuthWebsocket() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Query("username")
		password := c.Query("password")
		clientId := c.Query("clientId")

		if clientId == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -1,
				"message": "参数传递错误",
			})
			c.Abort()
			return
		}
		c.Set("clientId", clientId)

		switch orm.Config.Auth {
		case "redis":
			if username == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    -1,
					"message": "参数传递错误",
				})
				c.Abort()
				return
			}

			flag, err := orm.Redis.Exists(username)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    -1,
					"message": err.Error(),
				})
				c.Abort()
				return
			}
			if !flag {
				c.JSON(http.StatusGone, gin.H{
					"code":    -1,
					"message": "帐号不存在",
				})
				c.Abort()
				return
			}

			passwordRedis, err := orm.Redis.HGet(username, "password")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    -1,
					"message": err.Error(),
				})
				c.Abort()
				return
			}
			if password != passwordRedis {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    -1,
					"message": "password is wrong",
				})
				c.Abort()
				return
			}
		case "mysql":
			if username == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    -1,
					"message": "参数传递错误",
				})
				c.Abort()
				return
			}
			flag, err := orm.MySql.Check(username, password)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    -1,
					"message": err.Error(),
				})
				c.Abort()
				return
			}

			if !flag {
				c.JSON(http.StatusGone, gin.H{
					"code":    -1,
					"message": "帐号不存在",
				})
				c.Abort()
				return
			}
		case "wjt":
			if username == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    -1,
					"message": "参数传递错误",
				})
				c.Abort()
				return
			}
			if _, err := orm.Token.ParseToken(username); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    -1,
					"message": err.Error(),
				})
				c.Abort()
				return
			}
		default:
			//开放权限
		}
		c.Next()
	}
}

func AuthWJT() gin.HandlerFunc {
	return func(c *gin.Context) {
		if orm.Config.Auth != "wjt" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -1,
				"message": "验证方式错误#" + orm.Config.Auth,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
func AuthRedis() gin.HandlerFunc {
	return func(c *gin.Context) {
		if orm.Config.Auth != "redis" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -1,
				"message": "验证方式错误#" + orm.Config.Auth,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// 开启http服务
func HttpStart() error {
	// http server
	router := gin.Default()

	// api文档
	if orm.Config.Debug {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 服务器状态
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "PONG")
	})

	// 404
	router.NoRoute(func(c *gin.Context) {
		//返回404状态码
		c.JSON(http.StatusNotFound, gin.H{
			"code":    -1,
			"message": "404, page not exists!",
			"data": map[string]string{
				"host":   c.Request.Host,
				"method": c.Request.Method,
				"proto":  c.Request.Proto,
				"path":   c.Request.URL.Path,
				"ip":     c.ClientIP(),
			},
		})
		return
	})

	mqtt := router.Group("mqtt", AuthWebsocket())
	{
		mqtt.GET("", api.Mqtt)
	}

	account := router.Group("account")
	{
		wjt := account.Group("wjt", AuthWJT())
		{
			wjt.GET(":id", api.WjtGet)
		}
		redis := account.Group("redis", AuthRedis())
		{
			redis.POST("", api.RedisCreate)
			redis.GET(":username", api.RedisGet)
			redis.PUT(":username", api.RedisChange)
			redis.DELETE(":username", api.RedisDelete)
		}

	}

	return router.Run(":" + strconv.Itoa(int(orm.Config.Port)))
}
