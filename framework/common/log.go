/*
@Time : 2019/10/12 15:58
@Author : nickqnxie
@File : log.go
*/

package common

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	"github.com/tietang/go-utils"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var formatter *prefixed.TextFormatter
var lfh *utils.LineNumLogrusHook

func setLineNumLogrusHook() {
	lfh = utils.NewLineNumLogrusHook()
	lfh.EnableFileNameLog = true
	lfh.EnableFuncNameLog = true
	log.AddHook(lfh)
}

func LogInit(logConfig LogConfig) (io.Writer, error) {

	// 定义日志格式
	formatter = &prefixed.TextFormatter{}
	//设置高亮显示的色彩样式
	formatter.ForceColors = false
	formatter.DisableColors = true
	formatter.ForceFormatting = true

	//开启完整时间戳输出和时间戳格式
	formatter.FullTimestamp = true
	//设置时间格式
	formatter.TimestampFormat = "2006-01-02.15:04:05.000000"
	//设置日志formatter
	log.SetFormatter(formatter)

	log.SetOutput(colorable.NewColorableStdout())

	//开启调用函数、文件、代码行信息的输出
	log.SetReportCaller(true)

	logLevel, err := log.ParseLevel(logConfig.Level)
	if err != nil {
		errorInfo := fmt.Sprintf("log.level:%s Parser failed, error:%s", logConfig.Level, err)
		return nil, errors.New(errorInfo)
	}
	log.SetLevel(logLevel)

	//设置函数、文件、代码行信息的输出的hook
	setLineNumLogrusHook()

	if logConfig.FileName == "" {
		log.SetOutput(os.Stdout)
		return os.Stdout, nil
	}

	logFilePath := fmt.Sprintf("%s/%s", logConfig.Path, logConfig.FileName)
	maxAge, _ := time.ParseDuration(fmt.Sprintf("%ds", logConfig.MaxAge))
	rotationTime, _ := time.ParseDuration(fmt.Sprintf("%ds", logConfig.RotationTime))
	logf, err := rotatelogs.New(
		fmt.Sprintf("%s.%%Y%%m%%d%%H%%M", logFilePath),
		rotatelogs.WithLinkName(logFilePath),
		rotatelogs.WithMaxAge(maxAge),             // 日志最大保存时间
		rotatelogs.WithRotationTime(rotationTime), // 单个日志文件多长时间分割一次
	)
	if err != nil {
		errorInfo := fmt.Sprintf("failed to create rotatelogs: %s", err)
		return nil, errors.New(errorInfo)
	}

	log.SetOutput(logf)

	return logf, nil
}
