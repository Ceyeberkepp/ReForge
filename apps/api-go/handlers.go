package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-ldap/ldap/v3"
	"github.com/google/uuid"
)

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if json.NewDecoder(r.Body).Decode(&input) != nil {
		writeJSON(w, 400, map[string]string{"detail": "invalid request"})
		return
	}
	var u User
	if a.db.First(&u, "username = ? AND disabled = ?", input.Username, false).Error != nil || !verifyPassword(u.PasswordHash, input.Password) {
		a.audit(r, "login", "session", "", false, "invalid credentials")
		time.Sleep(250 * time.Millisecond)
		writeJSON(w, 401, map[string]string{"detail": "invalid credentials"})
		return
	}
	s := Session{
		Token:     randomToken(48),
		UserID:    u.ID,
		ExpiresAt: time.Now().UTC().Add(12 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	if a.db.Create(&s).Error != nil {
		writeJSON(w, 500, map[string]string{"detail": "unable to create session"})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "reforge_session",
		Value:    s.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   43200,
	})
	a.audit(r, "login", "session", s.Token[:8], true, "")
	writeJSON(w, 200, map[string]any{"username": u.Username, "role": u.Role})
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("reforge_session"); err == nil {
		_ = a.db.Delete(&Session{}, "token = ?", c.Value).Error
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "reforge_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) me(w http.ResponseWriter, r *http.Request) {
	u := a.currentUser(r)
	writeJSON(w, 200, map[string]any{"username": u.Username, "role": u.Role})
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	sqlDB, err := a.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		writeJSON(w, 503, map[string]any{"status": "degraded"})
		return
	}
	writeJSON(w, 200, map[string]any{
		"status": "ok", "service": "reforge-api", "version": "0.3.0", "database": a.cfg.DBDriver,
	})
}

func (a *App) dashboard(w http.ResponseWriter, r *http.Request) {
	count := func(model any) int64 {
		var n int64
		_ = a.db.Model(model).Count(&n).Error
		return n
	}
	var queued, running, waiting int64
	a.db.Model(&DeploymentJob{}).Where("status = ?", "queued").Count(&queued)
	a.db.Model(&DeploymentJob{}).Where("status = ?", "running").Count(&running)
	a.db.Model(&DeploymentJob{}).Where("status = ?", "waiting").Count(&waiting)
	writeJSON(w, 200, map[string]any{
		"images": count(&GoldImage{}), "departments": count(&Department{}),
		"software": count(&SoftwarePackage{}), "hosts": count(&Host{}),
		"queued_deployments": queued, "running_deployments": running, "waiting_deployments": waiting,
	})
}

func (a *App) listImages(w http.ResponseWriter, r *http.Request) {
	var rows []GoldImage
	a.db.Order("created_at desc").Find(&rows)
	writeJSON(w, 200, rows)
}

func (a *App) saveImage(w http.ResponseWriter, r *http.Request) {
	var row GoldImage
	if json.NewDecoder(r.Body).Decode(&row) != nil {
		writeJSON(w, 400, map[string]string{"detail": "invalid image"})
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		row.ID = uuid.NewString()
		row.CreatedAt = time.Now().UTC()
		if row.Architecture == "" {
			row.Architecture = "x86_64"
		}
		if row.ImageVersion == "" {
			row.ImageVersion = "1.0"
		}
		if a.db.Create(&row).Error != nil {
			writeJSON(w, 409, map[string]string{"detail": "image could not be created"})
			return
		}
	} else {
		row.ID = id
		if a.db.Model(&GoldImage{}).Where("id = ?", id).Updates(map[string]any{
			"name": row.Name, "os_name": row.OSName, "os_version": row.OSVersion,
			"architecture": row.Architecture, "image_version": row.ImageVersion,
			"image_path": row.ImagePath, "sysprep_ready": row.SysprepReady,
			"checksum": row.Checksum, "notes": row.Notes,
		}).Error != nil {
			writeJSON(w, 400, map[string]string{"detail": "image could not be saved"})
			return
		}
		a.db.First(&row, "id = ?", id)
	}
	a.audit(r, "save", "image", row.ID, true, "")
	writeJSON(w, 200, row)
}

