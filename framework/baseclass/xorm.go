/*
@Time : 2019/9/2 20:06
@Author : nickqnxie
@File : xorm.go
*/

package baseclass

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"reflect"
	"github.com/nickqnxie/tuna/framework"
)

var engine *xorm.Engine

func Xormclinet() *xorm.Engine {
	return engine
}

type XormDBStarter struct {
	framework.BaseStarter
}

type DBops struct {
	engine *xorm.Engine
}

func NewDBops(engine *xorm.Engine) *DBops {
	return &DBops{engine: engine}
}

func (s *XormDBStarter) Setup(ctx framework.StarterContext) {
	var err error
	conf := ctx.Props()
	//数据库配置
	user, _ := conf.Get("mysql.user")
	host, _ := conf.Get("mysql.host")
	password, _ := conf.Get("mysql.password")
	prot, _ := conf.Get("mysql.port")
	database, _ := conf.Get("mysql.database")
	driverName, _ := conf.Get("mysql.driverName")
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8", user, password, host, prot, database)

	engine, err = xorm.NewEngine(driverName, dsn)

	if err != nil {
		logrus.Error("数据库连接失败:", err.Error())
		panic(err)
	}
	dsn = fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8", user, "*********", host, prot, database)
	logrus.Info("数据库连接成功:", dsn)

}

func (db *DBops) Insert(objTab interface{}) (err error) {
	_, err = db.engine.Insert(objTab)
	return
}

//更新操作
func (db *DBops) Update(objTab interface{}) (err error) {
	var id int64
	rval := reflect.ValueOf(objTab)
	id = rval.FieldByName("Id").Int()
	_, err = db.engine.Id(id).Update(objTab)
	return
}

//删除操作
func (db *DBops) Delete(objTab interface{}) (err error) {
	var id int64
	rval := reflect.ValueOf(objTab)
	id = rval.FieldByName("Id").Int()
	_, err = db.engine.Id(id).Delete(objTab)
	return
}

//获取数据,根据sql进行查询
func (db *DBops) GetExeSQL(sql string) (data []byte, err error) {
	var results []map[string]string
	if results, err = db.engine.QueryString(sql); err != nil {
		return
	}
	data, err = json.Marshal(results)
	return
}

/*

	db := DBops{
		engine: engine,
	}
	添加
	u := users{Name: "张山疯", Age: 105}
	db.Insert(u)

	根据id进行更新。需要获取id
	u := users{Id: 5, Name: "张山疯", Age: 205}
	db.Update(u)

	根据id进行删除
	u := users{Id: 5}
	db.Delete(u)

	获取一条数据
	u := users{Id: 5}
	data, _ := db.GetOne(u)
	fmt.Println(string(data))

	自定义sql获取多条数据库。
	var u []users
	data, _ := db.GetExeSQL(u, "select * from users where id = 5;")
	fmt.Println(string(data))

	sql := "select * from userss where id =2;"
	data, err := db.GetExeSQL(sql)

*/
