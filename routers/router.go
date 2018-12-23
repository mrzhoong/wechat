package routers

import (
	"github.com/astaxie/beego"
	"wechat/controllers"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/config", &controllers.MsgController{})
}
