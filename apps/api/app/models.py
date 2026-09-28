import enum
import uuid
from datetime import datetime
from sqlalchemy import Boolean, DateTime, Enum, ForeignKey, Integer, JSON, String, Text
from sqlalchemy.orm import Mapped, mapped_column
from .database import Base

def uuid_str() -> str:
    return str(uuid.uuid4())

class DeploymentStatus(str, enum.Enum):
    queued = "queued"
    running = "running"
    succeeded = "succeeded"
    failed = "failed"
    canceled = "canceled"

class GoldImage(Base):
    __tablename__ = "gold_images"
    id: Mapped[str] = mapped_column(String, primary_key=True, default=uuid_str)
    name: Mapped[str] = mapped_column(String(160), unique=True, index=True)
    os_name: Mapped[str] = mapped_column(String(100))
    os_version: Mapped[str] = mapped_column(String(100), default="")
    architecture: Mapped[str] = mapped_column(String(32), default="x86_64")
    image_version: Mapped[str] = mapped_column(String(32), default="1.0")
    image_path: Mapped[str | None] = mapped_column(String(500), nullable=True)
    sysprep_ready: Mapped[bool] = mapped_column(Boolean, default=False)
    checksum: Mapped[str | None] = mapped_column(String(128), nullable=True)
    notes: Mapped[str] = mapped_column(Text, default="")
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)

class Department(Base):
    __tablename__ = "departments"
    id: Mapped[str] = mapped_column(String, primary_key=True, default=uuid_str)
    name: Mapped[str] = mapped_column(String(120), unique=True, index=True)
    code: Mapped[str] = mapped_column(String(32), unique=True)
    image_id: Mapped[str | None] = mapped_column(ForeignKey("gold_images.id"), nullable=True)
    computer_name_pattern: Mapped[str] = mapped_column(String(100), default="{DEPT}-{SERIAL}")
    ad_ou: Mapped[str] = mapped_column(String(500), default="")
    required_software_ids: Mapped[list] = mapped_column(JSON, default=list)
    optional_software_ids: Mapped[list] = mapped_column(JSON, default=list)
    printers: Mapped[list] = mapped_column(JSON, default=list)
    post_install_scripts: Mapped[list] = mapped_column(JSON, default=list)
    config: Mapped[dict] = mapped_column(JSON, default=dict)

class SoftwarePackage(Base):
    __tablename__ = "software_packages"
    id: Mapped[str] = mapped_column(String, primary_key=True, default=uuid_str)
    name: Mapped[str] = mapped_column(String(160), index=True)
    version: Mapped[str] = mapped_column(String(80), default="")
    installer_path: Mapped[str] = mapped_column(String(500), default="")
    silent_install: Mapped[str] = mapped_column(Text, default="")
    detection_rule: Mapped[str] = mapped_column(Text, default="")
    uninstall_command: Mapped[str] = mapped_column(Text, default="")
    install_order: Mapped[int] = mapped_column(Integer, default=100)
    required_by_default: Mapped[bool] = mapped_column(Boolean, default=False)

class Host(Base):
    __tablename__ = "hosts"
    id: Mapped[str] = mapped_column(String, primary_key=True, default=uuid_str)
    hostname: Mapped[str] = mapped_column(String(160), unique=True, index=True)
    mac_address: Mapped[str] = mapped_column(String(32), unique=True, index=True)
    serial_number: Mapped[str] = mapped_column(String(160), default="")
    manufacturer: Mapped[str] = mapped_column(String(160), default="")
    model: Mapped[str] = mapped_column(String(160), default="")
    department_id: Mapped[str | None] = mapped_column(ForeignKey("departments.id"), nullable=True)
    last_seen: Mapped[datetime | None] = mapped_column(DateTime, nullable=True)

class DirectoryConfig(Base):
    __tablename__ = "directory_config"
    id: Mapped[str] = mapped_column(String, primary_key=True, default="default")
    domain: Mapped[str] = mapped_column(String(255), default="")
    domain_controller: Mapped[str] = mapped_column(String(255), default="")
    protocol: Mapped[str] = mapped_column(String(16), default="ldaps")
    port: Mapped[int] = mapped_column(Integer, default=636)
    base_dn: Mapped[str] = mapped_column(String(500), default="")
    service_account: Mapped[str] = mapped_column(String(255), default="")
    default_computer_ou: Mapped[str] = mapped_column(String(500), default="")

class DeploymentJob(Base):
    __tablename__ = "deployment_jobs"
    id: Mapped[str] = mapped_column(String, primary_key=True, default=uuid_str)
    host_id: Mapped[str] = mapped_column(ForeignKey("hosts.id"))
    image_id: Mapped[str] = mapped_column(ForeignKey("gold_images.id"))
    department_id: Mapped[str | None] = mapped_column(ForeignKey("departments.id"), nullable=True)
    optional_software: Mapped[list] = mapped_column(JSON, default=list)
    status: Mapped[DeploymentStatus] = mapped_column(Enum(DeploymentStatus), default=DeploymentStatus.queued)
    progress: Mapped[int] = mapped_column(Integer, default=0)
    current_step: Mapped[str] = mapped_column(String(255), default="Queued")
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)
