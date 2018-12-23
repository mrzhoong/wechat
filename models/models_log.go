package models

import "github.com/astaxie/beego/logs"

func SetLevel(l int) {
	logs.SetLevel(l)
}

func init() {
	logs.SetLogger(logs.AdapterFile, `{"filename":"log/wechat.log",
		"level":6,"maxlines":3000000,"maxsize":0,"daily":true,"maxdays":5}`)

	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(2)
}
