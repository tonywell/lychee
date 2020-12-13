# lychee
荔枝RTSP转websocket代理，基于FFMPEG将RTSP视频流代理websocket，实现H5页面超低延时播放，前端播放器采用JSMpeg。

## 项目说明

### 工程结构
.
├── README.md
├── config --配置文件读取
│   ├── conf.go
│   └── conf.yaml
├── controller. --接口
│   └── rtsp.go
├── go.mod
├── go.sum
├── logger --日志模块
│   └── logger.go
├── lychee --可执行程序
├── response --web响应
│   └── response.go
├── router --路由
│   └── router.go
├── server.go
└── service --service层
    ├── WsService.go
    └── rtspTransService.go

### 技术栈
* gin 是一个用 Go (Golang) 编写的 web 框架，用来写接口和http服务
* viper 适用于Go应用程序的完整配置解决方案，读取yaml配置文件
* zap Uber开源的高性能日志库
* Lumberjack 用于将日志写入滚动文件。zap 不支持文件归档，如果要支持文件按大小或者时间归档，需要使用lumberjack，lumberjack也是zap官方推荐的。
* gorilla/websocket websocket工具包

### 注意事项
* 需要依赖FFMPEG，所以需要允许在安装有FFMPEG的系统中
* 这里编译好的lychee是在apline linux环境编译的，如果要支持其他系统需重新编译
* 如果在非linux编译，且需要运行在linux环境中，需要交叉编译环境

## 运行
在linux环境，只需要将lychee设置为可以运行，就可以直接运行
```
$ ./lychee
```
启动后，请求转码接口/stream/push
```
POST /stream/push
{
   "sourceUrl": "rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mov",
   "paramBefore": "-y -rtsp_transport tcp -re -i",
   "paramBehind": "-q 0 -f mpegts -c:v mpeg1video -an -s 960x540"
}
```
* sourceUrl： 为RTSP源地址，这里我是网上找到的一个用于测试的地址，比较卡，转的过程中会终端，测试的时候需要改为稳定且速度正常的RTSP源
* paramBefore： ffmpeg命令中，在源地址前面的参数
* paramBehind： ffmpeg命令中，在源地址后的参数
最终执行的ffmpeg命令为：ffmpeg -y -rtsp_transport tcp -re -i rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mov -q 0 -f mpegts -c:v mpeg1video -an -s 960x540 http://127.0.0.1:8000/stream/live/02385e5f-6307-3f77-8674-0ba26d233255

后台返回的响应为：
```
 {
  "code": 0,
  "data": {
    "channel": "02385e5f-6307-3f77-8674-0ba26d233255"
  },
  "msg": "success"
}
```
拼接上地址ws://ip:port/stream/live/02385e5f-6307-3f77-8674-0ba26d233255即可实现播放

