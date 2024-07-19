package klog

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	KLog struct {
		Filename  string `json:"file_name"`
		MaxSize   int32  `json:"max_size"`
		MaxBackup int32  `json:"max_backup"`
		MaxAge    int32  `json:"max_age"`
		Compress  bool   `json:"compress"`
	} `json:"klog"`
}

var config Config
var logger *zap.Logger

func LoadConfig() error {
	pwd, _ := os.Getwd()
	file,err := os.Open(path.Join(pwd, "config.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return err
	}

	return nil
}

func InitService() error {
	LoadConfig()
	lumberjackLogger := &lumberjack.Logger{
		Filename: config.KLog.Filename,
		MaxSize: int(config.KLog.MaxSize),
		MaxBackups: int(config.KLog.MaxBackup),
		MaxAge: int(config.KLog.MaxAge),
		Compress: config.KLog.Compress,
	}

	// file writer creation
	fileSyncer := zapcore.AddSync(lumberjackLogger)
	// console writer creation
	consoleSyncer := zapcore.AddSync(zapcore.Lock(os.Stdout))

	// Configure encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create logger core
	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			fileSyncer,
			zap.InfoLevel,
		),
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			consoleSyncer,
			zap.InfoLevel,
		),
	)

	logger = zap.New(core)
	defer logger.Sync()
	return nil
}