package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

// Init 初始化 zap，env 建议使用 dev / prod
// dev: 控制台友好输出
// prod: JSON + stdout + 文件滚动
func Init(env string) {
	isProd := env == "prod"

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	encoderCfg.EncodeDuration = zapcore.MillisDurationEncoder

	var encoder zapcore.Encoder
	if isProd {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	// 输出
	writers := []zapcore.WriteSyncer{zapcore.AddSync(os.Stdout)}
	if isProd {
		// 生产落盘，示例：/var/log/youlai-gin/app.log
		fileSync := zapcore.AddSync(&lumberjack.Logger{
			Filename:   "/var/log/youlai-gin/app.log",
			MaxSize:    100, // MB
			MaxBackups: 10,  // 个数
			MaxAge:     30,  // 天
			Compress:   true,
		})
		writers = append(writers, fileSync)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(writers...),
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	Log = logger
	zap.ReplaceGlobals(logger)
}

// Sync 刷盘
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

