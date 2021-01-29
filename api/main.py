from os import environ

from bson import ObjectId
from bson.errors import InvalidId
from fastapi import FastAPI, WebSocket
from pymongo import MongoClient
from rich.console import Console

import models

if environ.get("CDNV3_DEVELOPMENT"):
    DEVELOPMENT = True
else:
    DEVELOPMENT = False

# Rich console
console = Console()

app = FastAPI(title="Packetframe Control Plane", description="Control plane for the Packetframe CDN", version="3.0.0")

console.log("Connecting to MongoDB")
if DEVELOPMENT:
    console.log("Using local development database")
    db = MongoClient("localhost:27017")["cdnv3db"]
else:  # Production replicaset
    db = MongoClient("localhost:27017", replicaSet="cdnv3")["cdnv3db"]
console.log("Connected to MongoDB")


@app.post("/ecas/new")
async def new_eca(eca: models.ECA):
    _id = db["ecas"].insert_one({
        "provider": eca.provider,
        "latitude": eca.latitude,
        "longitude": eca.longitude,
        "roles": eca.roles,
        "authorized": False
    }).inserted_id

    return {"id": str(_id)}


@app.websocket("/ws")
async def websocket_stream(websocket: WebSocket):
    await websocket.accept()
    console.log(f"Opened websocket from {websocket.client.host} {websocket.client_state}")
    connection_request = await websocket.receive_json()

    try:
        object_id = ObjectId(connection_request.get("id"))
    except InvalidId:
        console.log(f"Rejecting invalid ObjectId: {connection_request}")
        await websocket.send_json({"permitted": False, "message": "Invalid node ID"})
        return

    eca_doc = db["ecas"].find_one({
        "_id": object_id
    })

    if not eca_doc.get("authorized"):
        await websocket.send_json({"permitted": False, "message": "Not authorized"})

    # By this point, the node is authorized
    await websocket.send_json({"permitted": True, "message": "Accepted connection request"})
