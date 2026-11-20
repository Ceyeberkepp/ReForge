// reforge-imaging-node provides fail-closed, file-backed capture and restore.
// Physical block devices are intentionally unsupported until signed deployment
// authorizations are verified by a dedicated imaging-node service.
package main

import (
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "flag"
 "fmt"
 "io"
 "os"
 "path/filepath"
)

type Manifest struct {
 Format string `json:"format"`
 Size int64 `json:"size"`
 SHA256 string `json:"sha256"`
}

const imageFormat = "reforge-raw-v1"

// regularInput rejects symlinks, device nodes, directories and other special files.
func regularInput(path string) (*os.File, int64, error) {
 info, err := os.Lstat(path)
 if err != nil { return nil, 0, err }
 if !info.Mode().IsRegular() { return nil, 0, errors.New("source must be a regular file (block devices and symlinks prohibited)") }
 f, err := os.Open(path)
 if err != nil { return nil, 0, err }
 actual, err := f.Stat()
 if err != nil || !actual.Mode().IsRegular() || !os.SameFile(info, actual) {
  f.Close()
  return nil, 0, errors.New("source changed during open")
 }
 return f, actual.Size(), nil
}

func newDestination(path string) (*os.File, error) {
 // Never follow a destination symlink or overwrite an existing file.
 return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
}

func transfer(src, dst string, expected *Manifest) (Manifest, error) {
 var result Manifest
 if src == "" || dst == "" { return result, errors.New("source and destination are required") }
 in, size, err := regularInput(src)
 if err != nil { return result, err }
 defer in.Close()
 // Existing output, symlink and special-file paths all fail with O_EXCL.
 out, err := newDestination(dst)
 if err != nil { return result, err }
 success := false
 defer func() {
  out.Close()
  if !success { os.Remove(dst) }
 }()
 hasher := sha256.New()
 n, err := io.Copy(io.MultiWriter(out, hasher), in)
 if err != nil { return result, err }
 if n != size { return result, errors.New("source size changed during transfer") }
 result = Manifest{Format: imageFormat, Size: n, SHA256: hex.EncodeToString(hasher.Sum(nil))}
 if expected != nil && (expected.Format != imageFormat || expected.Size != n || expected.SHA256 != result.SHA256) {
  return result, errors.New("copied data does not match image manifest")
 }
 if err := out.Sync(); err != nil { return result, err }
 if err := out.Close(); err != nil { return result, err }
 success = true
 return result, nil
}

func capture(src, dst string) error {
 if _, err := os.Lstat(dst + ".json"); err == nil { return errors.New("manifest already exists") } else if !os.IsNotExist(err) { return err }
 manifest, err := transfer(src, dst, nil)
 if err != nil { return err }
 manifestCreated := false
 data, err := json.MarshalIndent(manifest, "", "  ")
 if err == nil {
  var file *os.File
  file, err = newDestination(dst + ".json")
  if err == nil {
   manifestCreated = true
   _, err = file.Write(append(data, '\n'))
   if err == nil { err = file.Sync() }
   if cerr := file.Close(); err == nil { err = cerr }
  }
 }
 if err != nil {
  os.Remove(dst)
  if manifestCreated { os.Remove(dst + ".json") }
  return err
 }
 return nil
}

func readManifest(path string) (Manifest, error) {
 var m Manifest
 f, size, err := regularInput(path)
 if err != nil { return m, err }
 defer f.Close()
 if size > 4096 { return m, errors.New("manifest exceeds size limit") }
 // Reject duplicate object keys: encoding/json otherwise silently accepts the
 // last value, making integrity metadata ambiguous across parsers.
 keyDecoder := json.NewDecoder(f)
 opening, err := keyDecoder.Token()
 if err != nil || opening != json.Delim('{') { return m, errors.New("manifest must be an object") }
 seen := make(map[string]bool)
 for keyDecoder.More() {
  token, err := keyDecoder.Token()
  if err != nil { return m, err }
  key, ok := token.(string)
  if !ok || seen[key] { return m, errors.New("manifest contains duplicate or invalid key") }
  seen[key] = true
  var value json.RawMessage
  if err := keyDecoder.Decode(&value); err != nil { return m, err }
 }
 if _, err := keyDecoder.Token(); err != nil { return m, err }
 if _, err := f.Seek(0, io.SeekStart); err != nil { return m, err }
 decoder := json.NewDecoder(f)
 decoder.DisallowUnknownFields()
 if err := decoder.Decode(&m); err != nil { return m, err }
 // A manifest is exactly one JSON object. Reject appended JSON, garbage and
 // oversized metadata rather than accepting an ambiguous integrity record.
 var trailing any
 if err := decoder.Decode(&trailing); err != io.EOF { return m, errors.New("manifest contains trailing data or exceeds size limit") }
 if m.Format != imageFormat || m.Size < 0 || len(m.SHA256) != 64 { return m, errors.New("invalid manifest") }
 if _, err := hex.DecodeString(m.SHA256); err != nil { return m, err }
 return m, nil
}

func verify(image string) (Manifest, error) {
 m, err := readManifest(image + ".json")
 if err != nil { return m, err }
 f, size, err := regularInput(image)
 if err != nil { return m, err }
 defer f.Close()
 if size != m.Size { return m, errors.New("image size mismatch") }
 h := sha256.New()
 n, err := io.Copy(h, f)
 if err != nil { return m, err }
 if n != m.Size || hex.EncodeToString(h.Sum(nil)) != m.SHA256 {
  return m, errors.New("image SHA-256 verification failed")
 }
 return m, nil
}

func restore(image, target string) error {
 // Verify before creating any target to avoid writing an untrusted image.
 m, err := verify(image)
 if err != nil { return err }
 _, err = transfer(image, target, &m)
 return err
}

func main() {
 operation := flag.String("action", "", "capture, verify or restore")
 source := flag.String("source", "", "regular-file input/image")
 destination := flag.String("destination", "", "new regular-file output")
 flag.Parse()
 var err error
 switch *operation {
 case "capture":
  err = capture(*source, *destination)
 case "verify":
  var m Manifest
  m, err = verify(*source)
  if err == nil { fmt.Printf("verified %s (%d bytes, sha256 %s)\n", filepath.Base(*source), m.Size, m.SHA256) }
 case "restore":
  err = restore(*source, *destination)
 default:
  err = errors.New("action must be capture, verify or restore")
 }
 if err != nil {
  fmt.Fprintln(os.Stderr, "reforge imaging:", err)
  os.Exit(1)
 }
}
