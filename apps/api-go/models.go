package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type Config struct {
	Listen             string
	DBDriver           string
	DBDSN              string
	AdminUser          string
	AdminPassword      string
	AllowedOrigin      string
	WorkerToken        string
	AllowInsecureLDAP  bool
}

func env(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return d
}

func loadConfig() Config {
	return Config{
		Listen:            env("REFORGE_LISTEN", ":8080"),
		DBDriver:          strings.ToLower(env("REFORGE_DB_DRIVER", "postgres")),
		DBDSN:             env("REFORGE_DB_DSN", "host=db user=reforge password=change-me dbname=reforge port=5432 sslmode=disable"),
		AdminUser:         env("REFORGE_ADMIN_USER", "admin"),
		AdminPassword:     env("REFORGE_ADMIN_PASSWORD", ""),
		AllowedOrigin:     env("REFORGE_ALLOWED_ORIGIN", "http://localhost:5173"),
		WorkerToken:       env("REFORGE_WORKER_TOKEN", ""),
		AllowInsecureLDAP: strings.EqualFold(env("REFORGE_ALLOW_INSECURE_LDAP", "false"), "true"),
	}
}

type JSONList string

func list(v []string) JSONList {
	if v == nil {
		return "[]"
	}
	b, _ := jsonMarshal(v)
	return JSONList(b)
}

func (j JSONList) Values() []string {
	var v []string
	_ = jsonUnmarshal([]byte(j), &v)
	if v == nil {
		return []string{}
	}
	return v
}

type GoldImage struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"uniqueIndex;size:160" json:"name"`
	OSName        string    `json:"os_name"`
	OSVersion     string    `json:"os_version"`
	Architecture  string    `json:"architecture"`
	ImageVersion  string    `json:"image_version"`
	ImagePath     string    `json:"image_path"`
	SysprepReady  bool      `json:"sysprep_ready"`
	Checksum      string    `json:"checksum"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
}

type SoftwarePackage struct {
	ID                string `gorm:"primaryKey" json:"id"`
	Name              string `gorm:"index;size:160" json:"name"`
	Version           string `json:"version"`
	InstallerPath     string `json:"installer_path"`
	SilentInstall     string `json:"silent_install"`
	DetectionRule     string `json:"detection_rule"`
	UninstallCommand  string `json:"uninstall_command"`
	InstallOrder      int    `json:"install_order"`
	RequiredByDefault bool   `json:"required_by_default"`
}

type Department struct {
	ID                  string   `gorm:"primaryKey"`
	Name                string   `gorm:"uniqueIndex;size:120"`
	Code                string   `gorm:"uniqueIndex;size:32"`
	ImageID             *string
	ComputerNamePattern string
	ADOU                string
	RequiredSoftware    JSONList `gorm:"type:text"`
	OptionalSoftware    JSONList `gorm:"type:text"`
	Printers            JSONList `gorm:"type:text"`
	PostInstallScripts  JSONList `gorm:"type:text"`
}

type Host struct {
	ID           string     `gorm:"primaryKey" json:"id"`
	Hostname     string     `gorm:"uniqueIndex;size:160" json:"hostname"`
	MACAddress   *string    `gorm:"uniqueIndex;size:32" json:"mac_address"`
	HardwareUUID *string    `gorm:"uniqueIndex;size:80" json:"hardware_uuid"`
	SerialNumber string     `gorm:"index" json:"serial_number"`
	Manufacturer string     `json:"manufacturer"`
	Model        string     `json:"model"`
	DepartmentID *string    `json:"department_id"`
	LastSeen     *time.Time `json:"last_seen"`
}

type DirectoryConfig struct {
	ID                string `gorm:"primaryKey" json:"id"`
	Domain            string `json:"domain"`
	DomainController  string `json:"domain_controller"`
	Protocol          string `json:"protocol"`
	Port              int    `json:"port"`
	BaseDN            string `json:"base_dn"`
	ServiceAccount    string `json:"service_account"`
	DefaultComputerOU string `json:"default_computer_ou"`
}