func (a *App) deleteImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := a.db.Delete(&GoldImage{}, "id = ?", id).Error
	a.audit(r, "delete", "image", id, err == nil, "")
	if err != nil {
		writeJSON(w, 400, map[string]string{"detail": "delete failed"})
		return
	}
	w.WriteHeader(204)
}

func departmentFromDTO(d DepartmentDTO) Department {
	return Department{
		ID: d.ID, Name: d.Name, Code: d.Code, ImageID: d.ImageID,
		ComputerNamePattern: d.ComputerNamePattern, ADOU: d.ADOU,
		RequiredSoftware: list(d.RequiredSoftwareIDs), OptionalSoftware: list(d.OptionalSoftwareIDs),
		Printers: list(d.Printers), PostInstallScripts: list(d.PostInstallScripts),
	}
}

func departmentDTO(d Department) DepartmentDTO {
	return DepartmentDTO{
		ID: d.ID, Name: d.Name, Code: d.Code, ImageID: d.ImageID,
		ComputerNamePattern: d.ComputerNamePattern, ADOU: d.ADOU,
		RequiredSoftwareIDs: d.RequiredSoftware.Values(), OptionalSoftwareIDs: d.OptionalSoftware.Values(),
		Printers: d.Printers.Values(), PostInstallScripts: d.PostInstallScripts.Values(),
		Config: map[string]any{},
	}
}

func (a *App) listDepartments(w http.ResponseWriter, r *http.Request) {
	var rows []Department
	a.db.Order("name").Find(&rows)
	out := make([]DepartmentDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, departmentDTO(row))
	}
	writeJSON(w, 200, out)
}

func (a *App) saveDepartment(w http.ResponseWriter, r *http.Request) {
	var input DepartmentDTO
	if json.NewDecoder(r.Body).Decode(&input) != nil {
		writeJSON(w, 400, map[string]string{"detail": "invalid department"})
		return
	}
	if id := chi.URLParam(r, "id"); id != "" {
		input.ID = id
	} else {
		input.ID = uuid.NewString()
	}
	row := departmentFromDTO(input)
	if a.db.Save(&row).Error != nil {
		writeJSON(w, 409, map[string]string{"detail": "department could not be saved"})
		return
	}
	a.audit(r, "save", "department", row.ID, true, "")
	writeJSON(w, 200, departmentDTO(row))
}

func (a *App) deleteDepartment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := a.db.Delete(&Department{}, "id = ?", id).Error
	a.audit(r, "delete", "department", id, err == nil, "")
	if err != nil {
		writeJSON(w, 400, map[string]string{"detail": "delete failed"})
		return
	}
	w.WriteHeader(204)
}

func (a *App) listSoftware(w http.ResponseWriter, r *http.Request) {
	var rows []SoftwarePackage
	a.db.Order("install_order,name").Find(&rows)
	writeJSON(w, 200, rows)
}

func (a *App) saveSoftware(w http.ResponseWriter, r *http.Request) {
	var row SoftwarePackage
	if json.NewDecoder(r.Body).Decode(&row) != nil {
		writeJSON(w, 400, map[string]string{"detail": "invalid package"})
		return
	}
	if id := chi.URLParam(r, "id"); id != "" {
		row.ID = id
	} else {
		row.ID = uuid.NewString()
	}
	if a.db.Save(&row).Error != nil {
		writeJSON(w, 409, map[string]string{"detail": "package could not be saved"})
		return
	}
	a.audit(r, "save", "software", row.ID, true, "")
	writeJSON(w, 200, row)
}

func (a *App) deleteSoftware(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := a.db.Delete(&SoftwarePackage{}, "id = ?", id).Error
	a.audit(r, "delete", "software", id, err == nil, "")
	if err != nil {
		writeJSON(w, 400, map[string]string{"detail": "delete failed"})
		return
	}
	w.WriteHeader(204)
}

