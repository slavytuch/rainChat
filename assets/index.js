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

document.addEventListener("DOMContentLoaded", () => {
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