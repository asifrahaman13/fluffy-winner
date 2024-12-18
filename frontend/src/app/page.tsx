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

  const websocket = new WebSocket("ws://127.0.0.1:8000/ws");
  useEffect(() => {
    wsRef.current = websocket;

    websocket.onopen = () => {
      console.log("WebSocket is connected");
      const id = Math.floor(Math.random() * 1000);
      setClientId(id.toString());
    };

    websocket.onmessage = (evt) => {
      const message = JSON.parse(evt.data);

      setMessages((prevMessages) => [
        ...prevMessages,
        {
          clientId: message.clientId,
          messageId: message.messageId,
          payload: message.payload,
          msgType: "server",
        },
      ]);
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
    <div className="flex flex-col items-center justify-center h-screen bg-gray-100 p-4">
      <h1 className="text-2xl font-semibold mb-4 text-gray-800">
        Bhagavad Gita
      </h1>

      <div className="w-full max-w-md bg-white shadow-lg rounded-lg p-4 mb-4 overflow-y-auto no-scrollbar ">
        {messages.map((msg, index) => (
          <div key={index} className="flex flex-col mb-2">
            <div className={`p-2 rounded-lg `}>
              {msg.msgType === "client" && (
                <div className="flex flex-row justify-end ">
                  <div className="bg-gray-100 rounded-xl p-3">
                    {msg.payload}
                  </div>
                </div>
              )}

              {msg.msgType === "server" && <>{msg.payload}</>}
            </div>
          </div>
        ))}
      </div>

      <div className="flex w-full max-w-md">
        <input
          type="text"
          value={message}
          onChange={handleInputChange}
          className="flex-grow p-2 border border-gray-300 rounded-l-lg focus:outline-none "
          placeholder="Type your message..."
        />
        <button
          onClick={sendMessage}
          className="bg-gray-500 text-white p-2 rounded-r-lg transition"
        >
          Send
        </button>
      </div>
    </div>
  );
};

export default RealTimeUpdates;
