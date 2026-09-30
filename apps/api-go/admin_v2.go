package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var allPermissions = []string{
	"pxe.login","pxe.install","pxe.capture","pxe.iso","pxe.gold","pxe.clone","pxe.diagnostics",
	"admin.users","admin.groups","admin.pxe","admin.branding",
}

func groupDTO(g UserGroup) UserGroupDTO {
	return UserGroupDTO{ID:g.ID,Name:g.Name,Description:g.Description,Permissions:g.Permissions.Values(),CreatedAt:g.CreatedAt}
}

func (a *App) userPermissions(u *User) map[string]bool {
	out := map[string]bool{}
	if u == nil { return out }
	if strings.EqualFold(u.Role,"admin") {
		for _, p := range allPermissions { out[p]=true }
		return out
	}
	if u.GroupID != nil {
		var g UserGroup
		if a.db.First(&g,"id = ?",*u.GroupID).Error == nil {
			for _, p := range g.Permissions.Values() { out[p]=true }
		}
	}
	return out
}

func (a *App) hasPermission(u *User, permission string) bool {
	return a.userPermissions(u)[permission]
}

func (a *App) requireAdminPermission(w http.ResponseWriter, r *http.Request, permission string) bool {
	u := a.currentUser(r)
	if !a.hasPermission(u,permission) {
		writeJSON(w,http.StatusForbidden,map[string]string{"detail":"permission denied"})
		return false
	}
	return true
}

func (a *App) listAdminUsers(w http.ResponseWriter, r *http.Request) {
	if !a.requireAdminPermission(w,r,"admin.users") { return }
	var users []User
	a.db.Order("username").Find(&users)
	writeJSON(w,200,users)
}

func (a *App) saveAdminUser(w http.ResponseWriter, r *http.Request) {
	if !a.requireAdminPermission(w,r,"admin.users") { return }
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role string `json:"role"`
		GroupID *string `json:"group_id"`
		Disabled bool `json:"disabled"`
	}
	if decodeJSON(r,&input) != nil || strings.TrimSpace(input.Username)=="" {
		writeJSON(w,400,map[string]string{"detail":"username is required"}); return
	}
	id := chi.URLParam(r,"id")
	if id == "" {
		if len(input.Password) < 12 { writeJSON(w,400,map[string]string{"detail":"password must be at least 12 characters"}); return }
		u := User{ID:uuid.NewString(),Username:strings.TrimSpace(input.Username),PasswordHash:hashPassword(input.Password),Role:input.Role,GroupID:input.GroupID,Disabled:input.Disabled,CreatedAt:time.Now().UTC()}
		if u.Role == "" { u.Role="user" }
		if a.db.Create(&u).Error != nil { writeJSON(w,409,map[string]string{"detail":"user could not be created"}); return }
		a.audit(r,"create","user",u.ID,true,"")
		writeJSON(w,200,u); return
	}
	var u User
	if a.db.First(&u,"id = ?",id).Error != nil { writeJSON(w,404,map[string]string{"detail":"user not found"}); return }
	u.Username=strings.TrimSpace(input.Username); u.Role=input.Role; if u.Role=="" {u.Role="user"}; u.GroupID=input.GroupID; u.Disabled=input.Disabled
	if input.Password!="" {
		if len(input.Password)<12 { writeJSON(w,400,map[string]string{"detail":"password must be at least 12 characters"}); return }
		u.PasswordHash=hashPassword(input.Password)
	}
	if a.db.Save(&u).Error != nil { writeJSON(w,400,map[string]string{"detail":"user could not be saved"}); return }
	a.audit(r,"save","user",u.ID,true,"")
	writeJSON(w,200,u)
}

func (a *App) deleteAdminUser(w http.ResponseWriter, r *http.Request) {
	if !a.requireAdminPermission(w,r,"admin.users") { return }
	id:=chi.URLParam(r,"id")
	u:=a.currentUser(r)
	if u!=nil && u.ID==id { writeJSON(w,400,map[string]string{"detail":"you cannot delete your own account"}); return }
	err:=a.db.Delete(&User{},"id = ?",id).Error
	a.audit(r,"delete","user",id,err==nil,"")
	if err!=nil { writeJSON(w,400,map[string]string{"detail":"delete failed"}); return }
	w.WriteHeader(204)
}

