from os import environ
from time import time

import requests
from fastapi import FastAPI, Response, status
from pymongo import MongoClient, ASCENDING
from pymongo.errors import DuplicateKeyError
from rich.console import Console

from models.auth import *
from models.dns import *
from models.eca import *

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

# Create unique indices
db["zones"].create_index([("zone", ASCENDING)], unique=True)
db["users"].create_index([("email", ASCENDING)], unique=True)

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


# User authentication

@app.post("/auth/signup")
async def signup(user: User, response: Response):
    try:
        _user = user.dict()

        # Set user default values
        _user["enabled"] = False
        _user["admin"] = False
        db["users"].insert_one(_user)
    except DuplicateKeyError:
        response.status_code = status.HTTP_500_INTERNAL_SERVER_ERROR
        return {"detail": "User already exists"}
    else:
        return {"detail": "Signup success"}


# ECA routes

@app.post("/ecas/new")
async def new_eca(eca: ECA):
    eca_dict = eca.dict()
    eca_dict["authorized"] = False
    _id = db["ecas"].insert_one(eca_dict).inserted_id
    return {"id": str(_id)}


# DNS record management

def _add_record(zone: str, record: dict):
    return db["zones"].update_one(
        {"zone": zone}, {
            "$push": {"records": record},
            "$set": {"serial": str(int(time()))}
        }
    )


@app.post("/zones/add")
async def add_zone(zone: Zone, response: Response):
    _zone = zone.dict()
    # TODO: _zone["users"] = [authenticated_user]

    _resp = requests.get(CRYPTOD_URI + "/dnssec/newkey")
    if _resp.status_code != 200:
        console.log(f"cryptod dnssec/newkey error: HTTP {_resp.status_code}")
        response.status_code = status.HTTP_500_INTERNAL_SERVER_ERROR
        return {"detail": "unable to connect to cryptography service"}

    _zone["dnssec"] = _resp.json()

    try:
        db["zones"].insert_one(_zone)
    except DuplicateKeyError:
        response.status_code = status.HTTP_400_BAD_REQUEST
        return {"detail": "Zone already exists"}

    return {"success": True}


@app.delete("/zones/{zone}")
async def delete_zone(zone: str, response: Response):
    mod = db["zones"].delete_one({"zone": zone})
    if mod.deleted_count > 0:
        return {"detail": "Zone deleted successfully"}
    else:
        response.status_code = status.HTTP_400_BAD_REQUEST
        return {"detail": "Zone doesn't exist"}


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
