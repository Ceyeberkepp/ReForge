from datetime import datetime
from fastapi import Depends, FastAPI, HTTPException, Response
from fastapi.middleware.cors import CORSMiddleware
from ldap3 import Connection, Server, Tls
from sqlalchemy.orm import Session
import re
import ssl
from .database import Base, engine, get_db
from . import models, schemas

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="ReForge API",
    version="0.2.0",
    description="Control plane for ReForge endpoint imaging and deployment.",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

def normalize_mac(value: str) -> str:
    raw = re.sub(r"[^0-9a-fA-F]", "", value).lower()
    if len(raw) != 12:
        raise HTTPException(400, "Invalid MAC address")
    return ":".join(raw[i:i+2] for i in range(0, 12, 2))

def apply_payload(item, payload):
    for key, value in payload.model_dump().items():
        setattr(item, key, value)
    return item

@app.get("/health")
def health():
    return {"status": "ok", "service": "reforge-api", "version": "0.2.0"}

@app.get("/api/dashboard")
def dashboard(db: Session = Depends(get_db)):
    return {
        "images": db.query(models.GoldImage).count(),
        "departments": db.query(models.Department).count(),
        "software": db.query(models.SoftwarePackage).count(),
        "hosts": db.query(models.Host).count(),
        "queued_deployments": db.query(models.DeploymentJob).filter(
            models.DeploymentJob.status == models.DeploymentStatus.queued
        ).count(),
        "running_deployments": db.query(models.DeploymentJob).filter(
            models.DeploymentJob.status == models.DeploymentStatus.running
        ).count(),
    }

@app.get("/api/images", response_model=list[schemas.GoldImageOut])
def list_images(db: Session = Depends(get_db)):
    return db.query(models.GoldImage).order_by(models.GoldImage.created_at.desc()).all()

@app.post("/api/images", response_model=schemas.GoldImageOut)
def create_image(payload: schemas.GoldImageCreate, db: Session = Depends(get_db)):
    item = models.GoldImage(**payload.model_dump())
    db.add(item); db.commit(); db.refresh(item)
    return item

@app.put("/api/images/{item_id}", response_model=schemas.GoldImageOut)
def update_image(item_id: str, payload: schemas.GoldImageCreate, db: Session = Depends(get_db)):
    item = db.get(models.GoldImage, item_id)
    if not item: raise HTTPException(404, "Gold image not found")
    apply_payload(item, payload); db.commit(); db.refresh(item)
    return item

@app.delete("/api/images/{item_id}", status_code=204)
def delete_image(item_id: str, db: Session = Depends(get_db)):
    item = db.get(models.GoldImage, item_id)
    if not item: raise HTTPException(404, "Gold image not found")
    db.delete(item); db.commit()
    return Response(status_code=204)

@app.get("/api/departments", response_model=list[schemas.DepartmentOut])
def list_departments(db: Session = Depends(get_db)):
    return db.query(models.Department).order_by(models.Department.name).all()

@app.post("/api/departments", response_model=schemas.DepartmentOut)
def create_department(payload: schemas.DepartmentCreate, db: Session = Depends(get_db)):
    item = models.Department(**payload.model_dump())
    db.add(item); db.commit(); db.refresh(item)
    return item

@app.put("/api/departments/{item_id}", response_model=schemas.DepartmentOut)
def update_department(item_id: str, payload: schemas.DepartmentCreate, db: Session = Depends(get_db)):
    item = db.get(models.Department, item_id)
    if not item: raise HTTPException(404, "Department not found")
    apply_payload(item, payload); db.commit(); db.refresh(item)
    return item

@app.delete("/api/departments/{item_id}", status_code=204)
def delete_department(item_id: str, db: Session = Depends(get_db)):
    item = db.get(models.Department, item_id)
    if not item: raise HTTPException(404, "Department not found")
    db.delete(item); db.commit()
    return Response(status_code=204)

@app.get("/api/software", response_model=list[schemas.SoftwareOut])
def list_software(db: Session = Depends(get_db)):
    return db.query(models.SoftwarePackage).order_by(
        models.SoftwarePackage.install_order, models.SoftwarePackage.name
    ).all()

