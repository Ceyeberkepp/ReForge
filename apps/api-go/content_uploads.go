package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const maxUploadBytes int64 = 64 << 30 // 64 GiB

var safeFileName = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func sanitizeUploadName(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	name = safeFileName.ReplaceAllString(name, "_")
	name = strings.Trim(name, "._")
	if name == "" { name = "upload.bin" }
	return name
}

func allowedUploadExt(kind, name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch kind {
	case "gold":
		return map[string]bool{".img":true,".wim":true,".esd":true,".gz":true,".zst":true,".zip":true}[ext]
	case "iso":
		return ext == ".iso"
	case "driver":
		return map[string]bool{".zip":true,".cab":true,".7z":true}[ext]
	case "software":
		return map[string]bool{".msi":true,".exe":true,".msix":true,".appx":true,".pkg":true,".zip":true,".ps1":true,".sh":true}[ext]
	}
	return false
}

func (a *App) ensureContentRoot() error {
	for _, d := range []string{"gold","iso","drivers","software"} {
		if err := os.MkdirAll(filepath.Join(a.cfg.ContentRoot,d), 0o750); err != nil { return err }
	}
	return nil
}

func firstFilePart(r *http.Request) (*multipart.Part,error) {
	mr,err:=r.MultipartReader()
	if err!=nil{return nil,err}
	for {
		p,err:=mr.NextPart()
		if err!=nil{return nil,err}
		if p.FileName()!="" { return p,nil }
		p.Close()
	}
}

func (a *App) streamUpload(w http.ResponseWriter,r *http.Request,kind,id string) (string,string,int64,bool) {
	if !a.requireAdminPermission(w,r,"admin.pxe"){return "","",0,false}
	if err:=a.ensureContentRoot();err!=nil{writeJSON(w,500,map[string]string{"detail":"content storage unavailable"});return "","",0,false}
	r.Body=http.MaxBytesReader(w,r.Body,maxUploadBytes)
	part,err:=firstFilePart(r)
	if err!=nil{writeJSON(w,400,map[string]string{"detail":"multipart file upload required"});return "","",0,false}
	defer part.Close()
	name:=sanitizeUploadName(part.FileName())
	if !allowedUploadExt(kind,name){writeJSON(w,400,map[string]string{"detail":"unsupported file type for "+kind});return "","",0,false}
	dir:=filepath.Join(a.cfg.ContentRoot,kind)
	finalName:=id+"-"+name
	finalPath:=filepath.Join(dir,finalName)
	tmpPath:=finalPath+".part-"+uuid.NewString()
	out,err:=os.OpenFile(tmpPath,os.O_CREATE|os.O_WRONLY|os.O_EXCL,0o640)
	if err!=nil{writeJSON(w,500,map[string]string{"detail":"unable to create upload file"});return "","",0,false}
	hash:=sha256.New()
	n,copyErr:=io.Copy(io.MultiWriter(out,hash),part)
	closeErr:=out.Close()
	if copyErr!=nil||closeErr!=nil{_ = os.Remove(tmpPath);writeJSON(w,500,map[string]string{"detail":"upload failed"});return "","",0,false}
	if err:=os.Rename(tmpPath,finalPath);err!=nil{_ = os.Remove(tmpPath);writeJSON(w,500,map[string]string{"detail":"unable to finalize upload"});return "","",0,false}
	return finalPath,hex.EncodeToString(hash.Sum(nil)),n,true
}

func (a *App) uploadGoldImage(w http.ResponseWriter,r *http.Request){
	id:=chi.URLParam(r,"id")
	var row GoldImage
	if a.db.First(&row,"id = ?",id).Error!=nil{writeJSON(w,404,map[string]string{"detail":"gold image not found"});return}
	path,sum,size,ok:=a.streamUpload(w,r,"gold",id);if !ok{return}
	row.ImagePath=path;row.Checksum=sum
	if a.db.Save(&row).Error!=nil{writeJSON(w,500,map[string]string{"detail":"metadata update failed"});return}
	a.audit(r,"upload","gold-image",id,true,fmt.Sprintf("%d bytes",size))
	writeJSON(w,200,map[string]any{"path":path,"checksum":sum,"size":size,"image":row})
}