func (a *App) bootstrapSoftware(w http.ResponseWriter, r *http.Request) {
	starter := []struct {
		Name  string
		Cmd   string
		Order int
	}{
		{"Adobe Acrobat Reader", "AcroRdrDCx64.exe /sAll /rs /rps", 40},
		{"Google Chrome", "msiexec /i googlechromestandaloneenterprise64.msi /qn /norestart", 30},
		{"Microsoft 365 Apps", "setup.exe /configure configuration.xml", 20},
		{"Microsoft Teams", "teamsbootstrapper.exe -p", 35},
		{"7-Zip", "7z-x64.exe /S", 60},
		{"Power BI Desktop", "PBIDesktopSetup_x64.exe -quiet -norestart", 70},
		{"GlobalProtect", "msiexec /i GlobalProtect64.msi /qn /norestart", 80},
		{"NinjaOne Agent", "msiexec /i NinjaOneAgent.msi /qn /norestart", 90},
	}
	created := 0
	for _, item := range starter {
		var count int64
		a.db.Model(&SoftwarePackage{}).Where("name = ?", item.Name).Count(&count)
		if count == 0 {
			a.db.Create(&SoftwarePackage{ID: uuid.NewString(), Name: item.Name, SilentInstall: item.Cmd, InstallOrder: item.Order})
			created++
		}
	}
	writeJSON(w, 200, map[string]any{"created": created})
}

func normalizeMAC(value string) (string, error) {
	raw := strings.NewReplacer(":", "", "-", "", " ", "", ".", "").Replace(strings.ToLower(value))
	if len(raw) != 12 {
		return "", errors.New("invalid MAC address")
	}
	parts := make([]string, 6)
	for i := 0; i < 6; i++ {
		parts[i] = raw[i*2 : i*2+2]
	}
	return strings.Join(parts, ":"), nil
}

func (a *App) listHosts(w http.ResponseWriter, r *http.Request) {
	var rows []Host
	a.db.Order("hostname").Find(&rows)
	writeJSON(w, 200, rows)
}

func (a *App) saveHost(w http.ResponseWriter, r *http.Request) {
	var row Host
	if json.NewDecoder(r.Body).Decode(&row) != nil {
		writeJSON(w, 400, map[string]string{"detail": "invalid host"})
		return
	}
	mac, err := normalizeMAC(row.MACAddress)
	if err != nil {
		writeJSON(w, 400, map[string]string{"detail": err.Error()})
		return
	}
	row.MACAddress = mac
	if id := chi.URLParam(r, "id"); id != "" {
		row.ID = id
	} else {
		row.ID = uuid.NewString()
	}
	if a.db.Save(&row).Error != nil {
		writeJSON(w, 409, map[string]string{"detail": "host could not be saved"})
		return
	}
	a.audit(r, "save", "host", row.ID, true, "")
	writeJSON(w, 200, row)
}

func (a *App) deleteHost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := a.db.Delete(&Host{}, "id = ?", id).Error
	a.audit(r, "delete", "host", id, err == nil, "")
	if err != nil {
		writeJSON(w, 400, map[string]string{"detail": "delete failed"})
		return
	}
	w.WriteHeader(204)
}

func (a *App) registerHost(w http.ResponseWriter, r *http.Request) {
	var input struct {
		MACAddress   string `json:"mac_address"`
		SerialNumber string `json:"serial_number"`
		Manufacturer string `json:"manufacturer"`
		Model        string `json:"model"`
		Hostname     string `json:"hostname"`
	}
	if json.NewDecoder(r.Body).Decode(&input) != nil {
		writeJSON(w, 400, map[string]string{"detail": "invalid request"})
		return
	}
	mac, err := normalizeMAC(input.MACAddress)
	if err != nil {
		writeJSON(w, 400, map[string]string{"detail": err.Error()})
		return
	}
	var host Host
	now := time.Now().UTC()
	if a.db.First(&host, "mac_address = ?", mac).Error != nil {
		host = Host{
			ID: uuid.NewString(), MACAddress: mac, Hostname: input.Hostname,
			SerialNumber: input.SerialNumber, Manufacturer: input.Manufacturer,
			Model: input.Model, LastSeen: &now,
		}
		if host.Hostname == "" {
			host.Hostname = "NEW-" + strings.ToUpper(strings.ReplaceAll(mac, ":", "")[6:])
		}
		a.db.Create(&host)
	} else {
		host.LastSeen = &now
		if input.SerialNumber != "" { host.SerialNumber = input.SerialNumber }
		if input.Manufacturer != "" { host.Manufacturer = input.Manufacturer }
		if input.Model != "" { host.Model = input.Model }
		a.db.Save(&host)
	}
	writeJSON(w, 200, host)
}

