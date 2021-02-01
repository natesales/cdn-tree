import ipaddress
import re

from pydantic import BaseModel, validator


def label_validator(label: str) -> str:
    """
    Validate a DNS zone label
    :param label: DNS zone label
    :return: label if validation success
    """
    if not label \
            and (re.match(r"^(?![0-9]+$)(?!-)[a-zA-Z0-9-]{,63}(?<!-)$", label) is None) \
            and (not label.startswith(".")) \
            and (not label.strip().startswith(" ")) \
            and (" " not in label):
        raise ValueError("invalid record label")
    return label


class Zone(BaseModel):
    """
    Zone stores a DNS zone add request
    """
    zone: str  # DNS zone

    @validator("zone")
    def zone_validator(cls, v):
        if not (v and (re.match(r"^(((?!-))(xn--|_)?[a-z0-9-]{0,61}[a-z0-9]\.)*(xn--)?([a-z0-9][a-z0-9\-]{0,60}|[a-z0-9-]{1,30}\.[a-z]{2,})$", v) is not None) and (v != ".") and (" " not in v) and (not v.startswith("-")) and (not v.endswith("-"))):
            raise ValueError("invalid DNS zone")
        return v


class ARecord(BaseModel):
    """
    ARecord stores a DNS record of type A
    """
    label: str
    ttl: int
    value: ipaddress.IPv4Address

    @validator("label")
    def label_validator(cls, v):
        return label_validator(v)

    def __str__(self) -> str:
        return f"{self.value}"


class AAAARecord(BaseModel):
    """
    AAAARecord stores a DNS record of type AAAA
    """
    label: str
    ttl: int
    value: ipaddress.IPv6Address

    @validator("label")
    def label_validator(cls, v):
        return label_validator(v)

    def __str__(self) -> str:
        return f"{self.value}"


class MXRecord(BaseModel):
    """
    MXRecord stores a DNS record of type MX
    """
    label: str
    ttl: int
    priority: int
    host: str

    @validator("label")
    def label_validator(cls, v):
        return label_validator(v)

    def __str__(self) -> str:
        return f"{self.value}"
