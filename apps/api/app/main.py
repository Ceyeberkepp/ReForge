from fastapi import Depends, FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from ldap3 import Connection, Server, Tls
from sqlalchemy.orm import Session
import ssl
from .database import Base, engine, get_db
from . import models, schemas

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="ReForge API",
    version="0.1.0",
    description="Control plane for ReForge endpoint imaging and deployment.",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/health")
def health():
    return {"status": "ok", "service": "reforge-api", "version": "0.1.0"}

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
    }

@app.get("/api/images", response_model=list[schemas.GoldImageOut])
def list_images(db: Session = Depends(get_db)):
    return db.query(models.GoldImage).order_by(models.GoldImage.created_at.desc()).all()

@app.post("/api/images", response_model=schemas.GoldImageOut)
def create_image(payload: schemas.GoldImageCreate, db: Session = Depends(get_db)):
    item = models.GoldImage(**payload.model_dump())
    db.add(item); db.commit(); db.refresh(item)
    return item

@app.get("/api/departments", response_model=list[schemas.DepartmentOut])
def list_departments(db: Session = Depends(get_db)):
    return db.query(models.Department).order_by(models.Department.name).all()

@app.post("/api/departments", response_model=schemas.DepartmentOut)
def create_department(payload: schemas.DepartmentCreate, db: Session = Depends(get_db)):
    item = models.Department(**payload.model_dump())
    db.add(item); db.commit(); db.refresh(item)
    return item

@app.get("/api/software", response_model=list[schemas.SoftwareOut])
def list_software(db: Session = Depends(get_db)):
    return db.query(models.SoftwarePackage).order_by(
        models.SoftwarePackage.install_order,
        models.SoftwarePackage.name
    ).all()

@app.post("/api/software", response_model=schemas.SoftwareOut)
def create_software(payload: schemas.SoftwareCreate, db: Session = Depends(get_db)):
    item = models.SoftwarePackage(**payload.model_dump())
    db.add(item); db.commit(); db.refresh(item)
    return item

@app.get("/api/hosts", response_model=list[schemas.HostOut])
def list_hosts(db: Session = Depends(get_db)):
    return db.query(models.Host).order_by(models.Host.hostname).all()

@app.post("/api/hosts", response_model=schemas.HostOut)
def create_host(payload: schemas.HostCreate, db: Session = Depends(get_db)):
    item = models.Host(**payload.model_dump())
    db.add(item); db.commit(); db.refresh(item)
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
    for key, value in payload.model_dump().items():
        setattr(item, key, value)
    db.commit(); db.refresh(item)
    return item

@app.post("/api/directory/test")
def test_directory(payload: schemas.DirectoryTestRequest):
    use_ssl = payload.protocol.lower() == "ldaps"
    tls = Tls(validate=ssl.CERT_REQUIRED) if use_ssl else None
    try:
        server = Server(
            payload.domain_controller,
            port=payload.port,
            use_ssl=use_ssl,
            tls=tls,
            connect_timeout=5,
        )
        conn = Connection(
            server,
            user=payload.service_account,
            password=payload.password.get_secret_value(),
            auto_bind=True,
            receive_timeout=5,
        )
        conn.unbind()
        return {
            "ok": True,
            "message": "Directory bind succeeded",
            "server": payload.domain_controller,
            "protocol": payload.protocol,
        }
    except Exception as exc:
        raise HTTPException(
            status_code=400,
            detail=f"Directory bind failed: {type(exc).__name__}: {exc}",
        )

@app.get("/api/deployments", response_model=list[schemas.DeploymentOut])
def list_deployments(db: Session = Depends(get_db)):
    return db.query(models.DeploymentJob).order_by(models.DeploymentJob.created_at.desc()).all()

@app.post("/api/deployments", response_model=schemas.DeploymentOut)
def create_deployment(payload: schemas.DeploymentCreate, db: Session = Depends(get_db)):
    if not db.get(models.Host, payload.host_id):
        raise HTTPException(404, "Host not found")
    if not db.get(models.GoldImage, payload.image_id):
        raise HTTPException(404, "Gold image not found")
    item = models.DeploymentJob(**payload.model_dump())
    db.add(item); db.commit(); db.refresh(item)
    return item
