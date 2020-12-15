package controller

import (
	"bufio"
	"lychee/service"

	"github.com/gin-gonic/gin"
)

func AcceptRespAndPushToFfmpeg(c *gin.Context) {
	req := &service.RtspTransReq{}
	c.BindJSON(req)
	ret := req.Service()
	c.JSON(200, ret)
}

// Mpeg1Video 接收 mpeg1vido 数据流
func Mpeg1Video(c *gin.Context) {
	bodyReader := bufio.NewReader(c.Request.Body)

	for {
		data, err := bodyReader.ReadBytes('\n')
		if err != nil {
			break
		}
		service.WsManager.Groupbroadcast(c.Param("channel"), data)
	}
}

// Wsplay 通过 websocket 播放 mpegts 数据
func Wsplay(c *gin.Context) {
	service.WsManager.RegisterClient(c)
}