type DeploymentJob struct {
	ID               string    `gorm:"primaryKey"`
	HostID           string    `gorm:"index"`
	ImageID          string    `gorm:"index"`
	DepartmentID     *string   `gorm:"index"`
	OptionalSoftware JSONList  `gorm:"type:text"`
	Status           string    `gorm:"index;size:20"`
	Progress         int
	CurrentStep      string
	CreatedAt        time.Time
}

type PXEConfig struct {
	ID              string `gorm:"primaryKey" json:"id"`
	Enabled         bool   `json:"enabled"`
	ServerURL       string `json:"server_url"`
	TFTPServer      string `json:"tftp_server"`
	BIOSBootFile    string `json:"bios_boot_file"`
	UEFIBootFile    string `json:"uefi_boot_file"`
	DHCPMode        string `json:"dhcp_mode"`
	DHCPServer      string `json:"dhcp_server"`
	NextServer      string `json:"next_server"`
	BootMenuTimeout int    `json:"boot_menu_timeout"`
	AllowUnknown    bool   `json:"allow_unknown"`
	MenuTitle       string `json:"menu_title"`
	DefaultItem     string `json:"default_item"`
	ShowDeploy      bool   `json:"show_deploy"`
	ShowRegister    bool   `json:"show_register"`
	ShowDiagnostics bool   `json:"show_diagnostics"`
	ShowLocalBoot   bool   `json:"show_local_boot"`
	BrandName       string `json:"brand_name"`
	MenuSubtitle    string `json:"menu_subtitle"`
	SupportText     string `json:"support_text"`
	AccentColor     string `json:"accent_color"`
	ShowLogo        bool   `json:"show_logo"`
	ShowBackground  bool   `json:"show_background"`
	RequireLogin    bool   `json:"require_login"`
	InstallTitle     string `json:"install_title"`
	InstallSubtitle  string `json:"install_subtitle"`
	CaptureTitle     string `json:"capture_title"`
	CaptureSubtitle  string `json:"capture_subtitle"`
	LoadingTitle     string `json:"loading_title"`
	LoadingMessage   string `json:"loading_message"`
	TextColor        string `json:"text_color"`
	PanelColor       string `json:"panel_color"`
	OverlayOpacity   int    `json:"overlay_opacity"`
}

type AuditEvent struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	At         time.Time `gorm:"index" json:"at"`
	Actor      string    `gorm:"index" json:"actor"`
	Action     string    `gorm:"index" json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id"`
	RemoteIP   string    `json:"remote_ip"`
	Success    bool      `json:"success"`
	Detail     string    `json:"detail"`
}

type User struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:120" json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `gorm:"size:40" json:"role"`
	GroupID      *string   `gorm:"index" json:"group_id"`
	Disabled     bool      `json:"disabled"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserGroup struct {
	ID          string   `gorm:"primaryKey" json:"id"`
	Name        string   `gorm:"uniqueIndex;size:120" json:"name"`
	Description string   `json:"description"`
	Permissions JSONList `gorm:"type:text" json:"-"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserGroupDTO struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

type ISOImage struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"uniqueIndex;size:160" json:"name"`
	OSFamily     string    `json:"os_family"`
	Version      string    `json:"version"`
	Architecture string    `json:"architecture"`
	SourcePath   string    `json:"source_path"`
	Checksum     string    `json:"checksum"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
}

type CloneImage struct {
	ID             string    `gorm:"primaryKey" json:"id"`
	Name           string    `gorm:"uniqueIndex;size:160" json:"name"`
	SourceHostID   *string   `gorm:"index" json:"source_host_id"`
	OSFamily       string    `json:"os_family"`
	Architecture   string    `json:"architecture"`
	HardwareFamily string    `json:"hardware_family"`
	ImagePath      string    `json:"image_path"`
	Checksum       string    `json:"checksum"`
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
}

