package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego/logs"
)

var (
	TimedTask map[string]CurrencyInfo
	TimedLock sync.Mutex
)

func init() {
	TimedTask = make(map[string]CurrencyInfo, 0)
	time.Sleep(2 * time.Second)
	fmt.Println("启动提醒线程")
	logs.Info("价格到达定时提醒已启用")
	go TimeTasks()
}

type Market struct {
	Code    string      `json:"code"`
	Success bool        `json:"success"`
	Msg     interface{} `json:"msg"`
	Data    interface{} `json:"data"`
}

type MarketInfo struct {
	Currency CurrencyInfo
}

type CurrencyInfo struct {
	Uid        string  `json:"uid"`
	Currency   string  `json:"currency"`
	Price      float64 `json:"price"`
	Direct     string  `json:"direct"` // 交易方向：小于、大于
	Volume     float64 `json:"volume"`
	High       float64 `json:"high"`
	Last       float64 `json:"last"`
	Low        float64 `json:"low"`
	Buy        float64 `json:"buy"`
	Sell       float64 `json:"sell"`
	Id         float64 `json:"id"`
	ChangeRate float32 `json:"changeRate"`
	TurnoOver  float64 `json:"turnover"`
}

// 查价
func GetPriceTasks(c, uid string) {
	go getPriceTask(c, uid)
}
func getPriceTask(c, uid string) {
	cy, err := getPrice(c)
	if err != nil {
		return
	}

	var d TextMessage
	d.AgentID = 1000002
	d.MsgType = TextMsg
	d.Safe = 0
	//d.Text = TextContent{Content: "你的快递已到，请携带工卡前往邮件中心领取。\n出发前可查看<a href=\"http://work.weixin.qq.com\">邮件中心视频实况</a>，聪明避开排队。"}
	d.Text = TextContent{Content: fmt.Sprintf("交易对:%s/usdt\n"+
		"最高:%.4f\n最低:%.4f\n最后成交价:%.4f\n买入:%.4f\n卖出:%.4f",
		strings.ToLower(c),
		cy.High,
		cy.Low,
		cy.Last,
		cy.Buy,
		cy.Sell,
	)}
	d.ToParty = ""
	d.ToTag = ""
	d.ToUser = uid
	// 发送消息接口
	logs.Info(fmt.Sprintf("%s %s", uid, d.Text))
	SendMessage(d)
}

func SendTimedTaskMessage(uid, c string, price float64) error {
	var d TextMessage
	d.AgentID = 1000002
	d.MsgType = TextMsg
	d.Safe = 0

	d.Text = TextContent{Content: fmt.Sprintf("交易对：%s/USDT 已到达提醒价格，当前价格:%.4f",
		strings.ToUpper(c), price)}
	d.ToParty = ""
	d.ToTag = ""
	d.ToUser = uid
	logs.Info(fmt.Sprintf("%s %s", uid, d.Text))
	return SendMessage(d)
}

func getPrice(c string) (*CurrencyInfo, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	url := "http://xbtc.cx/api/market/getMarketInfo?market=" + c + "/usdt"
	data, err := json.Marshal("")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		//fmt.Println("[resp]", string(body))
		var cInfo Market
		err := json.Unmarshal(body, &cInfo)
		if err != nil {
			fmt.Println("get market errors :", err.Error())
			return nil, err
		}
		var cy CurrencyInfo

		ts, ok := cInfo.Data.(map[string]interface{})
		if !ok {
			return nil, errors.New("暂无该行情数据")
		}
		for i, _ := range ts {
			cy.Currency = i
			ok := false
			cy.High, ok = ts[i].(map[string]interface{})["high"].(float64)
			if !ok {
				return nil, errors.New("暂无行情数据")
			}
			cy.Sell, ok = ts[i].(map[string]interface{})["sell"].(float64)
			if !ok {
				return nil, errors.New("暂无行情数据")
			}
			cy.Buy, ok = ts[i].(map[string]interface{})["buy"].(float64)
			if !ok {
				return nil, errors.New("暂无行情数据")
			}
			cy.Last, ok = ts[i].(map[string]interface{})["last"].(float64)
			if !ok {
				return nil, errors.New("暂无行情数据")
			}
			cy.Low, ok = ts[i].(map[string]interface{})["low"].(float64)
			if !ok {
				return nil, errors.New("暂无行情数据")
			}
		}
		return &cy, nil
	} else {
		return nil, err
	}
}

func TimeTasks() {
	for {
		time.Sleep(2 * time.Second)
		GetCondition()
	}
}

func GetCondition() {
	TimedLock.Lock()
	defer TimedLock.Unlock()

	r, err := getPrice("bhb")
	if err != nil {
		logs.Error(err.Error())
		return
	}

	for i, v := range TimedTask {
		logs.Debug(fmt.Sprintf("BHB/USDT 当前价格[%.4f]", v.Price))
		if v.Direct == "大于" && r.Last >= v.Price {
			SendTimedTaskMessage(v.Uid, v.Currency, r.Last)
			delete(TimedTask, i)
			logs.Info("删除价格提醒", i)
		}
		if r.Last <= v.Price && v.Direct == "小于" {
			SendTimedTaskMessage(v.Uid, v.Currency, r.Last)
			delete(TimedTask, i)
			logs.Info("删除价格提醒", i)
		}
	}
}

func AddTimedTask(currency, direct, price, uid string) bool {
	TimedLock.Lock()
	defer TimedLock.Unlock()
	p, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return false
	}
	if 1000 < len(TimedTask) {
		fmt.Println("清理定价提醒")
		logs.Info("清理定价提醒")
		TimedTask = make(map[string]CurrencyInfo, 0)
	}
	logs.Info(fmt.Sprintf("添加定价提醒[%s][%s][%s][%s]", uid, currency, direct, price))
	TimedTask[strconv.Itoa(int(time.Now().Unix()))] = CurrencyInfo{Currency: currency,
		Price: p, Uid: uid, Direct: direct}
	return true
}
