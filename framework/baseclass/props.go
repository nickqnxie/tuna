package baseclass

import (
	log "github.com/sirupsen/logrus"
	"github.com/tietang/props/kvs"
	"github.com/nickqnxie/tuna/framework"
)

var props kvs.ConfigSource

func Props() kvs.ConfigSource {
	return props
}

type PropsStarter struct {
	framework.BaseStarter
}

func (p *PropsStarter) Init(ctx framework.StarterContext) {
	props = ctx.Props()
	log.Info("初始化配置.")
}
