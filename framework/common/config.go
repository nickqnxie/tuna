/*
@Time : 2019/10/12 18:15
@Author : nickqnxie
@File : config.go
*/

package common

type LogConfig struct {
	Path         string `yaml:"path"`         // 日志路径,默认值为当前目录
	FileName     string `yaml:"fileName"`     // 日志文件名称。 默认为app.log
	Level        string `yaml:"level"`        // 日志级别， warn/debug/info/error，默认为warn级别
	MaxAge       int64  `yaml:"maxAge"`       // 日志保存时间，默认存放30天，以秒为单位
	RotationTime int64  `yaml:"rotationTime"` // 日志分割时间，默认一天分割一次，以秒为单位
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Passwd   string `yaml:"passwd"`
	Database string `yaml:"database"`
}

type ConsulConfig struct {
	Address string `yaml:"address"`
	Token   string `yaml:"token"`
}