func (a *App) listGroups(w http.ResponseWriter, r *http.Request) {
	if !a.requireAdminPermission(w,r,"admin.groups") { return }
	var rows []UserGroup
	a.db.Order("name").Find(&rows)
	out:=make([]UserGroupDTO,0,len(rows))
	for _,g:=range rows { out=append(out,groupDTO(g)) }
	writeJSON(w,200,map[string]any{"groups":out,"available_permissions":allPermissions})
}

func (a *App) saveGroup(w http.ResponseWriter, r *http.Request) {
	if !a.requireAdminPermission(w,r,"admin.groups") { return }
	var input UserGroupDTO
	if decodeJSON(r,&input)!=nil || strings.TrimSpace(input.Name)=="" { writeJSON(w,400,map[string]string{"detail":"group name is required"}); return }
	id:=chi.URLParam(r,"id")
	if id=="" { id=uuid.NewString() }
	valid:=map[string]bool{}; for _,p:=range allPermissions {valid[p]=true}
	perms:=[]string{}; for _,p:=range input.Permissions {if valid[p] {perms=append(perms,p)}}
	g:=UserGroup{ID:id,Name:strings.TrimSpace(input.Name),Description:input.Description,Permissions:list(perms),CreatedAt:time.Now().UTC()}
	if id!=input.ID && input.ID!="" { g.CreatedAt=input.CreatedAt }
	if a.db.Save(&g).Error!=nil { writeJSON(w,409,map[string]string{"detail":"group could not be saved"}); return }
	a.audit(r,"save","group",g.ID,true,"")
	writeJSON(w,200,groupDTO(g))
}

func (a *App) deleteGroup(w http.ResponseWriter, r *http.Request) {
	if !a.requireAdminPermission(w,r,"admin.groups") { return }
	id:=chi.URLParam(r,"id")
	var count int64; a.db.Model(&User{}).Where("group_id = ?",id).Count(&count)
	if count>0 { writeJSON(w,409,map[string]string{"detail":"group is assigned to users"}); return }
	err:=a.db.Delete(&UserGroup{},"id = ?",id).Error
	a.audit(r,"delete","group",id,err==nil,"")
	if err!=nil { writeJSON(w,400,map[string]string{"detail":"delete failed"}); return }
	w.WriteHeader(204)
}

func (a *App) listISOs(w http.ResponseWriter,r *http.Request){var rows []ISOImage;a.db.Order("created_at desc").Find(&rows);writeJSON(w,200,rows)}
func (a *App) saveISO(w http.ResponseWriter,r *http.Request){
	if !a.requireAdminPermission(w,r,"admin.pxe"){return}
	var row ISOImage;if decodeJSON(r,&row)!=nil||strings.TrimSpace(row.Name)==""{writeJSON(w,400,map[string]string{"detail":"name is required"});return}
	if id:=chi.URLParam(r,"id");id!=""{row.ID=id}else{row.ID=uuid.NewString();row.CreatedAt=time.Now().UTC()}
	if row.Architecture==""{row.Architecture="x86_64"};if a.db.Save(&row).Error!=nil{writeJSON(w,409,map[string]string{"detail":"ISO could not be saved"});return};a.audit(r,"save","iso",row.ID,true,"");writeJSON(w,200,row)
}
func (a *App) deleteISO(w http.ResponseWriter,r *http.Request){if !a.requireAdminPermission(w,r,"admin.pxe"){return};id:=chi.URLParam(r,"id");err:=a.db.Delete(&ISOImage{},"id = ?",id).Error;a.audit(r,"delete","iso",id,err==nil,"");if err!=nil{writeJSON(w,400,map[string]string{"detail":"delete failed"});return};w.WriteHeader(204)}

func (a *App) listClones(w http.ResponseWriter,r *http.Request){var rows []CloneImage;a.db.Order("created_at desc").Find(&rows);writeJSON(w,200,rows)}
func (a *App) saveClone(w http.ResponseWriter,r *http.Request){
	if !a.requireAdminPermission(w,r,"admin.pxe"){return}
	var row CloneImage;if decodeJSON(r,&row)!=nil||strings.TrimSpace(row.Name)==""{writeJSON(w,400,map[string]string{"detail":"name is required"});return}
	if id:=chi.URLParam(r,"id");id!=""{row.ID=id}else{row.ID=uuid.NewString();row.CreatedAt=time.Now().UTC()}
	if a.db.Save(&row).Error!=nil{writeJSON(w,409,map[string]string{"detail":"clone image could not be saved"});return};a.audit(r,"save","clone-image",row.ID,true,"");writeJSON(w,200,row)
}
func (a *App) deleteClone(w http.ResponseWriter,r *http.Request){if !a.requireAdminPermission(w,r,"admin.pxe"){return};id:=chi.URLParam(r,"id");err:=a.db.Delete(&CloneImage{},"id = ?",id).Error;a.audit(r,"delete","clone-image",id,err==nil,"");if err!=nil{writeJSON(w,400,map[string]string{"detail":"delete failed"});return};w.WriteHeader(204)}

