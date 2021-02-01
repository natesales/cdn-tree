import re


def valid_zone(zone: str) -> bool:
    # Validates a DNS zone (example.com)
    return zone \
           and (re.match(r"^(((?!-))(xn--|_)?[a-z0-9-]{0,61}[a-z0-9]\.)*(xn--)?([a-z0-9][a-z0-9\-]{0,60}|[a-z0-9-]{1,30}\.[a-z]{2,})$", zone) is not None) \
           and (zone != ".") \
           and (" " not in zone) \
           and (not zone.startswith("-")) \
           and (not zone.endswith("-"))


def valid_label(label: str) -> bool:
    # Validates a DNS zone label (www, @, example.com.)
    return label \
           and (re.match(r"^(?![0-9]+$)(?!-)[a-zA-Z0-9-]{,63}(?<!-)$", label) is None) \
           and (not label.startswith(".")) \
           and (not label.strip().startswith(" ")) \
           and (" " not in label)


def valid_ipv4(ipv4: str) -> bool:
    # Validates an IPv4 address (192.0.2.1)
    return ipv4 and (re.match(r"^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$", ipv4) is not None)


def valid_ipv6(ipv6: str) -> bool:
    # Validates an IPv4 address (2001:db8::1)
    return ipv6 \
           and (re.match(r"(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]+|::(ffff(:0{1,4})?:)?((25[0-5]|(2[0-4]|1?[0-9])?[0-9])\.){3}(25[0-5]|(2[0-4]|1?[0-9])?[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1?[0-9])?[0-9])\.){3}(25[0-5]|(2[0-4]|1?[0-9])?[0-9]))", ipv6) is not None)
