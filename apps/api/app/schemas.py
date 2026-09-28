from datetime import datetime
from typing import Any
from pydantic import BaseModel, ConfigDict, Field, SecretStr
from .models import DeploymentStatus

class ORMModel(BaseModel):
    model_config = ConfigDict(from_attributes=True)

class GoldImageCreate(BaseModel):
    name: str
    os_name: str
    os_version: str = ""
    architecture: str = "x86_64"
    image_version: str = "1.0"
    image_path: str | None = None
    sysprep_ready: bool = False
    checksum: str | None = None
    notes: str = ""

class GoldImageOut(GoldImageCreate, ORMModel):
    id: str
    created_at: datetime

class DepartmentCreate(BaseModel):
    name: str
    code: str
    image_id: str | None = None
    computer_name_pattern: str = "{DEPT}-{SERIAL}"
    ad_ou: str = ""
    required_software_ids: list[str] = Field(default_factory=list)
    optional_software_ids: list[str] = Field(default_factory=list)
    printers: list[str] = Field(default_factory=list)
    post_install_scripts: list[str] = Field(default_factory=list)
    config: dict[str, Any] = Field(default_factory=dict)

class DepartmentOut(DepartmentCreate, ORMModel):
    id: str

class SoftwareCreate(BaseModel):
    name: str
    version: str = ""
    installer_path: str = ""
    silent_install: str = ""
    detection_rule: str = ""
    uninstall_command: str = ""
    install_order: int = 100
    required_by_default: bool = False

class SoftwareOut(SoftwareCreate, ORMModel):
    id: str

class HostCreate(BaseModel):
    hostname: str
    mac_address: str
    serial_number: str = ""
    manufacturer: str = ""
    model: str = ""
    department_id: str | None = None

class HostOut(HostCreate, ORMModel):
    id: str
    last_seen: datetime | None = None

class DirectoryConfigIn(BaseModel):
    domain: str = ""
    domain_controller: str = ""
    protocol: str = "ldaps"
    port: int = 636
    base_dn: str = ""
    service_account: str = ""
    default_computer_ou: str = ""

class DirectoryConfigOut(DirectoryConfigIn, ORMModel):
    id: str

class DirectoryTestRequest(DirectoryConfigIn):
    password: SecretStr

class DeploymentCreate(BaseModel):
    host_id: str
    image_id: str
    department_id: str | None = None
    optional_software: list[str] = Field(default_factory=list)

class DeploymentOut(DeploymentCreate, ORMModel):
    id: str
    status: DeploymentStatus
    progress: int
    current_step: str
    created_at: datetime
