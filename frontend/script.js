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
    let cart = {"cart": [], "total_cost": 0.0};

    // Stable per-browser id so each visitor gets their own server-side session
    // (history + cart) instead of everyone sharing the hardcoded "user123".
    let userId = localStorage.getItem("chatUserId");
    if (!userId) {
        userId = "user-" + Math.random().toString(36).slice(2) + "-" + Date.now();
        localStorage.setItem("chatUserId", userId);
    }

    // Derive the WebSocket URL from the page origin so it works regardless of
    // host/port instead of being pinned to localhost:8080. Falls back to the Go
    // dev server when the widget is opened directly from the filesystem.
    function chatSocketUrl() {
        if (location.protocol === "http:" || location.protocol === "https:") {
            const proto = location.protocol === "https:" ? "wss:" : "ws:";
            return proto + "//" + location.host + "/api/v1/ws/chat";
        }
        return "ws://localhost:8080/api/v1/ws/chat";
    }

    openChatBtn.onclick = function (event) {
        event.stopPropagation();

        if (!socket || socket.readyState === WebSocket.CLOSED) {
        socket = new WebSocket(chatSocketUrl());

            socket.onopen = () => {
                console.log("Connected to chat!");
            }
            socket.onmessage = (e) => {
                console.log("New message:", e.data);

                try {
                    const messageObject = JSON.parse(e.data);

                    console.log(messageObject);

                    let content = messageObject.content;

                    // safe parse
                    if (typeof content === "string") {
                        content = content
                            .replace(/```json/g, "")
                            .replace(/```/g, "")
                            .replace("\\\"", "\"")
                            .trim();

                        console.log(content);

                        content = JSON.parse(content);
                    }

                    const textBot = content;
                    // console.log("tesdf: ", textBot);
                    // console.log("tesdf: ", textBot.intent);

                    if (textBot["intent"] === "chat") {
                        addMessage(textBot["text"], "bot-message");
                    }
                    else if (textBot["intent"] === "products") {
                        console.log(textBot["payload"]);

                        const products_payload = textBot["payload"]
                        console.log(products_payload)

                        addMessage(textBot["text"], "bot-message");

                        for (let item of products_payload["products"]) {
                            addProductCard(item["product_name"], item["price"]);
                        }
                    }
                    else if (textBot["intent"] === "cart") {
                        console.log(textBot["payload"]);

                        const cart_payload = textBot["payload"]
                        console.log(cart_payload)

                        addMessage(textBot["text"], "bot-message");

                        cart = cart_payload

                        showCart(cart_payload);
                    } else if (textBot["intent"] === "submit") {
                        console.log(textBot["payload"]);
                        addMessage(textBot["text"], "bot-message");

                        cart = {"cart": [], "total_cost": 0.0}
                    }
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
        // Reset the local cart view so a cleared chat starts from a clean slate.
        cart = {"cart": [], "total_cost": 0.0};
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

        if (text.toLowerCase() === "карточка") {
            addProductCard(1);
            return;
        }
        
        if (text.toLowerCase() === "карточка2") {
            addProductCard(2);
            return;
        }
        
        const userMessage = {
            // RFC3339 string — the Go backend's time.Time field rejects a raw
            // numeric timestamp. (The server overwrites it server-side anyway.)
            timestamp: new Date().toISOString(),
            messageType: "wh_message",
            content: text,
            sender: userId,
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
    
            if (text === "Корзина") {
                showCart(cart);
                return;
            }
    
            addMessage(text, "user-message");
    
            setTimeout(function () {
                addMessage("Вы выбрали: " + text, "bot-message");
            }, 500);
        };
    });

    /// @prudct_name string
    /// @product_price float
    function addProductCard(product_name, product_price) {
        const card = document.createElement("div");
        card.classList.add("product-card");
    
        card.innerHTML = `
            <div class="product-content">
                <div class="product-title">${product_name}</div>
    
                <div class="product-bottom">
                    <div class="product-price">
                        Цена: ${product_price} ₸
                    </div>
                </div>
            </div>
        `;
    
        chatMessages.appendChild(card);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }

    function addCartProductTotalPrice(price) {
        const label = document.createElement("div");
        label.classList.add("product-card");

        label.innerHTML = `
            <div class="product-content">
                <div class="product-price">
                    Итоговая цена: ${price} ₸
                </div>
            </div>
        `;

        chatMessages.appendChild(label);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }

    // function addToCart(product) {
    //     const existingProduct = cart.find(function (item) {
    //         return item.id === product.id;
    //     });
    
    //     if (existingProduct) {
    //         existingProduct.quantity++;
    //     } else {
    //         cart.push({
    //             ...product,
    //             quantity: 1
    //         });
    //     }
    
    //     addMessage("Товар добавлен в корзину", "bot-message");
    // }

    function showCart(cart_payload) {

        let order_list = cart_payload["cart"];

        // Remove any previously rendered cart so repeated "Корзина" clicks don't
        // stack duplicate cart windows.
        const oldCart = chatMessages.querySelector(".cart-window");
        if (oldCart) {
            oldCart.remove();
        }

        const cartWindow = document.createElement("div");
        cartWindow.classList.add("cart-window");
    
        if (order_list.length === 0) {
            cartWindow.innerHTML = `
                <div class="cart-title">Корзина</div>
                <div class="cart-empty">Корзина пока пустая</div>
            `;
        } else {
            let total = 0;
    
            let itemsHtml = order_list.map(function (item) {
                total += item.price * item.quantity;
    
                return `
                <div class="cart-item">
            
                    <div class="cart-main">
            
                        <div>
                            <div class="cart-product-name">
                                ${item.product_name}
                            </div>
            
                            <div class="cart-product-price">
                                ${item.price} ₸ × ${item.quantity}
                            </div>
                        </div>
            
                        <div class="cart-location-wrap">
                            <button class="cart-details-btn">
                                <img src="geo2.svg" class="details-icon">
                            </button>
            
                            <div class="cart-details">
                                 ${item.storage_address}
                            </div>
                        </div>
            
                    </div>
            
                </div>
            `;
    
            }).join("");
    
            cartWindow.innerHTML = `
                <div class="cart-title">Корзина</div>
                ${itemsHtml}
                <div class="cart-total">
                    Итоговая цена: ${cart_payload["total_cost"]} ₸
                </div>
            `;

        // for (let item of cart_payload["cart"]) {
        //     addProductCard(item["product_name"], item["price"]);
        // }
        // addCartProductTotalPrice(cart_payload["total_cost"]);
    }
    
    // function showCart() {
    //     const oldCart = document.querySelector(".cart-window");
    
    //     if (oldCart) {
    //         oldCart.remove();
    //     }
    
    //     const cartWindow = document.createElement("div");
    //     cartWindow.classList.add("cart-window");
    
    //     if (cart.length === 0) {
    //         cartWindow.innerHTML = `
    //             <div class="cart-title">Корзина</div>
    //             <div class="cart-empty">Корзина пока пустая</div>
    //         `;
    //     } else {
    //         let total = 0;
    
    //         let itemsHtml = cart.map(function (item) {
    //             total += item.price * item.quantity;
    
    //             return `
    //             <div class="cart-item">
            
    //                 <div class="cart-main">
            
    //                     <div>
    //                         <div class="cart-product-name">
    //                             ${item.name}
    //                         </div>
            
    //                         <div class="cart-product-price">
    //                             ${item.price} ₸ × ${item.quantity}
    //                         </div>
    //                     </div>
            
    //                     <div class="cart-location-wrap">
    //                         <button class="cart-details-btn">
    //                             <img src="geo2.svg" class="details-icon">
    //                         </button>
            
    //                         <div class="cart-details">
    //                              ${item.address}
    //                         </div>
    //                     </div>
            
    //                 </div>
            
    //             </div>
    //         `;
    
    //         }).join("");
    
    //         cartWindow.innerHTML = `
    //             <div class="cart-title">Корзина</div>
    //             ${itemsHtml}
    //             <div class="cart-total">
    //                 Итоговая цена: ${total} ₸
    //             </div>
    //         `;
    //     }
    
        chatMessages.appendChild(cartWindow);
    
    
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }


});