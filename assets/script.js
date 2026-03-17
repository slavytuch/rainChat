function addMessage(message, author) {
    document.getElementById("chat").innerHTML += "<span class='author'>"+author+"</span>" + ": <span class='message'></span>" + message + "<br>"
}

function sendMessage(message, conn) {
    conn.send(JSON.stringify({
        type: "message-send",
        text: message
    }))
}

document.addEventListener("DOMContentLoaded", () => {
    var username = prompt("Enter username")
    let messagebox = document.getElementById("message")

    if (!username) {
        alert("Yeah, nah, get lost")
        return;
    }

    fetch("login", {
        method: "POST",
        body: JSON.stringify({
            name: username
        })
    })
        .then(response => response.json())
        .then((data) => {
        initWs(data.token, messagebox)
    }).catch(error => console.error("Ошибка авторизации", error))
})

function initWs(token, messagebox)
{
    let conn = new WebSocket("ws://" + document.location.host + "/ws?token=" + token)

    conn.onclose = function () {
        addMessage("Conn closed!", "System")
    }

    conn.onopen = function () {
        conn.onmessage = function (e) {
            let data = JSON.parse(e.data)

            addMessage(data.text, data.author ?? "who knows")
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