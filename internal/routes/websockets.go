package routes

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"encoding/json"
	"github.com/asifrahaman13/bhagabad_gita/internal/config"
)

type ChatResponse struct {
	Response string `json:"response"`
}

func chatBotResponse(prompt string, conn *websocket.Conn) {
	config  , err:= config.NewConfig()
	if err != nil {
		fmt.Println("Error getting config:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error getting config"))
		return
	}
	postUrl := config.LLamaUrl
	body := []byte(fmt.Sprintf(`{
		"model": "llama3.1",
		"stream": false,
		"prompt": "%s"
	}`, prompt))

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
	var chatResponse ChatResponse
	err = json.NewDecoder(res.Body).Decode(&chatResponse)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error decoding response"))
		return
	}
	err = conn.WriteMessage(websocket.TextMessage, []byte(chatResponse.Response))
	if err != nil {
		fmt.Println("Error sending message:", err)
		conn.Close()
		return
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
		go chatBotResponse(string(message), conn)
	}
}