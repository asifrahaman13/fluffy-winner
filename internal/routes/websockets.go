package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"github.com/asifrahaman13/bhagabad_gita/internal/config"
	"github.com/gorilla/websocket"
	"github.com/asifrahaman13/bhagabad_gita/internal/helper"
)

type ChatResponse struct {
	Response string `json:"response"`
}

type WebsocketMessage struct {
	ClientId  string `json:"clientId"`
	MessageId int    `json:"messageId"`
	Payload   string `json:"payload"`
	MsgType   string `json:"msgType"`
}

func chatBotResponse(prompt WebsocketMessage, conn *websocket.Conn) {
	config, err := config.NewConfig()
	if err != nil {
		fmt.Println("Error getting config:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error getting config"))
		return
	}
	postUrl := config.LLamaUrl
	fmt.Printf("The message is from the client: %s and the client is: %s, message id is: %d, message type is: %s", prompt.Payload, prompt.ClientId, prompt.MessageId, prompt.MsgType)
	body := []byte(fmt.Sprintf(`{
		"model": "llama3.1",
		"stream": true,
		"prompt": "%s"
	}`, prompt.Payload))

	req, err := http.NewRequest("POST", postUrl, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error creating request:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error creating request"))
		return
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error making request"))
		return
	}
	defer res.Body.Close()

	var buffer strings.Builder
	for {
		var chatResponse ChatResponse
		err = json.NewDecoder(res.Body).Decode(&chatResponse)
		if err != nil {
			fmt.Println("Error decoding response:", err)
			conn.WriteMessage(websocket.TextMessage, []byte(""))
			return
		}
		buffer.WriteString(chatResponse.Response)
		if helper.IsSentenceEnd(*bytes.NewBufferString(buffer.String())) {
			textMessages := WebsocketMessage{
				ClientId:  prompt.ClientId,
				MessageId: prompt.MessageId,
				Payload:   buffer.String(),
				MsgType:   "server",
			}

			jsonStringMessage, err := json.Marshal(textMessages)
			if err != nil {
				fmt.Println("Error marshaling message:", err)
				conn.WriteMessage(websocket.TextMessage, []byte("Error marshaling message"))
				return
			}

			err = conn.WriteMessage(websocket.TextMessage, jsonStringMessage)
			if err != nil {
				fmt.Println("Error sending message:", err)
				conn.Close()
				return
			}
			buffer.Reset()
		}
	}
}

func HandleWebSocketConnection(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			conn.Close()
			break
		}
		var messageStruct WebsocketMessage
		err = json.Unmarshal(message, &messageStruct)
		if err != nil {
			fmt.Println("Error decoding message:", err)
		}
		go chatBotResponse(messageStruct, conn)
	}
}
