// @Title  logger
// @Description  日志工具，使用的ubar的zap，以及使用lumberjack来支持滚动写入文件
// @Author  tongwei  2020.12.12
package logger

import (
	"lychee/config"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var ZapLogger *zap.Logger

func InitLogger() {
	name := config.AllConfig.Log.Name
	fileName := config.AllConfig.Log.FileName
	maxSize := config.AllConfig.Log.MaxSize
	maxAge := config.AllConfig.Log.MaxAge
	maxBackups := config.AllConfig.Log.MaxBackups
	stdout := config.AllConfig.Log.Stdout
	hook := lumberjack.Logger{
		Filename:   fileName,   // 日志文件路径
		MaxSize:    maxSize,    // 每个日志文件保存的大小 单位:M
		MaxAge:     maxAge,     // 文件最多保存多少天
		MaxBackups: maxBackups, // 日志文件最多保存多少个备份
		Compress:   false,      // 是否压缩
	}
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "file",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 短路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zap.DebugLevel)
	var writes = []zapcore.WriteSyncer{zapcore.AddSync(&hook)}
	// 如果是开发环境，同时在控制台上也输出
	if stdout {
		writes = append(writes, zapcore.AddSync(os.Stdout))
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(writes...),
		atomicLevel,
	)

	// 开启开发模式，堆栈跟踪
	caller := zap.AddCaller()
	// 开启文件及行号
	development := zap.Development()

	// 设置初始化字段
	field := zap.Fields(zap.String("appName", name))

	// 构造日志
	ZapLogger = zap.New(core, caller, development, field)
	ZapLogger.Info("log 初始化成功")
}
