/*
@Time : 2021/3/8 15:35
@Author : nickqnxie
@File : dbx.go
*/

package baseclass

import (
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/tietang/dbx"
	"github.com/tietang/props/kvs"
	"github.com/nickqnxie/tuna/framework"
)

//dbx 数据库实例
var database *dbx.Database

func DbxDatabase() *dbx.Database {
	return database
}

//dbx数据库starter，并且设置为全局
type DbxDatabaseStarter struct {
	framework.BaseStarter
}

func (s *DbxDatabaseStarter) Setup(ctx framework.StarterContext) {
	conf := ctx.Props()
	//数据库配置
	settings := dbx.Settings{}
	err := kvs.Unmarshal(conf, &settings, "mysql")

	if err != nil {
		panic(err)
	}
	log.Info("mysql.conn url:", settings.ShortDataSourceName())
	db, err := dbx.Open(settings)

	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	database = db
}
