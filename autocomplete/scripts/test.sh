
#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT_DIR"

BASE_URL="${BASE_URL:-http://localhost:8080}"
export BASE_URL

python3 <<'PY'
import json
import os
import urllib.parse
import urllib.error
import urllib.request

BASE_URL = os.environ["BASE_URL"]
passed = 0
failed = 0

def request(params):
    url = BASE_URL + "/v1/autocomplete"
    if params:
        url += "?" + urllib.parse.urlencode(params)

    try:
        with urllib.request.urlopen(url, timeout=10) as response:
            return response.status, json.load(response)
    except urllib.error.HTTPError as e:
        return e.code, json.load(e)

def check(name, condition, details=""):
    global passed, failed
    if condition:
        passed += 1
        print(f"PASS: {name}")
    else:
        failed += 1
        print(f"FAIL: {name} {details}")

def search(prefix, limit=10):
    status, body = request({"q": prefix, "limit": limit})
    suggestions = body.get("suggestions", [])
    titles = [m["title"] for m in suggestions]

    print(f"\nQuery: {prefix!r}, limit={limit}")
    for title in titles:
        print(f"  - {title}")

    check(f"{prefix}: HTTP 200", status == 200)
    check(
        f"{prefix}: returned <= limit",
        len(suggestions) <= limit
    )
    check(
        f"{prefix}: prefix matching",
        all(t.lower().startswith(prefix.lower()) for t in titles)
    )
    check(
        f"{prefix}: popularity sorted",
        all(
            suggestions[i]["popularity"] >=
            suggestions[i+1]["popularity"]
            for i in range(len(suggestions)-1)
        )
    )
    check(
        f"{prefix}: response count",
        body.get("pagination", {}).get("returned")
        == len(suggestions)
    )
    return titles

# 1. Search common prefixes
for prefix in [
    "star",
    "the",
    "titan",
    "incept",
    "inter",
    "avatar",
    "harry",
    "jurassic",
    "matrix",
    "zzzznotfound",
]:
    titles = search(prefix)

    if prefix == "zzzznotfound":
        check("No-match returns empty list", titles == [])

# 2. Case-insensitive matching
lower = search("star")
upper = search("STAR")
check("Case-insensitive results", lower == upper)

# 3. Limit behavior
limited = search("star", 3)
check("Result limit respected", len(limited) <= 3)

# 4. Invalid requests
for name, params in [
    ("Empty query", {"q": ""}),
    ("Missing query", {}),
    ("Invalid limit 0", {"q": "star", "limit": "0"}),
    ("Invalid limit 100", {"q": "star", "limit": "100"}),
    ("Non-numeric limit", {"q": "star", "limit": "abc"}),
]:
    status, body = request(params)
    check(
        name,
        status == 400 and "error" in body,
        f"status={status}"
    )

print(f"\nResults: {passed} passed, {failed} failed")

if failed:
    raise SystemExit(1)
PY
