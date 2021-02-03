import socketio

sio = socketio.Client()
sio.connect("http://localhost:8000", headers={"X-Packetframe-ECA-Auth": "test123"})


@sio.event
def connect():
    print("connected")
    print('my sid is', sio.sid)


@sio.event
def message(data):
    print('I received a message!')


@sio.on('my message')
def on_message(data):
    print('I received a message!')
