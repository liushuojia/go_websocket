package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gmqtt/orm"
	"net/http"
)

/*

	使用 rabbitMQ 推送消息
	使用 redis 记录在线用户

*/

// @Tags                mqtt链接
// @Summary             连接长链接
// @Produce             json
// @Param 				username 	query 	string 	true 	"帐号/token"
// @Param 				password 	query 	string 	false 	"密码"
// @Param 				clientId 	query 	string 	true 	"clientId"
// @Success             200 	{object} 	SuccessReturn
// @Failure             400 	{object} 	FailReturn
// @Router              /mqtt [get]
func Mqtt(c *gin.Context) {
	ws, err := orm.InitConnection(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}
	defer ws.Close()

	clientId := c.GetString("clientId")
	if err := orm.OnlineMap.Exist(clientId); err == nil {
		return
	}

	if err := orm.OnlineMap.Set(clientId, ws); err != nil {
		return
	}
	defer orm.OnlineMap.Del(clientId)

	for {
		j, err := ws.ReadMessage()
		if err != nil {
			goto END
		}
		var onelineMessage orm.OnelineMessage
		if err := json.Unmarshal(j, &onelineMessage); err != nil {
			continue
		}
		onelineMessage.ClientId = clientId

		switch onelineMessage.Action {
		case "publish":
			if err := Publish(string(j)); err != nil {
				goto END
			}
		case "subscribe":
			SubscribeAdd(onelineMessage, func(s string) {
				ws.WriteMessage([]byte(s))
			})
		case "unsubscribe":
			SubscribeDelete(onelineMessage)
		}
	}

END:
	fmt.Println("close")
	ws.Close()
}
