package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"wechat/crypter"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

var (
	Token          string
	EncodingAESKey string
	AppID          string
	AgentId        string
	Secret         string
	MsgCrypter     crypter.MessageCrypter
	AccessToken    *Tokener
)

func init() {
	Token = beego.AppConfig.String("Token")
	EncodingAESKey = beego.AppConfig.String("EncodingAESKey")
	AppID = beego.AppConfig.String("AppID")
	AgentId = beego.AppConfig.String("AgentId")
	Secret = beego.AppConfig.String("Secret")

	fmt.Println(fmt.Sprintf("[token]%s[EncodingAESKey]%s[AppID]%s[AgentId]%s[Secret]%s",
		Token,
		EncodingAESKey,
		AppID,
		AgentId,
		Secret))

	CreateObject()
	CreateToken()
}

func CreateObject() {
	MsgCrypter, _ = crypter.NewMessageCrypter(Token, EncodingAESKey, AppID)
}

func CreateToken() {
	var tk TokenFetcher
	AccessToken = NewTokener(tk)
	AccessToken.GetToken()
}

// 发送消息
// 发送消息
const (
	sendMessageURI = "https://qyapi.weixin.qq.com/cgi-bin/message/send"
)

func SendMessage(m interface{}) error {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	Tok, err := AccessToken.Token()
	if err != nil {
		Tok, _, _ = AccessToken.GetToken()
	}
	url := sendMessageURI + "?" + "access_token=" + Tok
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("[resp]", string(body))
		logs.Info(string(body))
		return nil
	} else {
		return err
	}
}
