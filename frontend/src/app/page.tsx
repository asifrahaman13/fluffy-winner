"use client";
import React, { useState, useEffect, useRef } from "react";

type MessageType = {
  clientId: string;
  messageId: number;
  payload: string;
  msgType: string; 
};

const RealTimeUpdates = () => {
  const [messages, setMessages] = useState<MessageType[]>([]);
  const [message, setMessage] = useState("");
  const [clientId, setClientId] = useState("");
  const messageIdRef = useRef<number>(0); 
  const wsRef = useRef<WebSocket | null>(null); 
  const messagesRef = useRef<MessageType[]>(messages);

  useEffect(() => {
    const websocket = new WebSocket("ws://127.0.0.1:8000/ws");
    wsRef.current = websocket; 

    websocket.onopen = () => {
      console.log("WebSocket is connected");
      const id = Math.floor(Math.random() * 1000);
      setClientId(id.toString());
    };

    websocket.onmessage = (evt) => {
      const message = JSON.parse(evt.data);
      console.log("Received message:", message.messageId);
      console.log("Current Message ID:", messageIdRef.current);
      console.log(messagesRef.current);
      if (message.messageId === messageIdRef.current) {
        setMessages((prevMessages) => {
          const updatedMessages = [...prevMessages];
          const existingIndex = updatedMessages.findIndex(
            (msg) => msg.messageId === message.messageId
          );
          if (existingIndex === -1) {
            console.log("Does not exist");
            updatedMessages.push({
              payload: message.payload,
              clientId: message.clientId,
              messageId: message.messageId,
              msgType: "server",
            });
          } else {
            const updatedPayload =
              updatedMessages[existingIndex].payload + message.payload;
            updatedMessages[existingIndex].payload = updatedPayload;
          }
          messagesRef.current = updatedMessages;
          return updatedMessages;
        });
      }
    };

    websocket.onclose = () => {
      console.log("WebSocket is closed");
    };

    return () => {
      websocket.close();
    };
  }, []); 

  const sendMessage = () => {
    if (wsRef.current) {
      setMessages((prevMessages) => [
        ...prevMessages,
        {
          clientId: clientId,
          messageId: 0,
          payload: message,
          msgType: "client", 
        },
      ]);
      const newMessageId = messageIdRef.current + 1; 
      messageIdRef.current = newMessageId; 
      const messageSend = {
        clientId: clientId,
        messageId: newMessageId,
        payload: message,
        msgType: "client", 
      };
      wsRef.current.send(JSON.stringify(messageSend));
      setMessage("");
    }
  };

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setMessage(event.target.value); 
  };

  return (
    <React.Fragment>
      <h1>
        Real-time Updates with WebSockets and React Hooks - Client {clientId}
      </h1>

      <div>
        {messages.map((msg, index) => (
          <div key={index} className="flex flex-col">
            <div
              className={
                msg.msgType === "client" ? "bg-red-200" : "bg-green-200"
              }
            >
              <strong>{msg.msgType === "client" ? "You: " : "Server: "}</strong>
              {msg.payload}
            </div>
          </div>
        ))}
      </div>
      <input type="text" value={message} onChange={handleInputChange} />
      <button onClick={sendMessage}>Send Message</button>
    </React.Fragment>
  );
};

export default RealTimeUpdates;
