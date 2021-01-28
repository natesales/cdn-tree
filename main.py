import logging
from enum import Enum
from typing import List

from fastapi import FastAPI
from pydantic import BaseModel
from pymongo import MongoClient
from rich.logging import RichHandler

# Init rich logging handler
logging.basicConfig(level="NOTSET", format="%(message)s", datefmt="[%X]", handlers=[RichHandler()])
log = logging.getLogger("rich")

app = FastAPI()

log.info("Connecting to MongoDB")
db = MongoClient("localhost:27017", replicaSet="cdnv3")["cdnv3db"]
log.info("Connected to MongoDB")


# Request structures

class ECAState(str, Enum):
    """
    ECAState defines temporary states that an ECA may be in
    """
    Pending = "pending"  # Not yet connected to control plane
    Established = "established"  # Connected and expected to serve traffic
    Faulted = "faulted"  # Something is wrong with the node
    Draining = "draining"  # ECA has been queued for decommissioning


class ECARole(str, Enum):
    """
    ECARole defines roles that an ECA might serve
    """
    DNS = "dns"
    HTTPCache = "httpcache"


class ECA(BaseModel):
    """
    ECA defines user-supplied parameters of an ECA
    """
    provider: str
    latitude: float
    longitude: float
    roles: List[ECARole]


@app.post("/ecas/new")
async def ecas_new(eca: ECA):
    return db["ecas"].insert_one({
        "provider": eca.provider,
        "latitude": eca.latitude,
        "longitude": eca.longitude,
        "roles": eca.roles
    }).inserted_id
