package controllers

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"wechat/models"

	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
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

//
//type TemplateRequest struct {
//	XMLName    xml.Name `xml:"xml"`
//	ToUserName string   `xml:"ToUserName"`
//	AgentID    string   `xml:"toAgentID"`
//	Encrypt    string   `xml:"msg_encrypt"`
//}

type TempLateResponse struct {
	XMLName      xml.Name `xml:"xml"`
	Encrypt      string   `xml:"msg_encrypt"`
	MsgSignature string   `xml:"msg_signature"`
	TimeStamp    string   `xml:"timestamp"`
	Nonce        string   `xml:"nonce"`
}

type Request struct {
	ToUserName string
	AgentID    string
	Encrypt    string
}

func (c *MainController) Post() {
	_, nonce, signature := c.GetString("timestamp"),
		c.GetString("nonce"),
		c.GetString("msg_signature")

	var req Request

	if err := xml.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		fmt.Println("xml errors:", err.Error())
	}
	fmt.Println("Receive message:", string(c.Ctx.Input.RequestBody))

	recvData, _, _ := models.MsgCrypter.Decrypt(req.Encrypt)
	// 解析消息类型
	probeData := &struct {
		MsgType models.MessageType
		Event   string
	}{}

	if err := xml.Unmarshal(recvData, probeData); err != nil {
		var resp TempLateResponse
		resp = TempLateResponse{MsgSignature: signature,
			TimeStamp: fmt.Sprintf("%d", time.Now().Unix()),
			Nonce:     nonce}
		respbody, _ := xml.Marshal(resp)
		respxml := []byte(xml.Header)
		respxml = append(respxml, respbody...)
		fmt.Println("Send message:", string(respxml))
		c.Ctx.WriteString(string(respxml))
		fmt.Println("recv errors")
		return
	}

	switch probeData.MsgType {
	case models.TextMsg:
		var t models.RecvTextMessage

		if err := xml.Unmarshal(recvData, &t); err != nil {
			return
		}

		if s := strings.Split(t.Content, "#"); 3 == len(s) {
			models.AddTimedTask("bhb", s[1], s[2], t.FromUserName)
		} else {
			models.GetPriceTasks(t.Content, t.FromUserName)
		}
	default:
		fmt.Println("unknown type")
		c.Ctx.WriteString("unknown type")
	}
	//// 回复消息
	//var resp TempLateResponse
	//resp = TempLateResponse{MsgSignature: signature,
	//	TimeStamp: fmt.Sprintf("%d", time.Now().Unix()),
	//	Nonce:     nonce}
	//respbody, _ := xml.Marshal(resp)
	//respxml := []byte(xml.Header)
	//respxml = append(respxml, respbody...)
	//fmt.Println("Send message:", string(respxml))
	//c.Ctx.WriteString(string(respxml))
	c.Ctx.WriteString("")
}
