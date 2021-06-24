package api

import (
	"github.com/gin-gonic/gin"
	"gmqtt/orm"
	"net/http"
	"strconv"
)

// @Tags                account wjt
// @Summary             获取 web json token
// @Produce             json
// @Param 				id 		path 	int 	true 	"唯一id"
// @Success             200 	{object} 	SuccessReturn
// @Failure             400 	{object} 	FailReturn
// @Router              /account/wjt/{id} [get]
func WjtGet(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数传递错误",
		})
		return
	}
	if id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数传递错误",
		})
		return
	}

	orm.Token.Claims.ID = id
	token, err := orm.Token.GenerateToken()
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
		"data":    token,
	})
	return
}
