package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/asifrahaman13/bhagabad_gita/internal/config"
	"github.com/gorilla/websocket"
	"net/http"
)

type ChatResponse struct {
	Response string `json:"response"`
}

type WebsocketMessage struct {
    ClientId string `json:"clientId"`
	MessageId string `json:"messageId"`
	Payload  string `json:"payload"`
	MsgType string `json:"msgType"`
}

func chatBotResponse(prompt WebsocketMessage, conn *websocket.Conn) {
	config, err := config.NewConfig()
	if err != nil {
		fmt.Println("Error getting config:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error getting config"))
		return
	}
	postUrl := config.LLamaUrl
	fmt.Printf("The message is from the client: %s and the client is: %s", prompt.Payload, prompt.ClientId)
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
	for {
		var chatResponse ChatResponse
		err = json.NewDecoder(res.Body).Decode(&chatResponse)
		if err != nil {
			fmt.Println("Error decoding response:", err)
			conn.WriteMessage(websocket.TextMessage, []byte(""))
			return
		}
		err = conn.WriteMessage(websocket.TextMessage, []byte(chatResponse.Response))
		if err != nil {
			fmt.Println("Error sending message:", err)
			conn.Close()
			return
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
