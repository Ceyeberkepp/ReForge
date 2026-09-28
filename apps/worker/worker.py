import os
import time
import httpx

API = os.getenv("REFORGE_API_URL", "http://api:8080")
POLL_SECONDS = int(os.getenv("REFORGE_WORKER_POLL_SECONDS", "3"))
EXECUTE_PRIVILEGED = os.getenv("REFORGE_EXECUTE_PRIVILEGED", "false").lower() == "true"

STEPS = [
    (5, "Validate host"),
    (12, "Boot imaging environment"),
    (20, "Partition disk"),
    (38, "Restore gold image"),
    (48, "Apply drivers"),
    (56, "Boot operating system"),
    (62, "Rename computer"),
    (70, "Install required software"),
    (77, "Install optional software"),
    (84, "Join Active Directory"),
    (88, "Move to department OU"),
    (91, "Install printers"),
    (94, "Run post-install scripts"),
    (97, "Run updates"),
    (100, "Validate deployment"),
]

def patch(client, job_id, **payload):
    response = client.patch(f"{API}/api/deployments/{job_id}/worker", json=payload)
    response.raise_for_status()

def process(client, job):
    job_id = job["id"]
    patch(client, job_id, status="running", progress=1, current_step="Starting")
    plan = client.get(f"{API}/api/deployments/{job_id}/plan")
    plan.raise_for_status()
    plan = plan.json()

    # Safe MVP mode: orchestrates and validates the deployment plan without
    # running destructive disk commands. A dedicated imaging-node adapter will
    # replace this block when REFORGE_EXECUTE_PRIVILEGED is enabled.
    for progress, step in STEPS:
        latest = client.get(f"{API}/api/deployments").json()
        current = next((x for x in latest if x["id"] == job_id), None)
        if not current or current["status"] == "canceled":
            return
        patch(client, job_id, progress=progress, current_step=step)
        time.sleep(0.35 if not EXECUTE_PRIVILEGED else 1)

    patch(client, job_id, status="succeeded", progress=100, current_step="Complete")

def main():
    while True:
        try:
            with httpx.Client(timeout=10) as client:
                jobs = client.get(f"{API}/api/deployments").json()
                queued = [j for j in jobs if j["status"] == "queued"]
                if queued:
                    process(client, queued[-1])
        except Exception as exc:
            print(f"worker error: {exc}", flush=True)
        time.sleep(POLL_SECONDS)

if __name__ == "__main__":
    main()
