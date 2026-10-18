"""Save a workload and its pods, as the API returns them, for KEL's tests.

hack/live/capture.py OUT_DIR NAME KIND/WORKLOAD SELECTOR [CHECK]
CHECK (optional) is a Python expression over `cs` (first container status) and
`ics` (first init container status, or {}); the script polls until it holds.
"""
import json, subprocess, sys, time

out_dir, name, workload, selector = sys.argv[1:5]
check = sys.argv[5] if len(sys.argv) > 5 else "True"

def get(*args):
    return json.loads(subprocess.check_output(["kubectl", "-n", "kel-lab", "get", *args, "-o", "json"]))

def strip(o):
    o.get("metadata", {}).pop("managedFields", None)
    for it in o.get("items", []):
        strip(it)
    return o

deadline = time.time() + 600
while True:
    pods = get("pods", "-l", selector)
    p = pods["items"][0] if pods["items"] else {}
    cs = (p.get("status", {}).get("containerStatuses") or [{}])[0]
    ics = (p.get("status", {}).get("initContainerStatuses") or [{}])[0]
    if pods["items"] and eval(check, {}, {"cs": cs, "ics": ics}):
        break
    if time.time() > deadline:
        sys.exit("timed out waiting for " + name)
    time.sleep(0.5)

server = json.loads(subprocess.check_output(["kubectl", "version", "-o", "json"]))["serverVersion"]["gitVersion"]
snap = {"kubernetes": server, "workload": strip(get(workload)), "pods": strip(pods)}
with open(f"{out_dir}/{name}.json", "w") as f:
    json.dump(snap, f, indent=1, sort_keys=True)
print("saved", name)
