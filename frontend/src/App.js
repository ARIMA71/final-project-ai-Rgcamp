import React, { useState, useEffect, useRef } from "react";
import axios from "axios";
import './index.css';
import bgImage from './bg.jpg';

function App() {
  const backgroundImage = bgImage;
  const [file, setFile] = useState(null);
  const [query, setQuery] = useState("");
  const [message, setMessage] = useState([]);
  const messagesEndRef = useRef(null);

  const handleFileChange = (e) => {
    setFile(e.target.files[0]);
  };

  const handleRemoveFile = () => {
    setFile(null);
  };

  const handleChat = async () => {
    const formData = new FormData();
    if (file) {
      formData.append("file", file);
      formData.append("in_query", query);
    } else {
      formData.append("query", query);
    }

    try {
      const url = file ? 'http://localhost:8080/upload' : 'http://localhost:8080/chat';
      const res = await axios.post(url, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
      const newMessage = { type: 'answer', text: res.data.answer };
      setMessage(prevMessages => [...prevMessages, { type: 'question', text: query }, newMessage]);
      setQuery("");
    } catch (error) {
      console.error("Error querying chat:", error);
      const errorMessage = { type: 'error', text: error.response?.data || "An error occurred" };
      setMessage(prevMessages => [...prevMessages, { type: 'question', text: query }, errorMessage]);
      setQuery("");
    }
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [message]);

  return (
    <div className="h-screen bg-cover bg-gradient-to-b from-blue-500 to-white" style={{ backgroundImage: `url(${backgroundImage})` }} >
      <div className="h-screen bg-white shadow-2xl rounded-lg max-w-4xl mx-auto p-7">
        <h1 className="text-3xl font-bold text-center text-white bg-blue-600 py-4 rounded-lg mb-5">
          Analytics AI Chat
        </h1>
        <div className="h-4/5 bg-white rounded-lg p-4">
          <h2 className="text-xl font-semibold mb-4 text-blue-600">
            AI Response
          </h2>
          <div className="overflow-y-auto mt-4 h-5/6 p-4 bg-gray-100 rounded-lg shadow-inner">
            {message.map((message, index) => (
              <div
                key={index}
                className={`py-2 px-4 my-3 rounded-lg w-fit ${
                  message.type === "question"
                    ? "bg-blue-200 text-right ml-auto"
                    : message.type === "answer"
                    ? "bg-white text-left mr-auto"
                    : "bg-red-200 text-right mr-auto"
                }`}
              >
                {message.text}
              </div>
            ))}
            <div ref={messagesEndRef} />
          </div>
        </div>
        <div className="flex items-center">
          <input
            id="file-upload"
            type="file"
            onChange={handleFileChange}
            className="hidden"
          />
          <label
            htmlFor="file-upload"
            className="flex-shrink-0 p-2 px-4 bg-blue-600 text-white rounded-lg cursor-pointer mr-4 hover:bg-blue-700"
          >
            +
          </label>
          {file && (
            <span className="mr-5 flex items-center">
              {file.name}
              <button
                onClick={handleRemoveFile}
                className="ml-2 text-red-500 hover:text-red-700"
              >
                ✕
              </button>
            </span>
          )}
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Ask a question..."
            className="p-2 flex-1 border border-gray-300 rounded-lg"
          />
          <button
            onClick={handleChat}
            className="flex-shrink-0 p-2 px-4 ml-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Chat
          </button>
        </div>
      </div>
    </div>
  );
}

export default App;