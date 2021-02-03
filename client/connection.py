import socketio

sio = socketio.Client()

eca_id = "601496f63d405ca2e88d8c6e"

print("Connecting to control plane")
try:
    sio.connect("http://localhost:8000", headers={"X-Packetframe-ECA-Auth": eca_id})
except socketio.exceptions.ConnectionError as e:
    print(e)
    exit(1)
else:
    print("Connected to control plane")


@sio.event
def connect():
    print("connected")


@sio.event
def disconnect():
    print("disconnected")
    sio.disconnect()


@sio.on
def message(data):
    print(data)


@sio.on("metadata")
def on_metadata(data):
    print("metadata", data)


@sio.on("global_ping")
def on_global_ping():
    sio.emit("global_pong")
    print("ping")


@sio.on("terminate")
def on_reject(data):
    """
    Forcefully terminate the connection
    :param data:
    :return:
    """
    print("terminating connection", data)
    sio.disconnect()
    exit(2)
