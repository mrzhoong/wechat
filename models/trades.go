package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/astaxie/beego/logs"
)

var (
	TradeMap  map[int64]interface{}
	TradeLock sync.Mutex
)

type Response struct {
	Code    string      `json:"code"`
	Success bool        `json:"success"`
	Msg     interface{} `json:"msg"`
	Data    []Trade     `json:"data"`
}

type Trade struct {
	Id           int64   `json:"id"`           //市场id
	Amount       float64 `json:"amount"`       //销售额
	Qty          float64 `json:"qty"`          //销售数量
	Price        float64 `json:"price"`        //市场价
	Type         string  `json:"type"`         //类型 1buy 2sell
	CreateTime   string  `json:"createTime"`   //创建时间
	CreateTimeMs int64   `json:"createTimeMs"` //创建时间毫秒值
}

func init() {
	TradeMap = make(map[int64]interface{})
	time.Sleep(3 * time.Second)
	logs.Info("大额成交监控已启用")
	go GoroutineGetTrade()
}

func GoroutineGetTrade() {
	for {
		time.Sleep(5 * time.Second)
		logs.Debug("获取交易记录")
		GetTrades()
	}
}

func GetTrades() (error, string) {
	// http
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	// 目标地址
	tarAddr := "https://xbtc.cx/api/market/getTrades?market=bhb/usdt"
	// 请求
	req, err := http.NewRequest(
		"POST",
		tarAddr,
		bytes.NewReader([]byte{}))
	if err != nil {
		return err, ""
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err, ""
	}

	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		var ret Response
		err = json.Unmarshal(body, &ret)
		if err != nil {
			return err, ""
		}

		if ret.Success != true {
			return errors.New("no jur project11"), ""
		}
		AnalyzeSlice(ret.Data)
		return nil, string(body)
	} else {
		logs.Error("大额成交监控请求失败")
		return errors.New(fmt.Sprintf("%s time out", time.Now().Format("2006-01-02 15:04:05"))), ""
	}
}

func AnalyzeSlice(t []Trade) {
	TradeLock.Lock()
	defer TradeLock.Unlock()

	for _, v := range t {
		if v.Amount > 5000.0 {
			_, ok := TradeMap[v.CreateTimeMs]
			if !ok {
				// 发送消息
				createTradeMessage(v)
				// 记录已发送记录
				TradeMap[v.CreateTimeMs] = ""
			}
		}
	}
	for i, _ := range TradeMap {
		tm := time.Unix(i/1000, 0)
		tm = tm.Add(6 * time.Hour)
		if tm.Unix() <= time.Now().Unix() {
			delete(TradeMap, i)
			logs.Info("删除超过12小时交易记录")
		}
	}
	if 1000 == len(TradeMap) {
		TradeMap = make(map[int64]interface{})
		logs.Info("删除所有交易记录")
	}
}

// 发送大额成交记录
func createTradeMessage(t Trade) error {
	var d TextMessage
	d.AgentID = 1000002
	d.MsgType = TextMsg
	d.Safe = 0
	if t.Type == "buy" {
		d.Text = TextContent{Content: fmt.Sprintf("%s 出现大额买入，成交价:%.4f，需求总量:%.4f，当前成交数量:%.4f",
			t.CreateTime, t.Price, t.Amount, t.Qty)}
	} else {
		d.Text = TextContent{Content: fmt.Sprintf("%s 出现大额卖出，成交价:%.4f，需求总量:%.4f，当前成交数量:%.4f",
			t.CreateTime, t.Price, t.Amount, t.Qty)}
	}
	d.ToParty = ""
	d.ToTag = ""
	d.ToUser = "@all"
	logs.Info(fmt.Sprintf("%v", d.Text))
	err := SendMessage(d)
	if err != nil {
		logs.Error(fmt.Sprintf("发送大额成交信息%s", err.Error()))
	}
	return err
}
