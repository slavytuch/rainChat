function addMessage(message, author, color) {
    document.getElementById("chat").innerHTML += "<span class='author' style='color: " + color + "' >" + author + "</span>" + ": <span class='message'></span>" + message + "<br>"
}

function systemMessage(message) {
    addMessage(message, "System", "red")
}

function sendMessage(message, conn) {
    conn.send(JSON.stringify({
        type: "message-send",
        text: message
    }))
}

function addUser(id, name, color) {
    document.getElementById("chat-users").innerHTML += "<div style='color:" + color + "' data-id='" + id + "'>" + name + "</div>"
}

function removeUser(id) {
    document.getElementById("chat-users").querySelector("[data-id=\"" + id + "\"]").remove()
}

var roomId= ""
document.addEventListener("DOMContentLoaded", () => {
    roomId = document.location.pathname.replace("/room/", "");
    console.log("RoomId", roomId)

    let username = prompt("Login")
    let route = "/login"
    if (!username) {
        username = prompt("Register")
        route = "/register"
    }

    fetch(route, {
        method: "POST",
        body: JSON.stringify({
            name: username
        })
    })
        .then(response => response.json())
        .then((data) => {
            addUser(data.token, data.name, data.color)
            initWs(data.token, document.getElementById("message"))
            loadInfo()
        }).catch(error => console.error("Ошибка авторизации", error))
})

function initWs(token, messagebox) {
    let conn = new WebSocket("ws://" + document.location.host + document.location.pathname + "/ws?token=" + token)

    conn.onclose = function () {
        addMessage("Conn closed!", "System", "green")
        console.log("closing conn")
    }

    conn.onopen = function () {
        conn.onmessage = function (e) {
            let data = JSON.parse(e.data)

            if (data.error) {
                alert("Error:" + data.error)

                conn.close()
                return
            }

            switch (data.type) {
                case 'message-send':
                    addMessage(data.message.text, data.message.author ?? "who knows", data.message.color ?? "orange")
                    break;
                case 'connect':
                    addUser(data.client.id, data.client.user.name, data.client.user.color)
                    systemMessage("User " + data.client.user.name + " connected" )
                    break
                case 'disconnect':
                    removeUser(data.client.id)
                    systemMessage("User " + data.client.user.name + " disconnected" )
                    break
                default:
                    console.log("Unknown message type")
            }
        }

        document.getElementById("send").onclick = function () {
            let message = messagebox.value

            if (!message) {
                return
            }

            messagebox.value = ""

            sendMessage(message, conn)
        }

        document.addEventListener("keyup", function (e) {
            if (e.ctrlKey && e.code === "Enter") {
                let message = messagebox.value
                if (!message) {
                    return
                }

                messagebox.value = ""

                sendMessage(message, conn)
            }
        })
    }
}

function loadInfo() {
    fetch( document.location.href + "/info")
        .then(response => response.json())
        .then(resp => {
            if (!resp.userList) {
                return
            }

            for (user of resp.userList) {
                addUser(user.id, user.name, user.color)
            }
        }).catch(error => console.error("Error loading room", error))
}