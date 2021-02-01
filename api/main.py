from os import environ
from time import time

import requests
from bson import ObjectId
from bson.errors import InvalidId
from fastapi import FastAPI, WebSocket, Response, status
from pymongo import MongoClient, ASCENDING
from pymongo.errors import DuplicateKeyError
from rich.console import Console

from models import eca as eca_models
from models.dns import *

MONGODB_URI = "mongodb://localhost:27017"
CRYPTOD_URI = "http://localhost:8081"

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
    db = MongoClient(MONGODB_URI)["cdnv3db"]
else:  # Production replicaset
    db = MongoClient(MONGODB_URI, replicaSet="cdnv3")["cdnv3db"]
console.log("Connected to MongoDB")
db["zones"].create_index([("zone", ASCENDING)], unique=True)

console.log("Checking cryptod connection")
try:
    resp = requests.get(CRYPTOD_URI + "/healthcheck")
except requests.exceptions.ConnectionError:
    console.log("Unable to connect to cryptod")
    exit(1)
else:
    if resp.status_code != 200:
        console.log(f"cryptod connection error: HTTP {resp.status_code}")
        exit(1)
    else:
        console.log("cryptod ok")


@app.post("/ecas/new")
async def new_eca(eca: eca_models.ECA):
    eca_dict = eca.dict()
    eca_dict["authorized"] = False
    _id = db["ecas"].insert_one(eca_dict).inserted_id
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

    while True:
        connection_request = await websocket.receive_json()


# DNS record management

def _add_record(zone: str, record: dict):
    return db["zones"].update_one(
        {"zone": zone}, {
            "$push": {"records": record},
            "$set": {"serial": str(int(time()))}
        }
    )


@app.post("/zones/add")
def add_zone(zone: Zone, response: Response):
    _zone = zone.dict()
    # TODO: _zone["users"] = [authenticated_user]

    resp = requests.get(CRYPTOD_URI + "/dnssec/newkey")
    if resp.status_code != 200:
        console.log(f"cryptod dnssec/newkey error: HTTP {resp.status_code}")
        response.status_code = status.HTTP_500_INTERNAL_SERVER_ERROR
        return {"detail": "unable to connect to cryptography service"}

    _zone["dnssec"] = resp.json()

    try:
        db["zones"].insert_one(_zone)
    except DuplicateKeyError:
        response.status_code = status.HTTP_400_BAD_REQUEST
        return {"detail": "Zone already exists"}

    return {"success": True}


# Record adders

@app.post("/records/{zone}/add/A")
async def add_a_record(zone: str, record: ARecord, response: Response):
    result = _add_record(zone, record.marshal())
    if not result.modified_count:
        response.status_code = status.HTTP_400_BAD_REQUEST
        return {"detail": "Zone doesn't exist"}
    return {"detail": "Record added successfully"}


@app.post("/records/{zone}/add/AAAA")
async def add_aaaa_record(zone: str, record: AAAARecord, response: Response):
    result = _add_record(zone, record.marshal())
    if not result.modified_count:
        response.status_code = status.HTTP_400_BAD_REQUEST
        return {"detail": "Zone doesn't exist"}
    return {"detail": "Record added successfully"}


@app.post("/records/{zone}/add/MX")
async def add_mx_record(zone: str, record: MXRecord, response: Response):
    result = _add_record(zone, record.marshal())
    if not result.modified_count:
        response.status_code = status.HTTP_400_BAD_REQUEST
        return {"detail": "Zone doesn't exist"}
    return {"detail": "Record added successfully"}


@app.post("/records/{zone}/add/TXT")
async def add_txt_record(zone: str, record: TXTRecord, response: Response):
    result = _add_record(zone, record.marshal())
    if not result.modified_count:
        response.status_code = status.HTTP_400_BAD_REQUEST
        return {"detail": "Zone doesn't exist"}
    return {"detail": "Record added successfully"}


@app.post("/records/{zone}/add/NS")
async def add_ns_record(zone: str, record: NSRecord, response: Response):
    result = _add_record(zone, record.marshal())
    if not result.modified_count:
        response.status_code = status.HTTP_400_BAD_REQUEST
        return {"detail": "Zone doesn't exist"}
    return {"detail": "Record added successfully"}


@app.post("/records/{zone}/add/CNAME")
async def add_cname_record(zone: str, record: CNAMERecord, response: Response):
    result = _add_record(zone, record.marshal())
    if not result.modified_count:
        response.status_code = status.HTTP_400_BAD_REQUEST
        return {"detail": "Zone doesn't exist"}
    return {"detail": "Record added successfully"}
