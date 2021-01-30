import ipaddress

from pydantic import BaseModel


class Zone(BaseModel):
    label: str
    ttl: int
    value: ipaddress.IPv6Address

    def __str__(self) -> str:
        return f"{self.value}"
