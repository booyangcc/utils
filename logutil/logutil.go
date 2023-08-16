package logutil

import (
	"os"
	"path/filepath"

	"github.com/booyangcc/utils/fileutil"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// DefaultLogPath 默认输出日志文件路径
const DefaultLogPath = "/var/log/test"

const (
	// LogStdout log stdout
	LogStdout = iota
	// LogFile logfile
	LogFile
	// LogStdoutAndFile log stdout and file.
	LogStdoutAndFile
)

// Config config
type Config struct {
	LogLevel          string // 日志打印级别 debug  info  warning  error
	LogFormat         string // 输出日志格式	logfmt, json
	LogPath           string // 输出日志文件路径
	LogFileName       string // 输出日志文件名称
	LogFileMaxSize    int    // 【日志分割】单个日志文件最多存储量 单位(mb)
	LogFileMaxBackups int    // 【日志分割】日志备份文件最多数量
	LogMaxAge         int    // 日志保留时间，单位: 天 (day)
	LogCompress       bool   // 是否压缩日志
	LogStdout         bool   // 是否输出到控制台
	LogType           int
	Caller            bool //是否输出调用链路
}

var defaultConfig = &Config{
	LogLevel:          "info",
	LogFormat:         "json",
	LogPath:           "./",
	LogFileName:       "app.log",
	LogFileMaxSize:    50,
	LogFileMaxBackups: 50,
	LogMaxAge:         7,
	LogCompress:       false,
	LogStdout:         true,
	Caller:            true,
}

// Option option.
type Option func(lc *Config)

// WithLogLevel with log level.
func WithLogLevel(level string) Option {
	return func(lc *Config) {
		lc.LogLevel = level
	}
}

// WithLogFormat with log format.
func WithLogFormat(logFormat string) Option {
	return func(lc *Config) {
		lc.LogFormat = logFormat
	}
}

// WithLogPath with log path.
func WithLogPath(logPath string) Option {
	return func(lc *Config) {
		lc.LogPath = logPath
	}
}

// WithLogFileName with log file name.
func WithLogFileName(fileName string) Option {
	return func(lc *Config) {
		lc.LogFileName = fileName
	}
}

// WithLogFileMaxSize withLogFileMaxSize
func WithLogFileMaxSize(size int) Option {
	return func(lc *Config) {
		lc.LogFileMaxSize = size
	}
}

// WithLogFileMaxBackups withLogFileMaxBackups
func WithLogFileMaxBackups(fileNum int) Option {
	return func(lc *Config) {
		lc.LogFileMaxBackups = fileNum
	}
}

// WithLogMaxAge withLogMaxAge
func WithLogMaxAge(logMaxAge int) Option {
	return func(lc *Config) {
		lc.LogMaxAge = logMaxAge
	}
}

// WithLogCompress withLogCompress
func WithLogCompress(logCompress bool) Option {
	return func(lc *Config) {
		lc.LogCompress = logCompress
	}
}

// WithStdout with stdout.
func WithStdout(isStdout bool) Option {
	return func(lc *Config) {
		lc.LogStdout = isStdout
	}
}

// WithLogType with log type.
func WithLogType(logType int) Option {
	return func(lc *Config) {
		lc.LogType = logType
	}
}

// WithCaller is print caller
func WithCaller(enabled bool) Option {
	return func(lc *Config) {
		lc.Caller = enabled
	}
}

// New new
func New(cfg *Config, opts ...Option) (*zap.Logger, error) {
	if cfg == nil {
		cfg = defaultConfig
	}

	for _, opt := range opts {
		opt(cfg)
	}
	log, err := initLogger(cfg)
	if err != nil {
		return nil, err
	}
	return log, nil
}

// GetLoggerFullPath get log full path.
func GetLoggerFullPath() string {
	return filepath.Join(defaultConfig.LogPath, defaultConfig.LogFileName)
}

func initLogger(cfg *Config) (*zap.Logger, error) {
	logLevel := map[string]zapcore.Level{
		"debug": zapcore.DebugLevel,
		"info":  zapcore.InfoLevel,
		"warn":  zapcore.WarnLevel,
		"error": zapcore.ErrorLevel,
	}
	writeSyncer, err := getLogWriter(cfg)
	if err != nil {
		return nil, err
	}
	encoder := getEncoder(cfg)
	level, ok := logLevel[cfg.LogLevel]
	if !ok {
		level = logLevel["info"]
	}

	core := zapcore.NewCore(encoder, writeSyncer, level)
	var l *zap.Logger
	if !cfg.Caller {
		l = zap.New(core)
	} else {
		l = zap.New(core, zap.AddCaller())
	}

	return l, nil
}

// getEncoder 编码器(如何写入日志)
func getEncoder(cfg *Config) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder   // log 时间格式 例如: 2021-09-11t20:05:54.852+0800
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 输出level序列化为全大写字符串，如 INFO DEBUG ERROR
	if cfg.LogFormat == "json" {
		return zapcore.NewJSONEncoder(encoderConfig) // 以json格式写入
	}
	return zapcore.NewConsoleEncoder(encoderConfig) // 以logfmt格式写入
}

func getLogWriter(cfg *Config) (zapcore.WriteSyncer, error) {
	err := fileutil.CreatePath(cfg.LogPath)
	if err != nil {
		return nil, err
	}

	lumberJackLogger := &lumberjack.Logger{
		Filename:   filepath.Join(cfg.LogPath, cfg.LogFileName), // 日志文件路径
		MaxSize:    cfg.LogFileMaxSize,                          // 单个日志文件最大多少 mb
		MaxBackups: cfg.LogFileMaxBackups,                       // 日志备份数量
		MaxAge:     cfg.LogMaxAge,                               // 日志最长保留时间
		Compress:   cfg.LogCompress,                             // 是否压缩日志
	}

	if cfg.LogType == LogStdout {
		return zapcore.AddSync(os.Stdout), nil
	} else if cfg.LogType == LogFile {
		return zapcore.AddSync(lumberJackLogger), nil
	} else if cfg.LogType == LogStdoutAndFile {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberJackLogger)), nil
	}
	return nil, err
}
