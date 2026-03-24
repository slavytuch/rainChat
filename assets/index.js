function appendRoom(id, name, link) {
    document.getElementById("room-list").innerHTML += "<div class='room-row'><a href='"+link+"'>"+name+"</a><button onclick=\"deleteRoom(this, '"+id+"')\">Delete</button>"
}

function deleteRoom(button, id) {
    if (!confirm("Detecting multiple leviathan class lifeforms in the region. Are you certain whatever you're doing is worth it?")) {
        alert("Coward")
        return
    }

    button.parentElement.remove()

    fetch("/delete-room", {
        method: "POST",
        body: JSON.stringify({
            id: id
        })
    })
}

function loginOrRegister() {
    let username = prompt("Login")
    let route = "/login"
    if (!username) {
        username = prompt("Register")
        route = "/register"
    }

    if (!username) {
        alert("Get lost")
        return
    }

    fetch(route, {
        method: "POST",
        body: JSON.stringify({
            name: username
        })
    })
        .then(response => response.json())
        .then((data) => {
            if (data.error) {
                alert(data.error)
                return
            }
            document.cookie = "user-token=" + data.token
            loadMe()
        }).catch(error => console.error("Error during auth", error))
}

function loadMe() {
    fetch("/me")
        .then(resp => resp.json())
        .then(data => {
            document.getElementById("user").innerText = data.user.name
        })
        .catch(error => {
            alert("Error loading user:" + error)
            loginOrRegister()
        })
}

document.addEventListener("DOMContentLoaded", () => {
    let token = getCookie("user-token")

    if (!token) {
        loginOrRegister()
    } else {
        loadMe()
    }

    document.getElementById("create-room-btn").addEventListener("click", () => {
        let roomName = prompt("Name")

        if (!roomName) {
            alert("Thought so")
            return
        }

        fetch("/create-room", {
            method: "POST",
            body: JSON.stringify({
                "name": roomName
            })
        })
            .then(resp => resp.json())
            .then(data => {
                if (data.error) {
                    alert("Error: " + data.error)
                    return
                }

                appendRoom(data.id, data.name, data.link)
            })
    })
})

function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
}