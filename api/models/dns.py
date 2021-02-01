import ipaddress

from pydantic import BaseModel


class ARecord(BaseModel):
    """
    ARecord stores a DNS record of type A
    """
    label: str
    ttl: int
    value: ipaddress.IPv4Address

    def __str__(self) -> str:
        return f"{self.value}"


class AAAARecord(BaseModel):
    """
    AAAARecord stores a DNS record of type AAAA
    """
    label: str
    ttl: int
    value: ipaddress.IPv6Address

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

    def __str__(self) -> str:
        return f"{self.priority} {self.host}"