func (a *App) uploadBrandingAsset(w http.ResponseWriter,r *http.Request){
	if !a.requireAdminPermission(w,r,"admin.branding"){return}
	kind:=chi.URLParam(r,"kind");if kind!="logo"&&kind!="background"{writeJSON(w,400,map[string]string{"detail":"asset must be logo or background"});return}
	if err:=r.ParseMultipartForm(8<<20);err!=nil{writeJSON(w,400,map[string]string{"detail":"invalid upload"});return}
	file,h,err:=r.FormFile("file");if err!=nil{writeJSON(w,400,map[string]string{"detail":"file is required"});return};defer file.Close()
	ct:=h.Header.Get("Content-Type");if !strings.HasPrefix(ct,"image/png")&&!strings.HasPrefix(ct,"image/jpeg")&&!strings.HasPrefix(ct,"image/webp"){writeJSON(w,400,map[string]string{"detail":"PNG, JPEG or WebP required"});return}
	data,err:=io.ReadAll(io.LimitReader(file,8<<20));if err!=nil||len(data)==0{writeJSON(w,400,map[string]string{"detail":"could not read image"});return}
	row:=BrandingAsset{ID:kind,ContentType:ct,Data:data,UpdatedAt:time.Now().UTC()}
	if a.db.Save(&row).Error!=nil{writeJSON(w,500,map[string]string{"detail":"asset could not be saved"});return}
	a.audit(r,"upload","pxe-branding",kind,true,"")
	writeJSON(w,200,map[string]any{"ok":true,"url":"/boot/assets/"+kind,"content_type":ct})
}

func (a *App) getBrandingAsset(w http.ResponseWriter,r *http.Request){
	kind:=chi.URLParam(r,"kind");var row BrandingAsset
	if a.db.First(&row,"id = ?",kind).Error!=nil{http.NotFound(w,r);return}
	w.Header().Set("Content-Type",row.ContentType);w.Header().Set("Cache-Control","public, max-age=60");w.Write(row.Data)
}

func (a *App) basicPXEUser(r *http.Request) *User {
	username,password,ok:=r.BasicAuth();if !ok{return nil}
	var u User;if a.db.First(&u,"username = ? AND disabled = ?",username,false).Error!=nil||!verifyPassword(u.PasswordHash,password){return nil}
	return &u
}

