package response

import (
	"lychee/logger"

	"github.com/gin-gonic/gin"
)

// Response 基础序列化器
type Response struct {
	Code  int         `json:"code"`
	Data  interface{} `json:"data,omitempty"`
	Msg   string      `json:"msg"`
	Error string      `json:"error,omitempty"`
}

// PlayDate RTSP play 信息序列化器
type PlayDate struct {
	Channel string `json:"channel"`
}

// Err 通用错误处理
func Err(errCode int, msg string, err error) *Response {
	res := Response{
		Code: errCode,
		Msg:  msg,
	}

	if err != nil {
		logger.ZapLogger.Error(err.Error())
		// 生产环境隐藏底层报错
		if gin.Mode() != gin.ReleaseMode {
			res.Error = err.Error()
		}
	}
	return &res
}

func Success(channel string) *Response {
	return &Response{
		Data: &PlayDate{
			Channel: channel,
		},
		Msg: "success",
	}
}
