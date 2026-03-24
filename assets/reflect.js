const pc = new RTCPeerConnection({
    iceServers: [
        {
            urls: 'stun:stun.l.google.com:19302'
        }
    ]
})
const log = msg => {
    document.getElementById('logs').innerHTML += msg + '<br>'
}

navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    .then(stream => {
        stream.getTracks().forEach(track => pc.addTrack(track, stream))
        pc.createOffer().then(d => pc.setLocalDescription(d)).catch(log)
    }).catch(log)

pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)
pc.onicecandidate = event => {
    if (event.candidate === null) {
        console.log("local description", pc.localDescription)
        document.getElementById("start-session").removeAttribute("disabled")
    }
}
pc.ontrack = function (event) {
    console.log("ontrack", event)

    const el = document.createElement(event.track.kind)
    el.srcObject = event.streams[0]
    el.autoplay = true
    el.controls = true

    document.getElementById('remoteVideos').appendChild(el)
}

window.startSession = () => {
    try {
        fetch("/reflect-connect", {
            method: "POST",
            body: JSON.stringify(pc.localDescription)
        })
            .then(resp => resp.json())
            .then(async data => {
                console.log("Resp", data)
                await pc.setRemoteDescription(data)
            })
            .catch(error => console.error("Error connecting", error))
    } catch (e) {
        alert(e)
    }
}