func (a *App) uploadISO(w http.ResponseWriter,r *http.Request){
	id:=chi.URLParam(r,"id")
	var row ISOImage
	if a.db.First(&row,"id = ?",id).Error!=nil{writeJSON(w,404,map[string]string{"detail":"ISO not found"});return}
	path,sum,size,ok:=a.streamUpload(w,r,"iso",id);if !ok{return}
	row.SourcePath=path;row.Checksum=sum
	if a.db.Save(&row).Error!=nil{writeJSON(w,500,map[string]string{"detail":"metadata update failed"});return}
	a.audit(r,"upload","iso",id,true,fmt.Sprintf("%d bytes",size))
	writeJSON(w,200,map[string]any{"path":path,"checksum":sum,"size":size,"iso":row})
}

func (a *App) listDrivers(w http.ResponseWriter,r *http.Request){var rows []DriverPack;a.db.Order("vendor,model,name").Find(&rows);writeJSON(w,200,rows)}
func (a *App) saveDriver(w http.ResponseWriter,r *http.Request){
	if !a.requireAdminPermission(w,r,"admin.pxe"){return}
	var row DriverPack
	if decodeJSON(r,&row)!=nil||strings.TrimSpace(row.Name)==""{writeJSON(w,400,map[string]string{"detail":"name is required"});return}
	if id:=chi.URLParam(r,"id");id!=""{row.ID=id}else{row.ID=uuid.NewString();row.CreatedAt=time.Now().UTC()}
	if row.Architecture==""{row.Architecture="x86_64"}
	if a.db.Save(&row).Error!=nil{writeJSON(w,409,map[string]string{"detail":"driver pack could not be saved"});return}
	a.audit(r,"save","driver-pack",row.ID,true,"")
	writeJSON(w,200,row)
}
func (a *App) deleteDriver(w http.ResponseWriter,r *http.Request){
	if !a.requireAdminPermission(w,r,"admin.pxe"){return}
	id:=chi.URLParam(r,"id");err:=a.db.Delete(&DriverPack{},"id = ?",id).Error
	a.audit(r,"delete","driver-pack",id,err==nil,"");if err!=nil{writeJSON(w,400,map[string]string{"detail":"delete failed"});return};w.WriteHeader(204)
}
func (a *App) uploadDriver(w http.ResponseWriter,r *http.Request){
	id:=chi.URLParam(r,"id");var row DriverPack
	if a.db.First(&row,"id = ?",id).Error!=nil{writeJSON(w,404,map[string]string{"detail":"driver pack not found"});return}
	path,sum,size,ok:=a.streamUpload(w,r,"driver",id);if !ok{return}
	row.PackagePath=path;row.Checksum=sum
	if a.db.Save(&row).Error!=nil{writeJSON(w,500,map[string]string{"detail":"metadata update failed"});return}
	a.audit(r,"upload","driver-pack",id,true,fmt.Sprintf("%d bytes",size))
	writeJSON(w,200,map[string]any{"path":path,"checksum":sum,"size":size,"driver":row})
}

func (a *App) uploadSoftware(w http.ResponseWriter,r *http.Request){
	id:=chi.URLParam(r,"id");var row SoftwarePackage
	if a.db.First(&row,"id = ?",id).Error!=nil{writeJSON(w,404,map[string]string{"detail":"software package not found"});return}
	path,sum,size,ok:=a.streamUpload(w,r,"software",id);if !ok{return}
	row.InstallerPath=path
	if a.db.Save(&row).Error!=nil{writeJSON(w,500,map[string]string{"detail":"metadata update failed"});return}
	a.audit(r,"upload","software",id,true,fmt.Sprintf("%d bytes sha256=%s",size,sum))
	writeJSON(w,200,map[string]any{"path":path,"checksum":sum,"size":size,"software":row})
}