type BrandingAsset struct {
	ID          string    `gorm:"primaryKey;size:32" json:"id"`
	ContentType string    `json:"content_type"`
	Data        []byte    `json:"-"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PXETask struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	HostMAC     string    `gorm:"index;size:32" json:"host_mac"`
	Action      string    `gorm:"index;size:32" json:"action"`
	SourceType  string    `gorm:"index;size:32" json:"source_type"`
	SourceID    string    `gorm:"index" json:"source_id"`
	RequestedBy string    `gorm:"index;size:120" json:"requested_by"`
	Status      string    `gorm:"index;size:24" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type PXEAccessToken struct {
	Token     string    `gorm:"primaryKey;size:96" json:"-"`
	UserID    string    `gorm:"index" json:"user_id"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Session struct {
	Token     string `gorm:"primaryKey;size:96"`
	UserID    string `gorm:"index"`
	ExpiresAt time.Time `gorm:"index"`
	CreatedAt time.Time
}

type DepartmentDTO struct {
	ID                   string         `json:"id,omitempty"`
	Name                 string         `json:"name"`
	Code                 string         `json:"code"`
	ImageID              *string        `json:"image_id"`
	ComputerNamePattern  string         `json:"computer_name_pattern"`
	ADOU                 string         `json:"ad_ou"`
	RequiredSoftwareIDs  []string       `json:"required_software_ids"`
	OptionalSoftwareIDs  []string       `json:"optional_software_ids"`
	Printers             []string       `json:"printers"`
	PostInstallScripts   []string       `json:"post_install_scripts"`
	Config               map[string]any `json:"config"`
}

type DeploymentDTO struct {
	ID               string    `json:"id,omitempty"`
	HostID           string    `json:"host_id"`
	ImageID          string    `json:"image_id"`
	DepartmentID     *string   `json:"department_id"`
	OptionalSoftware []string  `json:"optional_software"`
	Status           string    `json:"status,omitempty"`
	Progress         int       `json:"progress,omitempty"`
	CurrentStep      string    `json:"current_step,omitempty"`
	CreatedAt        time.Time `json:"created_at,omitempty"`
}

type DirectoryTestRequest struct {
	Domain            string `json:"domain"`
	DomainController  string `json:"domain_controller"`
	Protocol          string `json:"protocol"`
	Port              int    `json:"port"`
	BaseDN            string `json:"base_dn"`
	ServiceAccount    string `json:"service_account"`
	DefaultComputerOU string `json:"default_computer_ou"`
	Password          string `json:"password"`
}

func openDB(c Config) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch c.DBDriver {
	case "postgres", "postgresql":
		dialector = postgres.Open(c.DBDSN)
	case "mysql", "mariadb":
		dialector = mysql.Open(c.DBDSN)
	case "sqlserver", "mssql":
		dialector = sqlserver.Open(c.DBDSN)
	default:
		return nil, fmt.Errorf("unsupported REFORGE_DB_DRIVER %q", c.DBDriver)
	}

	var db *gorm.DB
	var err error
	for attempt := 1; attempt <= 30; attempt++ {
		db, err = gorm.Open(dialector, &gorm.Config{})
		if err == nil {
			sqlDB, sqlErr := db.DB()
			if sqlErr == nil && sqlDB.Ping() == nil {
				return db, nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("database unavailable: %w", err)
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&GoldImage{}, &SoftwarePackage{}, &Department{}, &Host{},
		&DirectoryConfig{}, &DeploymentJob{}, &PXEConfig{},
		&AuditEvent{}, &User{}, &UserGroup{}, &ISOImage{}, &CloneImage{}, &BrandingAsset{}, &PXETask{}, &PXEAccessToken{}, &Session{},
	)
}

func validateConfig(c Config) error {
	if len(c.AdminPassword) < 12 {
		return errors.New("REFORGE_ADMIN_PASSWORD must be at least 12 characters on first startup")
	}
	return nil
}