func (a *App) pxeCatalog(w http.ResponseWriter,r *http.Request){
	u:=a.basicPXEUser(r)
	if u==nil||!a.hasPermission(u,"pxe.login"){
		w.Header().Set("WWW-Authenticate",`Basic realm="ReForge PXE"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	token:=a.newPXEAccessToken(u)
	perms:=a.userPermissions(u)
	w.Header().Set("Content-Type","text/plain")
	fmt.Fprintln(w,"#!ipxe")
	fmt.Fprintln(w,"menu ReForge Deployment Portal")
	if perms["pxe.install"] {
		if perms["pxe.iso"] { fmt.Fprintln(w,"item install-iso Install from ISO") }
		if perms["pxe.gold"] { fmt.Fprintln(w,"item install-gold Deploy Gold Image") }
		if perms["pxe.clone"] { fmt.Fprintln(w,"item install-clone Deploy Clone Image") }
	}
	if perms["pxe.capture"] {
		if perms["pxe.gold"] { fmt.Fprintln(w,"item capture-gold Capture as Gold Image") }
		if perms["pxe.clone"] { fmt.Fprintln(w,"item capture-clone Capture as Clone Image") }
	}
	if perms["pxe.diagnostics"] { fmt.Fprintln(w,"item diagnostics Diagnostics") }
	fmt.Fprintln(w,"item local Boot local disk")
	fmt.Fprintln(w,"choose action && goto ${action}")
	if perms["pxe.install"]&&perms["pxe.iso"] {
		fmt.Fprintln(w,":install-iso")
		fmt.Fprintf(w,"chain %s/boot/sources.ipxe?type=iso&token=%s&mac=${net0/mac}&uuid=${uuid} || goto local\n",a.pxeServerURL(),token)
	}
	if perms["pxe.install"]&&perms["pxe.gold"] {
		fmt.Fprintln(w,":install-gold")
		fmt.Fprintf(w,"chain %s/boot/sources.ipxe?type=gold&token=%s&mac=${net0/mac}&uuid=${uuid} || goto local\n",a.pxeServerURL(),token)
	}
	if perms["pxe.install"]&&perms["pxe.clone"] {
		fmt.Fprintln(w,":install-clone")
		fmt.Fprintf(w,"chain %s/boot/sources.ipxe?type=clone&token=%s&mac=${net0/mac}&uuid=${uuid} || goto local\n",a.pxeServerURL(),token)
	}
	if perms["pxe.capture"]&&perms["pxe.gold"] {
		fmt.Fprintln(w,":capture-gold")
		fmt.Fprintf(w,"chain %s/boot/action.ipxe?action=capture&type=gold&token=%s&mac=${net0/mac}&uuid=${uuid} || goto local\n",a.pxeServerURL(),token)
	}
	if perms["pxe.capture"]&&perms["pxe.clone"] {
		fmt.Fprintln(w,":capture-clone")
		fmt.Fprintf(w,"chain %s/boot/action.ipxe?action=capture&type=clone&token=%s&mac=${net0/mac}&uuid=${uuid} || goto local\n",a.pxeServerURL(),token)
	}
	if perms["pxe.diagnostics"] {
		fmt.Fprintln(w,":diagnostics")
		fmt.Fprintln(w,"echo ReForge diagnostics")
		fmt.Fprintln(w,"sleep 2")
		fmt.Fprintln(w,"goto local")
	}
	fmt.Fprintln(w,":local")
	fmt.Fprintln(w,"exit")
}

func (a *App) pxeServerURL() string {
	var p PXEConfig
	if a.db.First(&p,"id = ?","default").Error==nil&&strings.TrimSpace(p.ServerURL)!="" {
		return strings.TrimRight(p.ServerURL,"/")
	}
	return "http://reforge.local:5173"
}

func (a *App) pxeSources(w http.ResponseWriter,r *http.Request){
	u:=a.pxeUserFromToken(r)
	if u==nil||!a.hasPermission(u,"pxe.login"){w.WriteHeader(http.StatusUnauthorized);return}
	token:=r.URL.Query().Get("token")
	typ:=r.URL.Query().Get("type")
	needed:="pxe."+typ
	if !a.hasPermission(u,"pxe.install")||!a.hasPermission(u,needed){w.WriteHeader(http.StatusForbidden);return}
	var p PXEConfig
	a.db.First(&p,"id = ?","default")
	title:=strings.TrimSpace(p.InstallTitle);if title==""{title="Install operating system"}
	subtitle:=strings.TrimSpace(p.InstallSubtitle)
	w.Header().Set("Content-Type","text/plain")
	fmt.Fprintln(w,"#!ipxe")
	if p.ShowBackground { fmt.Fprintln(w,"console --picture "+a.pxeServerURL()+"/boot/assets/background ||") }
	if strings.TrimSpace(p.BrandName)!="" { fmt.Fprintln(w,"echo "+p.BrandName) }
	fmt.Fprintln(w,"echo "+title)
	if subtitle!="" { fmt.Fprintln(w,"echo "+subtitle) }
	fmt.Fprintf(w,"menu Select %s source\n",strings.ToUpper(typ))
	count:=0
	switch typ{
	case "iso":
		var rows []ISOImage
		a.db.Where("enabled = ?",true).Order("name").Find(&rows)
		for _,x:=range rows{fmt.Fprintf(w,"item %s %s %s\n",x.ID,x.Name,x.Version);count++}
	case "gold":
		var rows []GoldImage
		a.db.Order("name").Find(&rows)
		for _,x:=range rows{fmt.Fprintf(w,"item %s %s v%s\n",x.ID,x.Name,x.ImageVersion);count++}
	case "clone":
		var rows []CloneImage
		a.db.Where("enabled = ?",true).Order("name").Find(&rows)
		for _,x:=range rows{fmt.Fprintf(w,"item %s %s\n",x.ID,x.Name);count++}
	default:
		w.WriteHeader(http.StatusBadRequest);return
	}
	if count==0 { fmt.Fprintln(w,"item none No deployment sources available") }
	fmt.Fprintln(w,"item cancel Cancel")
	fmt.Fprintln(w,"choose source && goto selected")
	fmt.Fprintln(w,":selected")
	fmt.Fprintln(w,"iseq ${source} cancel && exit ||")
	fmt.Fprintln(w,"iseq ${source} none && exit ||")
	fmt.Fprintf(w,"chain %s/boot/action.ipxe?action=install&type=%s&source=${source}&token=%s&mac=${net0/mac}&uuid=${uuid} || exit\n",a.pxeServerURL(),typ,token)
}

func (a *App) pxeAction(w http.ResponseWriter,r *http.Request){
	u:=a.pxeUserFromToken(r)
	if u==nil||!a.hasPermission(u,"pxe.login"){w.WriteHeader(http.StatusUnauthorized);return}
	action:=r.URL.Query().Get("action")
	typ:=r.URL.Query().Get("type")
	source:=r.URL.Query().Get("source")
	if action=="install"&&!a.hasPermission(u,"pxe.install"){w.WriteHeader(http.StatusForbidden);return}
	if action=="capture"&&!a.hasPermission(u,"pxe.capture"){w.WriteHeader(http.StatusForbidden);return}
	if !a.hasPermission(u,"pxe."+typ){w.WriteHeader(http.StatusForbidden);return}
	macRaw:=strings.TrimSpace(r.URL.Query().Get("mac"))
	uuidRaw:=strings.TrimSpace(r.URL.Query().Get("uuid"))
	mac:=""
	if macRaw!="" {
		if normalized,err:=normalizeMAC(macRaw);err==nil { mac=normalized }
	}
	if mac==""&&uuidRaw==""{w.WriteHeader(http.StatusBadRequest);return}
	task:=PXETask{ID:uuid.NewString(),HostMAC:mac,HostUUID:uuidRaw,Action:action,SourceType:typ,SourceID:source,RequestedBy:u.Username,Status:"queued",CreatedAt:time.Now().UTC()}
	if a.db.Create(&task).Error!=nil{w.WriteHeader(http.StatusInternalServerError);return}
	a.audit(r,"queue-"+action,"pxe-task",task.ID,true,typ)
	var p PXEConfig
	a.db.First(&p,"id = ?","default")
	w.Header().Set("Content-Type","text/plain")
	fmt.Fprintln(w,"#!ipxe")
	if p.ShowBackground { fmt.Fprintln(w,"console --picture "+a.pxeServerURL()+"/boot/assets/background ||") }
	if strings.TrimSpace(p.BrandName)!="" { fmt.Fprintln(w,"echo "+p.BrandName) }
	if action=="capture" {
		title:=strings.TrimSpace(p.CaptureTitle);if title==""{title="Capture image"}
		fmt.Fprintln(w,"echo "+title)
		if strings.TrimSpace(p.CaptureSubtitle)!="" { fmt.Fprintln(w,"echo "+p.CaptureSubtitle) }
	} else {
		title:=strings.TrimSpace(p.LoadingTitle);if title==""{title="ReForge is preparing this device"}
		fmt.Fprintln(w,"echo "+title)
		if strings.TrimSpace(p.LoadingMessage)!="" { fmt.Fprintln(w,"echo "+p.LoadingMessage) }
	}
	fmt.Fprintln(w,"echo")
	fmt.Fprintln(w,"echo Task: "+task.ID)
	fmt.Fprintln(w,"echo Requested by: "+u.Username)
	fmt.Fprintln(w,"echo Waiting for authorized imaging node...")
	fmt.Fprintln(w,"sleep 3")
	fmt.Fprintln(w,"exit")
}

func decodeJSON(r *http.Request,v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (a *App) newPXEAccessToken(u *User) string {
	token:=randomToken(32)
	_ = a.db.Create(&PXEAccessToken{Token:token,UserID:u.ID,ExpiresAt:time.Now().UTC().Add(10*time.Minute),CreatedAt:time.Now().UTC()}).Error
	return token
}

func (a *App) pxeUserFromToken(r *http.Request) *User {
	token:=strings.TrimSpace(r.URL.Query().Get("token"))
	if token=="" { return nil }
	var t PXEAccessToken
	if a.db.First(&t,"token = ? AND expires_at > ?",token,time.Now().UTC()).Error!=nil { return nil }
	var u User
	if a.db.First(&u,"id = ? AND disabled = ?",t.UserID,false).Error!=nil { return nil }
	return &u
}
