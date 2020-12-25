package service

import (
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
	Clarity     string
	ParamBefore string
	ParamBehind string
}

// processMap 保持FFMPEG进程，未在指定时间刷新的流将会被关闭
var processMap sync.Map

func (req *RtspTransReq) Service() *response.Response {
	simpleString := strings.Replace(req.SourceUrl, "//", "/", 1)
	simpleString = simpleString + req.ParamBefore
	channel := uuid.NewV3(uuid.NamespaceURL, simpleString).String()
	port := config.AllConfig.Server.Port
	var build strings.Builder
	build.WriteString(req.ParamBefore)
	build.WriteString(" ")
	build.WriteString(req.SourceUrl)
	build.WriteString(" ")
	build.WriteString(req.ParamBehind)
	build.WriteString(" ")
	build.WriteString(fmt.Sprintf("http://127.0.0.1:%s/stream/upload/%s", port, channel))
	commandStr := build.String()
	args := strings.Split(commandStr, " ")
	if ch, ok := processMap.Load(channel); ok {
		*ch.(*chan int) <- 1
	} else {
		reflush := make(chan int)
		if cmd, stdin, err := toTrans(args); err != nil {
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
func toTrans(params []string) (*exec.Cmd, io.WriteCloser, error) {
	cmd := exec.Command("ffmpeg", params...)
	logger.ZapLogger.Info("FFmpeg cmd args: " + strings.Join(params, " "))
	// cmd := exec.Command("ffmpeg", "-y", "-rtsp_transport", "tcp", "-re", "-i", "rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mov", "-q", "5", "-f", "mpegts", "-c:v", "mpeg1video", "-an", "-s", "960x540", "http://127.0.0.1:8000/stream/upload/02385e5f-6307-3f77-8674-0ba26d233255")
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
	return cmd, stdin, nil
}
