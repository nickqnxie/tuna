/*
@Time : 2019/10/16 20:36
@Author : nickqnxie
@File : grom.go
*/

package common

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

func NewdbClient(confg DBConfig) (db *gorm.DB, err error) {

	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8&parseTime=true&loc=Local",
		confg.User, confg.Passwd, confg.Host, confg.Port, confg.Database)
	if db, err = gorm.Open("mysql", dsn); err != nil {
		logrus.Debugf("数据库初始化失败,", err)
	}
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	return
}
