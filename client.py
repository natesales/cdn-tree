import json
from os import environ

from websocket import create_connection
from rich.console import Console

console = Console()

node_id = environ.get("CDNV3_NODE_ID")
if not node_id:
    console.log("Required environment variable CDNV3_NODE_ID is not defined")
    exit(1)

ws = create_connection("ws://localhost:4000/ws")

console.log("Sending connection request")
ws.send(json.dumps({
    "id": node_id
}))

console.log("Waiting for connection response")
try:
    connection_response = json.loads(ws.recv_frame().data)
except Exception as e:
    console.log(f"Received invalid JSON connection response payload: {e}")
    exit(1)
else:
    if not connection_response.get("permitted"):
        console.log(f"Connection request denied: {connection_response.get('message')}")
    else:
        console.log(f"Connection request accepted")
