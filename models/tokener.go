package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// TokenFetcher 包含向 API 服务器获取令牌信息的操作
type TokenFetcher interface {
	FetchToken() (token string, expiresIn int64, err error)
}

// Tokener 用于管理应用套件或企业号的令牌信息
type Tokener struct {
	token        string
	expiresIn    int64
	tokenFetcher TokenFetcher
}

// NewTokener 方法用于创建 Tokener 实例
func NewTokener(tokenFetcher TokenFetcher) *Tokener {
	return &Tokener{tokenFetcher: tokenFetcher}
}

// Token 方法用于获取应用套件令牌
func (t *Tokener) Token() (token string, err error) {
	if !t.isValidToken() {
		if err = t.RefreshToken(); err != nil {
			return "", err
		}
	}

	return t.token, nil
}

// RefreshToken 方法用于刷新令牌信息
func (t *Tokener) RefreshToken() error {
	token, expiresIn, err := t.GetToken()
	if err != nil {
		return err
	}

	expiresIn = time.Now().Add(time.Second * time.Duration(expiresIn)).Unix()
	t.token = token
	t.expiresIn = expiresIn

	return nil
}

func (t *Tokener) isValidToken() bool {
	now := time.Now().Unix()

	if now >= t.expiresIn || t.token == "" {
		return false
	}

	return true
}

const (
	getTokenURL = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"
)

type RespToken struct {
	Errcode     int64  `json:"errcode"`
	Errmsg      string `json:"errmsg"`
	AccessToken string `json:"access_Token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (t *Tokener) GetToken() (token string, expiresIn int64, err error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	url := getTokenURL + "?" + fmt.Sprintf("corpid=%s&corpsecret=%s", AppID, Secret)
	data, err := json.Marshal("")
	if err != nil {
		return "", 0, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}

	if resp.StatusCode == http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("[resp]", string(body))
		var tk RespToken
		err := json.Unmarshal(body, &tk)
		if err != nil {
			fmt.Println("get token errors :", err.Error())
			return "", 0, err
		}
		t.token = tk.AccessToken
		t.expiresIn = tk.ExpiresIn - 200
		return t.token, t.expiresIn, nil
	} else {
		return "", 0, err
	}
}
