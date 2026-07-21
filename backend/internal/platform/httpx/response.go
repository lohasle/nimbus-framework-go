package httpx

import "github.com/gin-gonic/gin"

type Response struct {
	Code int    `json:"code" example:"0"`
	Data any    `json:"data"`
	Msg  string `json:"msg" example:""`
}

func OK(c *gin.Context, data any) { c.JSON(200, Response{Code: 0, Data: data, Msg: ""}) }

func Fail(c *gin.Context, status, code int, message string) {
	c.AbortWithStatusJSON(status, Response{Code: code, Msg: message})
}
