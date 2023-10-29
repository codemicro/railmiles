#!/usr/bin/env python3

import requests
import json
import sys
import re
from pypdf import PdfReader


NRT_STATION_INDEX_FILENAME = sys.argv[1]


def case_insensitive_rstrip(inp, snip):
    snip = snip.lower()
    if inp.lower().endswith(snip):
        return inp[:-len(snip)].strip()
    return inp


def is_extraneous_code(code):
    dprs_codes = {
        # Source: http://www.railwaycodes.org.uk/crs/crs2.shtm
        "EBF": "EBD", # Ebbsfleet International
        "GCL": "GLC", # Glasgow Central        
        "GQL": "GLQ", # Glasgow Queen Street   
        "HEZ": "HEW", # Heworth                
        "HII": "HHY", # Highbury & Islington   
        "XHZ": "HHY",
        "LIF": "LTV", # Lichfield Trent Valley 
        "LVL": "LIV", # Liverpool Lime Street  
        "ALE": "LPY", # Liverpool South Parkway
        "SPL": "STP", # London St Pancras      
        "SPX": "STP",
        "XRO": "RET", # Retford                
        "GTI": "SGB", # Smethwick Galton Bridge
        "TAH": "TAM", # Tamworth               
        "WJH": "WIJ", # Willesden Junction     
        "WJL": "WIJ",
        "WPH": "WOP", # Worcestershire Parkway 
    }

    return code.upper() in dprs_codes


def get_names_from_pdf(filename):
    reader = PdfReader(filename)


    def extract_text_from_page(page):
        items = []

        def visitor_body(text, cm, tm, font_dict, font_size):
            x, y = tm[4], tm[5]
            if text == "" or (x == 0 and y == 0):
                return

            items.append((text, x, y))

        page.extract_text(visitor_text=visitor_body)

        return items


    stations = {}

    print("Processing station index PDF", file=sys.stderr)

    for page_number in range(len(reader.pages)):
        items = extract_text_from_page(reader.pages[page_number])    

        i = 0
        while i < len(items):
            x = items[i][0].strip()

            try:
                next_text = items[i+1][0]
            except IndexError:
                next_text = ""
            if re.match("[A-Z]{3}", x) and next_text != "\n" and "(continued)" not in next_text:
                stations[x] = items[i+1][0].strip()

            i += 1
        
    return stations


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
    if crs is None or is_extraneous_code(crs):
        continue

    name = tags.get("name", "")
    name = case_insensitive_rstrip(name, "high level")
    name = case_insensitive_rstrip(name, "low level")

    out[crs] = {
        "lat": elem.get("lat", 0),
        "lon": elem.get("lon", 0),
        "name": name,
    }


nrt_names = get_names_from_pdf(NRT_STATION_INDEX_FILENAME)

print("Overwriting station names with NRT names", file=sys.stderr)

for crs in out:
    if crs in nrt_names:
        out[crs]["name"] = nrt_names[crs]

json.dump(out, open("stationData.json", "w"))