@app.post("/api/software", response_model=schemas.SoftwareOut)
def create_software(payload: schemas.SoftwareCreate, db: Session = Depends(get_db)):
    item = models.SoftwarePackage(**payload.model_dump())
    db.add(item); db.commit(); db.refresh(item)
    return item

@app.put("/api/software/{item_id}", response_model=schemas.SoftwareOut)
def update_software(item_id: str, payload: schemas.SoftwareCreate, db: Session = Depends(get_db)):
    item = db.get(models.SoftwarePackage, item_id)
    if not item: raise HTTPException(404, "Software package not found")
    apply_payload(item, payload); db.commit(); db.refresh(item)
    return item

@app.delete("/api/software/{item_id}", status_code=204)
def delete_software(item_id: str, db: Session = Depends(get_db)):
    item = db.get(models.SoftwarePackage, item_id)
    if not item: raise HTTPException(404, "Software package not found")
    db.delete(item); db.commit()
    return Response(status_code=204)

@app.post("/api/software/bootstrap")
def bootstrap_software(db: Session = Depends(get_db)):
    starter = [
        ("Adobe Acrobat Reader", "AcroRdrDCx64.exe /sAll /rs /rps", 40),
        ("Google Chrome", "msiexec /i googlechromestandaloneenterprise64.msi /qn /norestart", 30),
        ("Mozilla Firefox", "FirefoxSetup.exe /S", 50),
        ("Microsoft 365 Apps", "setup.exe /configure configuration.xml", 20),
        ("Microsoft Teams", "teamsbootstrapper.exe -p", 35),
        ("7-Zip", "7z-x64.exe /S", 60),
        ("VLC", "vlc-win64.exe /S", 65),
        ("Notepad++", "npp-installer.exe /S", 70),
        ("PuTTY", "msiexec /i putty.msi /qn /norestart", 70),
        ("WinSCP", "winscp.exe /VERYSILENT /NORESTART", 70),
        ("Visual Studio Code", "VSCodeSetup.exe /VERYSILENT /NORESTART /MERGETASKS=!runcode", 70),
        ("Power BI Desktop", "PBIDesktopSetup_x64.exe -quiet -norestart", 70),
        ("Zoom Workplace", "msiexec /i ZoomInstallerFull.msi /qn /norestart", 70),
        ("GlobalProtect", "msiexec /i GlobalProtect64.msi /qn /norestart", 80),
        ("NinjaOne Agent", "msiexec /i NinjaOneAgent.msi /qn /norestart", 90),
    ]
    created = 0
    for name, command, order in starter:
        if db.query(models.SoftwarePackage).filter(models.SoftwarePackage.name == name).first():
            continue
        db.add(models.SoftwarePackage(name=name, silent_install=command, install_order=order))
        created += 1
    db.commit()
    return {"created": created, "message": "Starter software catalog loaded"}

@app.get("/api/hosts", response_model=list[schemas.HostOut])
def list_hosts(db: Session = Depends(get_db)):
    return db.query(models.Host).order_by(models.Host.hostname).all()

@app.post("/api/hosts", response_model=schemas.HostOut)
def create_host(payload: schemas.HostCreate, db: Session = Depends(get_db)):
    data = payload.model_dump()
    data["mac_address"] = normalize_mac(data["mac_address"])
    item = models.Host(**data)
    db.add(item); db.commit(); db.refresh(item)
    return item

@app.put("/api/hosts/{item_id}", response_model=schemas.HostOut)
def update_host(item_id: str, payload: schemas.HostCreate, db: Session = Depends(get_db)):
    item = db.get(models.Host, item_id)
    if not item: raise HTTPException(404, "Host not found")
    data = payload.model_dump()
    data["mac_address"] = normalize_mac(data["mac_address"])
    for key, value in data.items(): setattr(item, key, value)
    db.commit(); db.refresh(item)
    return item

@app.delete("/api/hosts/{item_id}", status_code=204)
def delete_host(item_id: str, db: Session = Depends(get_db)):
    item = db.get(models.Host, item_id)
    if not item: raise HTTPException(404, "Host not found")
    db.delete(item); db.commit()
    return Response(status_code=204)

