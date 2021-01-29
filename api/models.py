from enum import Enum
from typing import List

from pydantic import BaseModel


# Request structures

class ECAState(str, Enum):
    """
    ECAState defines temporary states that an ECA may be in
    """
    pending = "pending"  # Not yet connected to control plane
    established = "established"  # Connected and expected to serve traffic
    faulted = "faulted"  # Something is wrong with the node
    draining = "draining"  # ECA has been queued for decommissioning


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
