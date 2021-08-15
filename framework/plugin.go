/*
@Time : 2019/8/11 23:17
@Author : nickqnxie
@File : plugin.go
*/

package framework

var plunginapiInitializerRegister *InitializeRegister = new(InitializeRegister)

//注册WEB API初始化对象

func PluginRegisterApi(ai Initializer) {
	plunginapiInitializerRegister.Register(ai)
}

//获取注册的web api初始化对象
func GetPluginApiInitializers() []Initializer {
	return plunginapiInitializerRegister.Initializers
}

type PluginApiStarter struct {
	BaseStarter
}

func (w *PluginApiStarter) Setup(ctx StarterContext) {
	for _, v := range GetPluginApiInitializers() {
		v.Init()
	}
}
