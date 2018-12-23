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
	NoticeMap  map[int64]time.Time
	NoticeLock sync.Mutex
)

type ResponseNotice struct {
	Code    string      `json:"code"`
	Success bool        `json:"success"`
	Msg     interface{} `json:"msg"`
	Data    Notice      `json:"data"`
}
type Notice struct {
	NewCoinList   []CoinList    `json:"newCoinList"`
	NewNoticeList []interface{} `json:"newNoticeList"`
}

type CoinList struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	ReleaseTime string `json:"releaseTime"`
	Type        int64  `json:"type"`
}

var log *logs.BeeLogger

func init() {
	NoticeMap = make(map[int64]time.Time)
	time.Sleep(4 * time.Second)
	logs.Info("今日公告监控已启用")
	go GoroutineGetNoticeList()
}

func GoroutineGetNoticeList() {
	for {
		time.Sleep(5 * time.Second)
		logs.Debug("获取今日公告")
		GetNoticeList()
	}
}

func GetNoticeList() (error, string) {
	// http
	client := http.Client{
		Timeout: 15 * time.Second,
	}
	// 目标地址
	tarAddr := "https://xbtc.cx/noticeCenter/list"
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
		var ret ResponseNotice
		err = json.Unmarshal(body, &ret)
		if err != nil {
			return err, ""
		}

		if ret.Success != true {
			return errors.New("no jur project11"), ""
		}
		AnalyzeNotice(ret.Data.NewCoinList)
		return nil, string(body)
	} else {
		logs.Error("获取今日公告请求失败")
		return errors.New(fmt.Sprintf("%s time out", time.Now().Format("2006-01-02 15:04:05"))), ""
	}
}

func AnalyzeNotice(c []CoinList) {
	NoticeLock.Lock()
	defer NoticeLock.Unlock()
	tm := time.Now()
	for _, v := range c {
		LastNotice, _ := time.ParseInLocation("2006-01-02 15:04:05", v.ReleaseTime, time.Local)
		LastNotice = LastNotice.Add(60 * time.Minute)
		logs.Debug(fmt.Sprintf("当前时间[%s] 公告时间[%s]",
			tm.Format("2006-01-02 15:04:05"),
			LastNotice.Format("2006-01-02 15:04:05")))

		_, ok := NoticeMap[v.Id]
		// 公告一小时以内、且首次发布
		if LastNotice.Unix() > tm.Unix() && !ok {
			// 发送消息
			createNoticeMessage(v)
			logs.Info(fmt.Sprintf("发送公告提醒[%d]", v.Id))
			// 记录已发送记录
			NoticeMap[v.Id] = tm
		}
	}

	for i, _ := range NoticeMap {
		DelTime := NoticeMap[i].Add(70 * time.Minute)
		if DelTime.Unix() < time.Now().Unix() {
			logs.Info(fmt.Sprintf("删除已提醒公告[%d]", i))
			delete(NoticeMap, i)
		}
	}

	// 未完成，只存当天的新闻
}

// 发送大额成交记录
func createNoticeMessage(c CoinList) error {
	var d TextMessage
	d.AgentID = 1000002
	d.MsgType = TextMsg
	d.Safe = 0
	d.Text = TextContent{Content: fmt.Sprintf("【公告提醒】\n发布时间:%s\n详情请单击：<a href=\"https://xbtc.cx/#/notice/notice_detail/%d/1\">新公告</a>",
		c.ReleaseTime, c.Id)}
	d.ToParty = ""
	d.ToTag = ""
	d.ToUser = "@all"
	//d.ToUser = "Fu"
	logs.Info(fmt.Sprintf("%v", d.Text))
	err := SendMessage(d)
	if err != nil {
		logs.Error(fmt.Sprintf("新公告发送错误%s", err.Error()))
	}
	return err
}
