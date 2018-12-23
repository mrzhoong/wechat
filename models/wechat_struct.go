package models

// MessageType 消息类型定义
type MessageType string

// 各种消息类型值
const (
	// 发送和接收的消息类型

	TextMsg  MessageType = "text"
	ImageMsg MessageType = "image"
	VoiceMsg MessageType = "voice"
	VideoMsg MessageType = "video"

	// 发送的额外消息类型

	FileMsg   MessageType = "file"
	NewsMsg   MessageType = "news"
	MpNewsMsg MessageType = "mpnews"

	// 接收的额外消息类型

	LocationMsg MessageType = "location"
	EventMsg    MessageType = "event"
)

// RecvBaseData 描述接收到的各类消息或事件的公共结构
type RecvBaseData struct {
	ToUserName   string
	FromUserName string
	CreateTime   int
	MsgType      MessageType
	AgentID      int64
}

// RecvTextMessage 描述接收到的文本类型消息结构
type RecvTextMessage struct {
	RecvBaseData
	MsgID   uint64 `xml:"MsgId"`
	Content string
}

// TextContent 为文本类型消息的文本内容
type TextContent struct {
	Content string `json:"content"`
}

// TextMessage 为发送的文本类型消息
type TextMessage struct {
	ToUser  string      `json:"touser,omitempty"`
	ToParty string      `json:"toparty,omitempty"`
	ToTag   string      `json:"totag,omitempty"`
	MsgType MessageType `json:"msgtype"`
	AgentID int64       `json:"agentid"`
	Text    TextContent `json:"text"`
	Safe    int         `json:"safe"`
}
