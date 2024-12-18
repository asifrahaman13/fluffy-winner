"use client";
import React, { useState, useEffect, useRef } from "react";

type PageSearch = {
  content: string;
  pageNum: number;
};

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

  const [source, setSource] = useState<PageSearch[] | null>(null);

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

      if (message.msgType === "status") {
        setSource(message.payload);
      }
      if (message.msgType === "server") {
        setMessages((prevMessages) => [
          ...prevMessages,
          {
            clientId: message.clientId,
            messageId: message.messageId,
            payload: message.payload,
            msgType: "server",
          },
        ]);
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
    <div className="flex w-screen gap-8 flex-row items-center justify-center h-screen bg-gray-100 p-4">
      <div className="w-1/3 h-full">
        {source?.map((item, index) => (
          <div key={index} className="flex flex-col ">
            <div className="text-md font-bold">Page: {item.pageNum}</div>
            <div>Content: {item.content}</div>
          </div>
        ))}
      </div>

      <div className="w-1/3 h-full flex flex-col items-center">
        <h1 className="text-2xl font-semibold mb-4 text-gray-800">
          Bhagavad Gita
        </h1>

        <div className="w-full  bg-white shadow-lg rounded-lg p-4 mb-4 overflow-y-auto no-scrollbar ">
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
        <div className="flex w-full">
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
      <div className="w-1/3 h-full">
        <div className="flex flex-col ">
          <div className="font-medium text-lg">About the application</div>
          <div>
            This application leverages the power of Retrieval-Augmented
            Generation (RAG) and Large Language Models (LLMs) to provide
            in-depth knowledge and insights from spiritual texts. Designed for
            researchers, educators, and spiritual seekers, the system integrates
            advanced natural language understanding with a curated database of
            sacred scriptures, philosophical treatises, and spiritual
            commentaries. By combining the contextual retrieval capabilities of
            RAG with the generative reasoning of LLMs, the application ensures
            precise and meaningful answers to user queries. The core
            functionality includes semantic search, question answering, and
            personalized learning paths, allowing users to explore intricate
            spiritual concepts in an intuitive manner. The system also supports
            multilingual input and output, enabling access to wisdom from
            diverse cultural and linguistic traditions. By bridging the gap
            between ancient knowledge and modern AI, this application empowers
            users to gain a deeper understanding of spiritual teachings while
            fostering personal growth and enlightenment.
          </div>
        </div>
      </div>
    </div>
  );
};

export default RealTimeUpdates;
