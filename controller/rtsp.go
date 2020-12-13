package controller

import (
	"lychee/service"

	"github.com/gin-gonic/gin"
)

func AcceptRespAndPushToFfmpeg(c *gin.Context) {
	req := &service.RtspTransReq{}
	c.BindJSON(req)
	ret := req.Service()
	c.JSON(200, ret)
}

// Wsplay 通过 websocket 播放 mpegts 数据
func Wsplay(c *gin.Context) {
	service.WsManager.RegisterClient(c)
}
