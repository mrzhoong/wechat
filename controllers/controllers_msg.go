package controllers

import (
	"common"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"wechat/models"
)

type MsgController struct {
	beego.Controller
}

func (c *MsgController) Get() {
	timestamp, nonce, signature, echostr := c.GetString("timestamp"),
		c.GetString("nonce"),
		c.GetString("msg_signature"),
		c.GetString("echostr")
	fmt.Println(fmt.Sprintf("[timestamp]%s", timestamp))
	fmt.Println(fmt.Sprintf("[nonce]%s", nonce))
	fmt.Println(fmt.Sprintf("[msg_signature]%s", signature))
	fmt.Println(fmt.Sprintf("[echostr]%s", echostr))

	ret := models.MsgCrypter.GetSignature(timestamp, nonce, echostr)
	fmt.Println(fmt.Sprintf("[signature]%s", ret))
	b, s, err := models.MsgCrypter.Decrypt(echostr)
	if err != nil {
		fmt.Println("[Decrypt]", err.Error())
	}
	fmt.Println(fmt.Sprintf("[msgDecrypt]%s[id]%s", string(b), s))
	c.Ctx.WriteString(string(b))
}

type RequestCommon struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (c *MsgController) Post() {
	var ob RequestCommon
	json.Unmarshal(c.Ctx.Input.RequestBody, &ob)
	switch ob.Type {
	case "set_level":
		l := ob.Data.(float64)
		models.SetLevel(int(l))
		c.Ctx.WriteString(common.RespData("set log level success", nil))
	default:
		c.Ctx.WriteString(common.RespError("00001", "undefined type"))
	}
}
