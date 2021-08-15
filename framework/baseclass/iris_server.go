package baseclass

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/kataras/iris"
	"github.com/kataras/iris/core/host"
	"github.com/kataras/iris/middleware/logger"
	irisrecover "github.com/kataras/iris/middleware/recover"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"github.com/nickqnxie/tuna/framework"
	"time"
)

var irisApplication *iris.Application

func Iris() *iris.Application {
	Check(irisApplication)
	return irisApplication
}

type IrisServerStarter struct {
	framework.BaseStarter
}

func (i *IrisServerStarter) Init(ctx framework.StarterContext) {

	//创建iris application实例
	irisApplication = initIris()
	//日志组件配置和扩展
	logger := irisApplication.Logger()
	logger.Install(logrus.StandardLogger())

}

func Cors(ctx iris.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")
	if ctx.Request().Method == "OPTIONS" {
		ctx.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization")
		ctx.StatusCode(204)
		return
	}
	ctx.Next()
}

func (i *IrisServerStarter) Start(ctx framework.StarterContext) {
	//把路由信息打印到控制台

	Iris().Logger().SetLevel(ctx.Props().GetDefault("log.level", "info"))

	routes := Iris().GetRoutes()
	for _, r := range routes {
		logrus.Info(r.Trace())
	}

	webroot := ctx.Props().GetDefault("webroot", "./webroot")
	Iris().StaticWeb("/", webroot)

	// 如果没有命中/api，则直接转默认路由
	Iris().WrapRouter(func(w http.ResponseWriter, r *http.Request, router http.HandlerFunc) {
		path := r.URL.Path
		//请注意，如果path的后缀为“index.html”，则会自动重定向到“/”，
		//所以我们的第一个处理程序将被执行。
		if strings.Contains(path, "api/") || strings.Contains(path, "v1/") || strings.Contains(path, ".") || path == "/" {
			//如果它不是资源，那么就像正常情况一样继续使用路由器. <-- IMPORTANT
			logrus.Info(path)
			router(w, r)
			return
		}

		ctxIris := Iris().ContextPool.Acquire(w, r)
		ctxIris.Redirect("/")
		Iris().ContextPool.Release(ctxIris)
	})

	port := ctx.Props().GetDefault("app.server.port", "18080")

	enableSSL := ctx.Props().GetBoolDefault("ssl.enable", false)
	certFile := ctx.Props().GetDefault("ssl.crt", "server.crt")
	keyFile := ctx.Props().GetDefault("ssl.key", "server.key")
	if enableSSL {

		Iris().Run(iris.TLS(":"+port, certFile, keyFile, func(su *host.Supervisor) {

			caCertPath := "/private/data/dev_ops/golang/src/tencent.com/certi_ops/certi_agent_ops/ssl/ca.crt"
			caCrt, _ := ioutil.ReadFile(caCertPath)
			pool := x509.NewCertPool()
			pool.AppendCertsFromPEM(caCrt)
			su.Server.TLSConfig = &tls.Config{
				ClientAuth: tls.RequireAndVerifyClientCert,
				ClientCAs:  pool,
			}
		}))
		//app.Run(iris.Raw(&http.Server{Addr:":8080"}).ListenAndServe)
		//Iris().Run(iris.Raw(&http.Server{Addr: ""}))
	} else {
		Iris().Run(iris.Addr(":" + port))

		//app.Run(iris.Server(&http.Server{Addr:":8080"}))
	}

}
func (i *IrisServerStarter) StartBlocking() bool {
	return true
}

func DoAuth(ctx iris.Context) {
	var staffname string
	staffname = ctx.GetHeader("staffname")
	logrus.Debug("staffname=", staffname)
	if staffname == "" {
		logrus.Info("认证跳转...")
		ctx.Redirect("http://passport.oa.com/modules/passport/signin.ashx?url=http://essearch.qcloud.oa.com", 301)
		return
	}
	ctx.Next()
}

func isExists(certsPath string) bool {
	_, err := os.Stat(certsPath)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func initIris() *iris.Application {

	enableSSL := Props().GetBoolDefault("ssl.enable", false)

	if enableSSL {
		logrus.Info("服务器启用了https")
		crt, err := Props().Get("ssl.crt")
		if err != nil {
			logrus.Error("缺少公钥配置..")
		} else {
			if !isExists(crt) {
				logrus.Errorf("公钥文件:%s 不存在..", crt)
				os.Exit(1)
			}
		}
		key, err := Props().Get("ssl.key")

		if err != nil {
			logrus.Error("缺少私钥配置..")
		} else {
			if !isExists(key) {
				logrus.Errorf("私钥文件:%s 不存在..", key)
				os.Exit(1)
			}
		}

		logrus.Info("开启https服务")
		logrus.Info("公钥路径:" + crt)
		logrus.Info("私钥路径:" + key)

	}

	app := iris.New()
	app.Use(irisrecover.New())
	// 主要中间件的配置:recover,日志输出中间件的自定义
	cfg := logger.Config{
		Status:             true,
		IP:                 true,
		Method:             true,
		Path:               true,
		Query:              true,
		Columns:            true,
		MessageContextKeys: []string{"logger_message"},
		MessageHeaderKeys:  []string{"User-Agent"},
		LogFunc: func(now time.Time, latency time.Duration,
			status, ip, method, path string,
			message interface{},
			headerMessage interface{}) {
			app.Logger().Infof("| %s | %s | %s | %s | %s | %s | %+v | %+v",
				now.Format("2006-01-02.15:04:05.000000"),
				latency.String(), status, ip, method, path, headerMessage, message,
			)
		},
	}
	app.Use(logger.New(cfg))
	app.Use(iris.NoCache)
	app.Use(Cors)
	//if authboolDefault {
	//	app.Use(DoAuth)
	//}
	return app
}
