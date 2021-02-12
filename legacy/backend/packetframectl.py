#!/usr/bin/python3
# CLI utility for the PacketFrame CDN

import os
import sys
import requests
from terminaltables import SingleTable

help_text = """Usage: packetframectl command category [args]

Commands:
  get    Retrieve data
  
Categories:
  zones         Get list of zones
  records zone  Get records for zone
  acl           Get user ACL
  users zone    Get users for zone
"""

API_KEY = os.environ.get("PACKETFRAME_API_KEY")
if not API_KEY:
    print("PACKETFRAME_API_KEY environment variable not found.")
    exit(1)


def _get(endpoint, body):
    return requests.get("https://packetframe.com/api/" + endpoint, headers={"X-API-Key": API_KEY}, json=body)


def _truncate(string):
    return string if len(string) < 50 else string[:50] + "..."


if len(sys.argv) == 1:
    print(help_text)
    exit(1)

if sys.argv[1] == "get":
    if sys.argv[2] == "zones":
        table = [("\033[4mZone\033[0m", "\033[4mRecords\033[0m", "\033[4mUsers\033[0m")]
        for zone in _get("zones/list", None).json()["message"]:
            table.append((zone["zone"], len(zone["records"]), len(zone["users"])))
        print(SingleTable(table).table)

    elif sys.argv[2] == "records":
        zone = sys.argv[3]
        table = [("\033[4mLabel\033[0m", "\033[4mTTL\033[0m", "\033[4mProxied\033[0m", "\033[4mValue\033[0m")]
        for record in _get("zone/" + zone + "/records", None).json()["message"]:
            table.append((_truncate(record["label"]), record["ttl"], "✓" if record.get("proxied") else "x", _truncate(record["value"])))
        print(SingleTable(table).table)

    elif sys.argv[2] == "acl":
        print("ACL:")
        for address in _get("user/acl", None).json()["message"]:
            print("- " + address)

    elif sys.argv[2] == "users":
        zone = sys.argv[3]
        print(f"Users for {zone}")
        for user in _get("zone/" + zone + "/users", None).json()["message"]:
            print("- " + user)

else:
    print(help_text)
