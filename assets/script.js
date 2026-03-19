function addMessage(message, author, color) {
    document.getElementById("chat").innerHTML += "<span class='author' style='color: "+color+"' >"+author+"</span>" + ": <span class='message'></span>" + message + "<br>"
}

function sendMessage(message, conn) {
    conn.send(JSON.stringify({
        type: "message-send",
        text: message
    }))
}

document.addEventListener("DOMContentLoaded", () => {
    let username = prompt("Login")
    let route = "login"
    if (!username) {
        username = prompt("Register")
        route = "register"
    }

    fetch(route, {
        method: "POST",
        body: JSON.stringify({
            name: username
        })
    })
        .then(response => response.json())
        .then((data) => {
        initWs(data.token, document.getElementById("message"))
    }).catch(error => console.error("Ошибка авторизации", error))
})

function initWs(token, messagebox)
{
    let conn = new WebSocket("ws://" + document.location.host + "/ws?token=" + token)

    conn.onclose = function () {
        addMessage("Conn closed!", "System", "green")
        console.log("closing conn")
    }

    conn.onopen = function () {
        conn.onmessage = function (e) {
            let data = JSON.parse(e.data)

            addMessage(data.text, data.author ?? "who knows", data.color ?? "orange")
        }

        document.getElementById("send").onclick = function () {
            let message = messagebox.value
            messagebox.value = ""

            sendMessage(message, conn)
        }

        document.addEventListener("keyup", function (e) {
            if (e.ctrlKey && e.code === "Enter") {
                let message = messagebox.value
                messagebox.value = ""

                sendMessage(message, conn)
            }
        })
    }
}