func (a *App) getDirectory(w http.ResponseWriter, r *http.Request) {
	var row DirectoryConfig
	if a.db.First(&row, "id = ?", "default").Error != nil {
		row = DirectoryConfig{ID: "default", Protocol: "ldaps", Port: 636}
		a.db.Create(&row)
	}
	writeJSON(w, 200, row)
}

func (a *App) saveDirectory(w http.ResponseWriter, r *http.Request) {
	var row DirectoryConfig
	if json.NewDecoder(r.Body).Decode(&row) != nil {
		writeJSON(w, 400, map[string]string{"detail": "invalid directory config"})
		return
	}
	row.ID = "default"
	if a.db.Save(&row).Error != nil {
		writeJSON(w, 400, map[string]string{"detail": "save failed"})
		return
	}
	a.audit(r, "save", "directory", "default", true, "")
	writeJSON(w, 200, row)
}

func (a *App) testDirectory(w http.ResponseWriter, r *http.Request) {
	var input DirectoryTestRequest
	if json.NewDecoder(r.Body).Decode(&input) != nil {
		writeJSON(w, 400, map[string]string{"detail": "invalid directory request"})
		return
	}
	if input.DomainController == "" || input.ServiceAccount == "" || input.Password == "" {
		writeJSON(w, 400, map[string]string{"detail": "domain controller, service account, and password are required"})
		return
	}
	host := input.DomainController
	port := input.Port
	if port == 0 {
		if input.Protocol == "ldaps" { port = 636 } else { port = 389 }
	}
	address := host + ":" + strconv.Itoa(port)
	var conn *ldap.Conn
	var err error
	switch strings.ToLower(input.Protocol) {
	case "ldaps":
		conn, err = ldap.DialURL("ldaps://"+address, ldap.DialWithTLSConfig(&tls.Config{
			MinVersion: tls.VersionTLS12, ServerName: host,
		}))
	case "starttls":
		conn, err = ldap.DialURL("ldap://"+address)
		if err == nil {
			err = conn.StartTLS(&tls.Config{MinVersion: tls.VersionTLS12, ServerName: host})
		}
	case "ldap":
		if !a.cfg.AllowInsecureLDAP {
			writeJSON(w, 400, map[string]string{"detail": "plaintext LDAP is disabled; use LDAPS or StartTLS"})
			return
		}
		conn, err = ldap.DialURL("ldap://"+address)
	default:
		writeJSON(w, 400, map[string]string{"detail": "protocol must be ldaps, starttls, or ldap"})
		return
	}
	if err != nil {
		a.audit(r, "test", "directory", "default", false, "connection failed")
		writeJSON(w, 400, map[string]string{"detail": "directory connection failed: " + err.Error()})
		return
	}
	defer conn.Close()
	if err = conn.Bind(input.ServiceAccount, input.Password); err != nil {
		a.audit(r, "test", "directory", "default", false, "bind failed")
		writeJSON(w, 400, map[string]string{"detail": "directory bind failed: " + err.Error()})
		return
	}
	a.audit(r, "test", "directory", "default", true, "")
	writeJSON(w, 200, map[string]any{"ok": true, "message": "Directory bind succeeded", "server": host, "protocol": input.Protocol})
}

func deploymentDTO(row DeploymentJob) DeploymentDTO {
	return DeploymentDTO{
		ID: row.ID, HostID: row.HostID, ImageID: row.ImageID,
		DepartmentID: row.DepartmentID, OptionalSoftware: row.OptionalSoftware.Values(),
		Status: row.Status, Progress: row.Progress, CurrentStep: row.CurrentStep, CreatedAt: row.CreatedAt,
	}
}

func (a *App) listDeployments(w http.ResponseWriter, r *http.Request) {
	var rows []DeploymentJob
	a.db.Order("created_at desc").Find(&rows)
	out := make([]DeploymentDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, deploymentDTO(row))
	}
	writeJSON(w, 200, out)
}

