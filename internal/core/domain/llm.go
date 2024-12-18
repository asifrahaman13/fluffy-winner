package domain

type Query struct {
	Search string `json:"search" bson:"search"`
}

type ChatResponse struct {
	Response string `json:"response"`
}

type WebsocketMessage struct {
	ClientId  string `json:"clientId"`
	MessageId int    `json:"messageId"`
	Payload   string `json:"payload"`
	MsgType   string `json:"msgType"`
}
type VectorSearchResult struct {
	PageNum uint64 `json:"rank"`
	Content string `json:"content"`
}