@app.post("/api/hosts/register", response_model=schemas.HostOut)
def register_host(payload: schemas.HostRegistration, db: Session = Depends(get_db)):
    mac = normalize_mac(payload.mac_address)
    item = db.query(models.Host).filter(models.Host.mac_address == mac).first()
    now = datetime.utcnow()
    if item:
        item.last_seen = now
        if payload.serial_number: item.serial_number = payload.serial_number
        if payload.manufacturer: item.manufacturer = payload.manufacturer
        if payload.model: item.model = payload.model
    else:
        suffix = mac.replace(":", "")[-6:].upper()
        item = models.Host(
            hostname=payload.hostname or f"NEW-{suffix}",
            mac_address=mac,
            serial_number=payload.serial_number,
            manufacturer=payload.manufacturer,
            model=payload.model,
            last_seen=now,
        )
        db.add(item)
    db.commit(); db.refresh(item)
    return item

@app.get("/api/directory", response_model=schemas.DirectoryConfigOut)
def get_directory(db: Session = Depends(get_db)):
    item = db.get(models.DirectoryConfig, "default")
    if not item:
        item = models.DirectoryConfig(id="default")
        db.add(item); db.commit(); db.refresh(item)
    return item

@app.put("/api/directory", response_model=schemas.DirectoryConfigOut)
def update_directory(payload: schemas.DirectoryConfigIn, db: Session = Depends(get_db)):
    item = db.get(models.DirectoryConfig, "default")
    if not item:
        item = models.DirectoryConfig(id="default")
        db.add(item)
    apply_payload(item, payload); db.commit(); db.refresh(item)
    return item

@app.post("/api/directory/test")
def test_directory(payload: schemas.DirectoryTestRequest):
    use_ssl = payload.protocol.lower() == "ldaps"
    tls = Tls(validate=ssl.CERT_REQUIRED) if use_ssl else None
    try:
        server = Server(payload.domain_controller, port=payload.port, use_ssl=use_ssl, tls=tls, connect_timeout=5)
        conn = Connection(
            server, user=payload.service_account,
            password=payload.password.get_secret_value(),
            auto_bind=True, receive_timeout=5
        )
        conn.unbind()
        return {"ok": True, "message": "Directory bind succeeded", "server": payload.domain_controller}
    except Exception as exc:
        raise HTTPException(400, detail=f"Directory bind failed: {type(exc).__name__}: {exc}")

@app.get("/api/deployments", response_model=list[schemas.DeploymentOut])
def list_deployments(db: Session = Depends(get_db)):
    return db.query(models.DeploymentJob).order_by(models.DeploymentJob.created_at.desc()).all()

@app.post("/api/deployments", response_model=schemas.DeploymentOut)
def create_deployment(payload: schemas.DeploymentCreate, db: Session = Depends(get_db)):
    if not db.get(models.Host, payload.host_id): raise HTTPException(404, "Host not found")
    if not db.get(models.GoldImage, payload.image_id): raise HTTPException(404, "Gold image not found")
    if payload.department_id and not db.get(models.Department, payload.department_id):
        raise HTTPException(404, "Department not found")
    item = models.DeploymentJob(**payload.model_dump())
    db.add(item); db.commit(); db.refresh(item)
    return item

@app.post("/api/deployments/{item_id}/cancel", response_model=schemas.DeploymentOut)
def cancel_deployment(item_id: str, db: Session = Depends(get_db)):
    item = db.get(models.DeploymentJob, item_id)
    if not item: raise HTTPException(404, "Deployment not found")
    item.status = models.DeploymentStatus.canceled
    item.current_step = "Canceled"
    db.commit(); db.refresh(item)
    return item

@app.post("/api/deployments/{item_id}/retry", response_model=schemas.DeploymentOut)
def retry_deployment(item_id: str, db: Session = Depends(get_db)):
    item = db.get(models.DeploymentJob, item_id)
    if not item: raise HTTPException(404, "Deployment not found")
    item.status = models.DeploymentStatus.queued
    item.progress = 0
    item.current_step = "Queued"
    db.commit(); db.refresh(item)
    return item

