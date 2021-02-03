import socketio
from rich.console import Console

console = Console()
sio = socketio.Client()

eca_id = "601496f63d405ca2e88d8c6e"

console.log("Connecting to control plane")
try:
    sio.connect("http://localhost:3000", headers={"X-Packetframe-ECA-Auth": eca_id})
except socketio.exceptions.ConnectionError as e:
    console.log(e)
    exit(1)
else:
    console.log("Connected to control plane")


@sio.event
def connect():
    console.log("connected")


@sio.event
def disconnect():
    console.log("disconnected")
    sio.disconnect()


@sio.on
def message(data):
    console.log(data)


@sio.on("metadata")
def on_metadata(data):
    console.log("metadata", data)


@sio.on("global_ping")
def on_global_ping():
    sio.emit("global_pong")
    console.log("ping")


@sio.on("terminate")
def on_reject(data):
    """
    Forcefully terminate the connection
    :param data:
    :return:
    """
    console.log("terminating connection", data)
    sio.disconnect()
    exit(2)
