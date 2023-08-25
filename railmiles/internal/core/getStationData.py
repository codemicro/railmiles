#!/usr/bin/env python3

import requests
import json
import sys

url = "https://overpass-api.de/api/interpreter"

query = """[out:json][timeout:25];
(node["ref:crs"];);
out body;"""

print("Querying Overpass", file=sys.stderr)

r = requests.post(url, data={"data": query})
r.raise_for_status()

rj = r.json()

out = {}

print("Processing results", file=sys.stderr)

for elem in rj.get("elements", []):
    tags = elem.get("tags", {})

    crs = tags.get("ref:crs")
    if crs is None:
        continue

    out[crs] = {
        "lat": elem.get("lat", 0),
        "lon": elem.get("lon", 0),
        "name": tags.get("name", ""),
    }

json.dump(out, open("stationData.json", "w"))