func (a *App) createDeployment(w http.ResponseWriter, r *http.Request) {
	var input DeploymentDTO
	if json.NewDecoder(r.Body).Decode(&input) != nil || input.HostID == "" || input.ImageID == "" {
		writeJSON(w, 400, map[string]string{"detail": "host and image are required"})
		return
	}
	var hostCount, imageCount int64
	a.db.Model(&Host{}).Where("id = ?", input.HostID).Count(&hostCount)
	a.db.Model(&GoldImage{}).Where("id = ?", input.ImageID).Count(&imageCount)
	if hostCount == 0 || imageCount == 0 {
		writeJSON(w, 404, map[string]string{"detail": "host or image not found"})
		return
	}
	row := DeploymentJob{
		ID: uuid.NewString(), HostID: input.HostID, ImageID: input.ImageID,
		DepartmentID: input.DepartmentID, OptionalSoftware: list(input.OptionalSoftware),
		Status: "queued", Progress: 0, CurrentStep: "Queued", CreatedAt: time.Now().UTC(),
	}
	if a.db.Create(&row).Error != nil {
		writeJSON(w, 400, map[string]string{"detail": "deployment could not be created"})
		return
	}
	a.audit(r, "create", "deployment", row.ID, true, "")
	writeJSON(w, 200, deploymentDTO(row))
}

func (a *App) deploymentAction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	action := chi.URLParam(r, "action")
	var row DeploymentJob
	if a.db.First(&row, "id = ?", id).Error != nil {
		writeJSON(w, 404, map[string]string{"detail": "deployment not found"})
		return
	}
	switch action {
	case "cancel":
		row.Status, row.CurrentStep = "canceled", "Canceled"
	case "retry":
		row.Status, row.Progress, row.CurrentStep = "queued", 0, "Queued"
	default:
		writeJSON(w, 404, map[string]string{"detail": "unknown action"})
		return
	}
	a.db.Save(&row)
	a.audit(r, action, "deployment", id, true, "")
	writeJSON(w, 200, deploymentDTO(row))
}

func (a *App) deploymentPlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var job DeploymentJob
	if a.db.First(&job, "id = ?", id).Error != nil {
		writeJSON(w, 404, map[string]string{"detail": "deployment not found"})
		return
	}
	var host Host
	var image GoldImage
	a.db.First(&host, "id = ?", job.HostID)
	a.db.First(&image, "id = ?", job.ImageID)

	var dep *Department
	var softwareIDs []string
	if job.DepartmentID != nil {
		var row Department
		if a.db.First(&row, "id = ?", *job.DepartmentID).Error == nil {
			dep = &row
			softwareIDs = append(softwareIDs, row.RequiredSoftware.Values()...)
		}
	}
	softwareIDs = append(softwareIDs, job.OptionalSoftware.Values()...)
	seen := map[string]bool{}
	packages := []SoftwarePackage{}
	for _, sid := range softwareIDs {
		if sid == "" || seen[sid] { continue }
		seen[sid] = true
		var pkg SoftwarePackage
		if a.db.First(&pkg, "id = ?", sid).Error == nil {
			packages = append(packages, pkg)
		}
	}
	response := map[string]any{
		"job_id": job.ID, "host": host, "image": image, "software": packages,
		"steps": []string{
			"Validate host", "Boot imaging environment", "Partition disk", "Restore gold image",
			"Apply drivers", "Boot operating system", "Rename computer", "Install required software",
			"Install optional software", "Join directory", "Move to department OU",
			"Install printers", "Run post-install scripts", "Run updates", "Validate deployment",
		},
	}
	if dep != nil {
		response["department"] = departmentDTO(*dep)
	}
	writeJSON(w, 200, response)
}

func (a *App) workerUpdate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status      *string `json:"status"`
		Progress    *int    `json:"progress"`
		CurrentStep *string `json:"current_step"`
	}
	if json.NewDecoder(r.Body).Decode(&input) != nil {
		writeJSON(w, 400, map[string]string{"detail": "invalid update"})
		return
	}
	var row DeploymentJob
	if a.db.First(&row, "id = ?", id).Error != nil {
		writeJSON(w, 404, map[string]string{"detail": "deployment not found"})
		return
	}
	if input.Status != nil { row.Status = *input.Status }
	if input.Progress != nil { row.Progress = *input.Progress }
	if input.CurrentStep != nil { row.CurrentStep = *input.CurrentStep }
	a.db.Save(&row)
	writeJSON(w, 200, deploymentDTO(row))
}

