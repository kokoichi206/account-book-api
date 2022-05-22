package util

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapのLoggerの初期化を行う。
func InitLogger() *zap.Logger {

	sy := getLogWriter()
	encoder := getEncoder()

	// TODO: 常にzapcore.InfoLevelでいいか？
	core := zapcore.NewCore(encoder, sy, zapcore.InfoLevel)
	lg := zap.New(core, zap.AddCaller())

	return lg
}

// jsonエンコーダーの設定。
func getEncoder() zapcore.Encoder {

	encoderConfig := zap.NewProductionEncoderConfig()

	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	return zapcore.NewJSONEncoder(encoderConfig)
}

// ログの出力先を取得する。
func getLogWriter() zapcore.WriteSyncer {
	// 標準出力に出力するようにする。
	return zapcore.AddSync(os.Stdout)
}
