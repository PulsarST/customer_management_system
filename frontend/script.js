document.addEventListener("DOMContentLoaded", function () {

    const openChatBtn = document.getElementById("openChat");
    const closeChatBtn = document.getElementById("closeChat");
    const clearChatBtn = document.getElementById("clearChat");

    const chatWindow = document.getElementById("chat");
    const sendBtn = document.getElementById("sendBtn");
    const messageInput = document.getElementById("messageInput");
    const chatMessages = document.getElementById("messages");

    const quickButtons = document.querySelectorAll(".quick-questions button");

    let socket = null;

    openChatBtn.onclick = function (event) {
        event.stopPropagation();

        if (!socket || socket.readyState === WebSocket.CLOSED) {
            socket = new WebSocket("ws://localhost:8080/api/v1/ws/chat");

            socket.onopen = () => {
                console.log("Connected to chat!");
            }
            socket.onmessage = (e) => {
                console.log("New message:", e.data);

                try {
                    const messageObject = JSON.parse(e.data);
                    addMessage(messageObject["content"], "bot-message")       
                } catch (error) {
                    console.error("Error parsing server JSON:", error);
                }
            }
            socket.onerror = (err) => console.error("Socket error:", err);
            socket.onclose = () => console.log("Chat connection closed.");
        }

        chatWindow.style.display = "flex";
    };

    chatWindow.onclick = function (event) {
        event.stopPropagation();
    };

    document.addEventListener("click", function () {
        chatWindow.style.display = "none";
    });

    closeChatBtn.onclick = function () {
        chatWindow.style.display = "none";
        
        if (socket)
            socket.close();
    };

    clearChatBtn.onclick = function () {
        chatMessages.innerHTML = `
            <div class="bot-message">
                Здравствуйте! Чем я могу помочь?
            </div>
        `;
    };

    sendBtn.onclick = function () {
        sendMessage();
    };

    messageInput.addEventListener("keydown", function (event) {
        if (event.key === "Enter") {
            sendMessage();
        }
    });

    function sendMessage() {
        const text = messageInput.value.trim();

        if (text === "") {
            return;
        }

        addMessage(text, "user-message");

        messageInput.value = "";
        
        const userMessage = {
            timestamp: Date.now,
            messageType: "wh_message",
            content: text,
            sender: "user123",
            status: 200
        };

        if (socket && socket.readyState === WebSocket.OPEN)
            socket.send(JSON.stringify(userMessage));
    }

    function addMessage(text, className) {
        const message = document.createElement("div");

        message.classList.add(className);
        message.textContent = text;

        chatMessages.appendChild(message);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }

    quickButtons.forEach(function (button) {
        button.onclick = function () {
            const text = button.textContent.trim();

            addMessage(text, "user-message");

            setTimeout(function () {
                addMessage("Вы выбрали: " + text, "bot-message");
            }, 500);
        };
    });

});