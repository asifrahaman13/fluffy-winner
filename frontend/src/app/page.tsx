"use client";
import React, { useState, useEffect } from "react";

type MessageType = {
  messageId: string;
  payload: string;
  msgType: string;
};

const RealTimeUpdates = () => {
  const [messages, setMessages] = useState<MessageType[]>([]);
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [message, setMessage] = useState("");
  const [clientId, setClientId] = useState("");

  useEffect(() => {
    const websocket = new WebSocket("ws://127.0.0.1:8000/ws");
    websocket.onopen = () => {
      console.log("WebSocket is connected");
      const id = Math.floor(Math.random() * 1000);
      setClientId(id.toString());
    };
    websocket.onmessage = (evt) => {
      const message = evt.data;
      console.log(message);
      setMessages((prevMessages) => [
        ...prevMessages,
        {
          messageId: "2",
          payload: message,
          msgType: "server",
        },
      ]);
    };
    websocket.onclose = () => {
      console.log("WebSocket is closed");
    };
    setWs(websocket);
    return () => {
      websocket.close();
    };
  }, []);

  const sendMessage = () => {
    if (ws) {
      ws.send(
        JSON.stringify({
          payload: message,
          clientId: clientId,
        })
      );
      setMessage("");
      setMessages((prevMessages) => [
        ...prevMessages,
        {
          messageId: "1",
          payload: message,
          msgType: "client",
        },
      ]);
    }
  };

  const handleInputChange = (event: {
    target: { value: React.SetStateAction<string> };
  }) => {
    setMessage(event.target.value);
  };

  return (
    <React.Fragment>
      <h1>
        Real-time Updates with WebSockets and React Hooks - Client {clientId}
      </h1>
      {/* <div>{messages.length !== 0 && <div>{messages.toString()}</div>}</div> */}
      <div>
        {messages.map((message, val)=>(
          <>
          <div key={val}>
            {message.payload}
          </div>
          </>
        ))}
      </div>
      <input type="text" value={message} onChange={handleInputChange} />
      <button onClick={sendMessage}>Send Message</button>
    </React.Fragment>
  );
};

export default RealTimeUpdates;