@app.patch("/api/deployments/{item_id}/worker", response_model=schemas.DeploymentOut)
def worker_update(item_id: str, payload: schemas.WorkerUpdate, db: Session = Depends(get_db)):
    item = db.get(models.DeploymentJob, item_id)
    if not item: raise HTTPException(404, "Deployment not found")
    for key, value in payload.model_dump(exclude_none=True).items():
        setattr(item, key, value)
    db.commit(); db.refresh(item)
    return item

@app.get("/api/deployments/{item_id}/plan")
def deployment_plan(item_id: str, db: Session = Depends(get_db)):
    job = db.get(models.DeploymentJob, item_id)
    if not job: raise HTTPException(404, "Deployment not found")
    host = db.get(models.Host, job.host_id)
    image = db.get(models.GoldImage, job.image_id)
    dept = db.get(models.Department, job.department_id) if job.department_id else None
    ids = list(dept.required_software_ids or []) if dept else []
    ids.extend(job.optional_software or [])
    packages = []
    for software_id in dict.fromkeys(ids):
        pkg = db.get(models.SoftwarePackage, software_id)
        if pkg:
            packages.append({
                "id": pkg.id, "name": pkg.name, "version": pkg.version,
                "installer_path": pkg.installer_path, "silent_install": pkg.silent_install,
                "detection_rule": pkg.detection_rule, "install_order": pkg.install_order,
            })
    packages.sort(key=lambda x: x["install_order"])
    return {
        "job_id": job.id,
        "host": {"id": host.id, "hostname": host.hostname, "mac_address": host.mac_address},
        "image": {"id": image.id, "name": image.name, "path": image.image_path, "version": image.image_version},
        "department": None if not dept else {
            "id": dept.id, "name": dept.name, "code": dept.code,
            "ad_ou": dept.ad_ou, "computer_name_pattern": dept.computer_name_pattern,
            "printers": dept.printers, "post_install_scripts": dept.post_install_scripts,
        },
        "software": packages,
        "steps": [
            "Validate host", "Boot imaging environment", "Partition disk", "Restore gold image",
            "Apply drivers", "Boot operating system", "Rename computer",
            "Install required software", "Install optional software", "Join Active Directory",
            "Move to department OU", "Install printers", "Run post-install scripts",
            "Run updates", "Validate deployment"
        ],
    }

@app.get("/boot/ipxe", response_class=Response)
def ipxe_menu(api_url: str = "http://reforge.local:8080"):
    script = f"""#!ipxe
set reforge-api {api_url}
menu ReForge Network Deployment
item register Register / inventory this computer
item deploy Check for assigned deployment
item local Boot from local disk
choose --default deploy --timeout 5000 target && goto ${target}

:register
chain ${reforge-api}/boot/register.ipxe?mac=${net0/mac} || shell

:deploy
chain ${reforge-api}/boot/deploy.ipxe?mac=${net0/mac} || goto local

:local
exit
"""
    return Response(content=script, media_type="text/plain")

@app.get("/boot/register.ipxe", response_class=Response)
def ipxe_register(mac: str):
    normalized = normalize_mac(mac)
    return Response(content=f"""#!ipxe
echo ReForge registration
echo MAC: {normalized}
echo Boot the ReForge inventory environment to collect hardware details.
sleep 2
exit
""", media_type="text/plain")

@app.get("/boot/deploy.ipxe", response_class=Response)
def ipxe_deploy(mac: str, db: Session = Depends(get_db)):
    normalized = normalize_mac(mac)
    host = db.query(models.Host).filter(models.Host.mac_address == normalized).first()
    if not host:
        return Response(content="#!ipxe\necho Host is not registered in ReForge\nsleep 3\nexit\n", media_type="text/plain")
    job = db.query(models.DeploymentJob).filter(
        models.DeploymentJob.host_id == host.id,
        models.DeploymentJob.status == models.DeploymentStatus.queued
    ).order_by(models.DeploymentJob.created_at.asc()).first()
    if not job:
        return Response(content="#!ipxe\necho No deployment is assigned to this host\nsleep 3\nexit\n", media_type="text/plain")
    return Response(content=f"""#!ipxe
echo ReForge deployment assigned
echo Job: {job.id}
echo Host: {host.hostname}
echo Waiting for the configured imaging environment.
sleep 2
exit
""", media_type="text/plain")
