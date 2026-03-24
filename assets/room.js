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

function addUser(name, color) {
    document.getElementById("chat-users").innerHTML += "<div style='color:" + color + "'>" + name + "</div>"
}

function removeUser(id) {
    document.getElementById("chat-users").querySelector("[data-id=\"" + id + "\"]").remove()
}

var roomId= ""
document.addEventListener("DOMContentLoaded", () => {
    roomId = document.location.pathname.replace("/room/", "");
    console.log("RoomId", roomId)

    fetch("/me")
        .then(resp => resp.json())
        .then(data => {
            if (data.error) {
                alert(data.error)
                return
            }

            initWs(data.id, document.getElementById("message"))
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
                    addUser(data.client.user.name, data.client.user.color)
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

        loadInfo()
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
                addUser(user.name, user.color)
            }
        }).catch(error => console.error("Error loading room", error))
}