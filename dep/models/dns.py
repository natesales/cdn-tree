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


def ttl_validator(ttl: int) -> int:
    """
    Validate a DNS ttl
    :param ttl: DNS TTL
    :return: TTL if validation success
    """

    if not (ttl >= 30) and (ttl < 2147483646):
        raise ValueError("TTL out of bounds")
    return ttl


def fqdn_validator(fqdn: str) -> str:
    """
    Validates a FQDN or IP address
    :param fqdn: FQDN
    :return: FQDN if validation success
    """

    # Check if FQDN has trailing dot or is an IP address
    try:
        ipaddress.ip_address(fqdn)
    except ValueError:  # Not an IP address
        if not (fqdn.endswith(".")):
            raise ValueError("Invalid FQDN or IP address")

    return fqdn


def safe_dict(d: dict, rr_type: str) -> dict:
    """
    Cast unsafe types to strings
    :param d: dict of potentially BSON-unsafe kv pairs
    :param rr_type: DNS record type
    :return: dict with casted safe types
    """
    for k in d:
        if isinstance(d[k], ipaddress.IPv4Address) or isinstance(d[k], ipaddress.IPv6Address):
            d[k] = str(d[k])
    d["type"] = rr_type
    return d


class Zone(BaseModel):
    """
    Zone stores a DNS zone add request
    """
    zone: str  # DNS zone

    @validator("zone")
    def zone_validator(v):
        if not (v and (re.match(r"^(((?!-))(xn--|_)?[a-z0-9-]{0,61}[a-z0-9]\.)*(xn--)?([a-z0-9][a-z0-9\-]{0,60}|[a-z0-9-]{1,30}\.[a-z]{2,})$", v) is not None) and (v != ".") and (" " not in v) and (not v.startswith("-")) and (not v.endswith("-"))):
            raise ValueError("invalid DNS zone")
        return v


# DNS Records

class ARecord(BaseModel):
    """
    ARecord stores a DNS record of type A
    """
    label: str
    ttl: int
    value: ipaddress.IPv4Address

    @validator("label")
    def label_validator(v):
        return label_validator(v)

    @validator("ttl")
    def ttl_validator(v):
        return ttl_validator(v)

    def marshal(self) -> dict:
        return safe_dict(self.dict(), "A")


class AAAARecord(BaseModel):
    """
    AAAARecord stores a DNS record of type AAAA
    """
    label: str
    ttl: int
    value: ipaddress.IPv6Address

    @validator("label")
    def label_validator(v):
        return label_validator(v)

    @validator("ttl")
    def ttl_validator(v):
        return ttl_validator(v)

    def marshal(self) -> dict:
        return safe_dict(self.dict(), "AAAA")


class MXRecord(BaseModel):
    """
    MXRecord stores a DNS record of type MX
    """
    label: str
    ttl: int
    priority: int
    host: str

    @validator("label")
    def label_validator(v):
        return label_validator(v)

    @validator("ttl")
    def ttl_validator(v):
        return ttl_validator(v)

    @validator("host")
    def fqdn_validator(v):
        return fqdn_validator(v)

    def marshal(self) -> dict:
        return safe_dict(self.dict(), "MX")


class TXTRecord(BaseModel):
    """
    TXTRecord stores a DNS record of type TXT
    """
    label: str
    ttl: int
    value: str

    @validator("label")
    def label_validator(v):
        return label_validator(v)

    @validator("ttl")
    def ttl_validator(v):
        return ttl_validator(v)

    def marshal(self) -> dict:
        return safe_dict(self.dict(), "TXT")


class NSRecord(BaseModel):
    """
    NSRecord stores a DNS record of type NS
    """
    label: str
    ttl: int
    value: str

    @validator("label")
    def label_validator(v):
        return label_validator(v)

    @validator("value")
    def fqdn_validator(v):
        return fqdn_validator(v)

    def marshal(self) -> dict:
        return safe_dict(self.dict(), "NS")


class CNAMERecord(BaseModel):
    """
    CNAMERecord stores a DNS record of type CNAME
    """
    label: str
    ttl: int
    value: str

    @validator("label")
    def label_validator(v):
        return label_validator(v)

    @validator("value")
    def fqdn_validator(v):
        return fqdn_validator(v)

    def marshal(self) -> dict:
        return safe_dict(self.dict(), "CNAME")
