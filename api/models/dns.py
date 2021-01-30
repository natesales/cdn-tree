import ipaddress

from pydantic import BaseModel


class ARecord(BaseModel):
    label: str
    ttl: int
    value: ipaddress.IPv4Address

    def __str__(self) -> str:
        return f"{self.value}"


class AAAARecord(BaseModel):
    label: str
    ttl: int
    value: ipaddress.IPv6Address

    def __str__(self) -> str:
        return f"{self.value}"


class MXRecord(BaseModel):
    label: str
    ttl: int
    priority: int
    host: str

    def __str__(self) -> str:
        return f"{self.priority} {self.host}"