func (a *App) getPXE(w http.ResponseWriter, r *http.Request) {
	var row PXEConfig
	if a.db.First(&row, "id = ?", "default").Error != nil {
		row = PXEConfig{
			ID: "default", Enabled: true, ServerURL: "http://reforge.local:8080",
			BIOSBootFile: "undionly.kpxe", UEFIBootFile: "ipxe.efi",
			DHCPMode: "existing", BootMenuTimeout: 5,
		}
		a.db.Create(&row)
	}
	writeJSON(w, 200, row)
}

func (a *App) savePXE(w http.ResponseWriter, r *http.Request) {
	var row PXEConfig
	if json.NewDecoder(r.Body).Decode(&row) != nil {
		writeJSON(w, 400, map[string]string{"detail": "invalid PXE config"})
		return
	}
	row.ID = "default"
	if row.BootMenuTimeout < 1 { row.BootMenuTimeout = 5 }
	if a.db.Save(&row).Error != nil {
		writeJSON(w, 400, map[string]string{"detail": "save failed"})
		return
	}
	a.audit(r, "save", "pxe", "default", true, "")
	writeJSON(w, 200, row)
}

func (a *App) ipxe(w http.ResponseWriter, r *http.Request) {
	var p PXEConfig
	if a.db.First(&p, "id = ?", "default").Error != nil || !p.Enabled {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "#!ipxe")
		fmt.Fprintln(w, "echo ReForge PXE is disabled")
		fmt.Fprintln(w, "exit")
		return
	}
	timeout := p.BootMenuTimeout * 1000
	apiVar := "$" + "{reforge-api}"
	macVar := "$" + "{net0/mac}"
	targetVar := "$" + "{target}"
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, "#!ipxe")
	fmt.Fprintln(w, "set reforge-api "+p.ServerURL)
	fmt.Fprintln(w, "menu ReForge Deployment")
	fmt.Fprintln(w, "item deploy Deploy assigned image")
	fmt.Fprintln(w, "item register Register this device")
	fmt.Fprintln(w, "item local Boot local disk")
	fmt.Fprintf(w, "choose --default deploy --timeout %d target && goto %s\n", timeout, targetVar)
	fmt.Fprintln(w, ":deploy")
	fmt.Fprintln(w, "chain "+apiVar+"/boot/deploy.ipxe?mac="+macVar+" || goto local")
	fmt.Fprintln(w, ":register")
	fmt.Fprintln(w, "chain "+apiVar+"/boot/register.ipxe?mac="+macVar+" || goto local")
	fmt.Fprintln(w, ":local")
	fmt.Fprintln(w, "exit")
}

func (a *App) deployIPXE(w http.ResponseWriter, r *http.Request) {
	mac, err := normalizeMAC(r.URL.Query().Get("mac"))
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, "#!ipxe")
	if err != nil {
		fmt.Fprintln(w, "echo Invalid MAC")
		fmt.Fprintln(w, "exit")
		return
	}
	var host Host
	if a.db.First(&host, "mac_address = ?", mac).Error != nil {
		fmt.Fprintln(w, "echo Device is not registered in ReForge")
		fmt.Fprintln(w, "sleep 3")
		fmt.Fprintln(w, "exit")
		return
	}
	var job DeploymentJob
	if a.db.Where("host_id = ? AND status IN ?", host.ID, []string{"queued", "waiting"}).Order("created_at").First(&job).Error != nil {
		fmt.Fprintln(w, "echo No deployment assigned")
		fmt.Fprintln(w, "sleep 3")
		fmt.Fprintln(w, "exit")
		return
	}
	fmt.Fprintln(w, "echo ReForge deployment assigned")
	fmt.Fprintln(w, "echo Host: "+host.Hostname)
	fmt.Fprintln(w, "echo Job: "+job.ID)
	fmt.Fprintln(w, "echo Imaging-node handoff is required")
	fmt.Fprintln(w, "sleep 2")
	fmt.Fprintln(w, "exit")
}

func (a *App) registerIPXE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, "#!ipxe")
	fmt.Fprintln(w, "echo ReForge inventory boot")
	fmt.Fprintln(w, "echo Start the ReForge inventory environment to collect hardware details")
	fmt.Fprintln(w, "sleep 2")
	fmt.Fprintln(w, "exit")
}

func (a *App) listAudit(w http.ResponseWriter, r *http.Request) {
	var rows []AuditEvent
	a.db.Order("at desc").Limit(500).Find(&rows)
	writeJSON(w, 200, rows)
}
