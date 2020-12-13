package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"lychee/config"
	"lychee/logger"
	"lychee/response"
	"os/exec"
	"strings"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
)

//RtspTransReq rtsp转websocket代理请求 struct
type RtspTransReq struct {
	SourceUrl   string `json:"sourceUrl"`
	ParamBefore string
	ParamBehind string
}

// processMap 保持FFMPEG进程，未在指定时间刷新的流将会被关闭
var processMap sync.Map

func (req *RtspTransReq) Service() *response.Response {
	simpleString := strings.Replace(req.SourceUrl, "//", "/", 1)
	// splitList := strings.Split(simpleString, "/")
	channel := uuid.NewV3(uuid.NamespaceURL, simpleString).String()
	if ch, ok := processMap.Load(channel); ok {
		*ch.(*chan int) <- 1
	} else {
		reflush := make(chan int)
		if cmd, stdin, err := toTrans(req, channel); err != nil {
			return response.Err(400, err.Error(), err)
		} else {
			go keepAlive(cmd, stdin, &reflush, channel)
		}
	}
	return response.Success(channel)
}

// keepAlive 进程保活，如果超过5分钟没有请求，则释放进程
func keepAlive(cmd *exec.Cmd, stdin io.WriteCloser, ch *chan int, channel string) {
	processMap.Store(channel, ch)
	defer func() {
		processMap.Delete(channel)
		_ = stdin.Close()
		logger.ZapLogger.Info(fmt.Sprintf("Stop translate rtsp id %v", channel))
	}()
	for {
		select {
		case <-*ch:
			logger.ZapLogger.Info(fmt.Sprintf("Reflush channel %s", channel))
		case <-time.After(5 * 60 * time.Second):
			_, _ = stdin.Write([]byte("q"))
			err := cmd.Wait()
			if err != nil {
				logger.ZapLogger.Error(fmt.Sprintf("Run ffmpeg err %v", err.Error()))
			}
			return
		}
	}
}

// toTrans 通过FFMPEG实现rtsp转websocket代理
func toTrans(req *RtspTransReq, channel string) (*exec.Cmd, io.WriteCloser, error) {
	port := config.AllConfig.Server.Port
	var build strings.Builder
	build.WriteString(req.ParamBefore)
	build.WriteString(",")
	build.WriteString(req.SourceUrl)
	build.WriteString(",")
	build.WriteString(req.ParamBehind)
	build.WriteString(",")
	build.WriteString(fmt.Sprintf("http://127.0.0.1:%s/stream/live/%s", port, channel))
	tempParam := build.String()
	params := strings.Split(tempParam, ",")
	logger.ZapLogger.Info("FFmpeg cmd: ffmpeg " + strings.Join(params, " "))
	outInfo := bytes.Buffer{}
	cmd := exec.Command("ffmpeg", params...)
	logger.ZapLogger.Info(fmt.Sprintf("command args is %s", cmd.Args))
	cmd.Stdout = &outInfo
	cmd.Stderr = nil
	stdin, err := cmd.StdinPipe()
	if err != nil {
		logger.ZapLogger.Error(fmt.Sprintf("Get ffmpeg stdin err:%v", err.Error()))
		return nil, nil, errors.New("拉流进程启动失败")
	}

	err = cmd.Start()
	if err != nil {
		logger.ZapLogger.Info(fmt.Sprintf("Start ffmpeg err: %v", err.Error()))
		return nil, nil, errors.New("打开摄像头视频流失败")
	}
	logger.ZapLogger.Info(fmt.Sprintf("command output is %s", outInfo.String()))
	logger.ZapLogger.Info(fmt.Sprintf("Translate rtsp %v to %v", req.SourceUrl, channel))
	return cmd, stdin, nil
}
