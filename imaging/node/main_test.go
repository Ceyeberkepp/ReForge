package main

import (
 "bytes"
 "os"
 "path/filepath"
 "strings"
 "testing"
)

func TestCaptureVerifyRestore(t *testing.T) {
 dir := t.TempDir()
 src, image, target := filepath.Join(dir, "source.bin"), filepath.Join(dir, "image.raw"), filepath.Join(dir, "clone.bin")
 original := bytes.Repeat([]byte("ReForge image data\x00"), 4096)
 if err := os.WriteFile(src, original, 0600); err != nil { t.Fatal(err) }
 if err := capture(src, image); err != nil { t.Fatal(err) }
 if _, err := verify(image); err != nil { t.Fatal(err) }
 if err := restore(image, target); err != nil { t.Fatal(err) }
 copied, err := os.ReadFile(target)
 if err != nil { t.Fatal(err) }
 if !bytes.Equal(original, copied) { t.Fatal("restore differs from source") }
 if err := restore(image, target); err == nil { t.Fatal("existing target was overwritten") }
}

func TestTamperingFailsBeforeRestore(t *testing.T) {
 dir := t.TempDir()
 src, image, target := filepath.Join(dir, "src"), filepath.Join(dir, "image"), filepath.Join(dir, "target")
 os.WriteFile(src, []byte("original"), 0600)
 if err := capture(src, image); err != nil { t.Fatal(err) }
 if err := os.WriteFile(image, []byte("tampered"), 0600); err != nil { t.Fatal(err) }
 if _, err := verify(image); err == nil { t.Fatal("tampering not detected") }
 if err := restore(image, target); err == nil { t.Fatal("tampered image restored") }
 if _, err := os.Stat(target); !os.IsNotExist(err) { t.Fatal("target created despite integrity failure") }
}

func TestRejectsSymlinkAndExistingOutput(t *testing.T) {
 dir := t.TempDir()
 src, link, image := filepath.Join(dir, "src"), filepath.Join(dir, "link"), filepath.Join(dir, "image")
 os.WriteFile(src, []byte("data"), 0600)
 if err := os.Symlink(src, link); err != nil { t.Skip(err) }
 if err := capture(link, image); err == nil { t.Fatal("symlink source accepted") }
 if err := capture(src, image); err != nil { t.Fatal(err) }
 if err := capture(src, image); err == nil { t.Fatal("existing destination overwritten") }
 if err := restore(image, link); err == nil { t.Fatal("symlink destination overwritten") }
 if b, _ := os.ReadFile(src); !strings.EqualFold(string(b), "data") { t.Fatal("original modified") }
}

func TestRejectsDeviceAndMissingManifest(t *testing.T) {
 if _, err := os.Stat("/dev/null"); err == nil {
  if _, _, err := regularInput("/dev/null"); err == nil { t.Fatal("device accepted") }
 }
 dir := t.TempDir()
 path := filepath.Join(dir, "unverified.img")
 os.WriteFile(path, []byte("raw"), 0600)
 if _, err := verify(path); err == nil { t.Fatal("missing manifest accepted") }
}

func TestEmptyImage(t *testing.T) {
 dir := t.TempDir()
 src, image, target := filepath.Join(dir, "empty"), filepath.Join(dir, "image"), filepath.Join(dir, "target")
 os.WriteFile(src, nil, 0600)
 if err := capture(src, image); err != nil { t.Fatal(err) }
 if err := restore(image, target); err != nil { t.Fatal(err) }
 if b, err := os.ReadFile(target); err != nil || len(b) != 0 { t.Fatal("empty image roundtrip failed") }
}

func TestRejectsAmbiguousManifestsBeforeRestore(t *testing.T) {
 dir := t.TempDir()
 src, image := filepath.Join(dir, "source"), filepath.Join(dir, "image")
 if err := os.WriteFile(src, []byte("recovery data"), 0600); err != nil { t.Fatal(err) }
 if err := capture(src, image); err != nil { t.Fatal(err) }
 valid, err := os.ReadFile(image + ".json")
 if err != nil { t.Fatal(err) }
 cases := map[string][]byte{
  "trailing-object": append(append([]byte{}, valid...), []byte("{}")...),
  "trailing-garbage": append(append([]byte{}, valid...), []byte("garbage")...),
  "unknown-field": bytes.Replace(valid, []byte("{"), []byte("{\"untrusted\":true,"), 1),
  "duplicate-size": bytes.Replace(valid, []byte("\"size\":"), []byte("\"size\":0,\"size\":"), 1),
  "casefold-duplicate-size": bytes.Replace(valid, []byte("\"size\":"), []byte("\"SIZE\":0,\"size\":"), 1),
  "casefold-duplicate-sha256": bytes.Replace(valid, []byte("\"sha256\":"), []byte("\"SHA256\":\"ignored\",\"sha256\":"), 1),
  "duplicate-sha256": bytes.Replace(valid, []byte("\"sha256\":"), []byte("\"sha256\":\"ignored\",\"sha256\":"), 1),
  "oversized": append(append([]byte{}, valid...), bytes.Repeat([]byte(" "), 4097)...),
 }
 for name, malformed := range cases {
  t.Run(name, func(t *testing.T) {
   if err := os.WriteFile(image + ".json", malformed, 0600); err != nil { t.Fatal(err) }
   target := filepath.Join(dir, name+"-target")
   if err := restore(image, target); err == nil { t.Fatal("ambiguous manifest accepted") }
   if _, err := os.Lstat(target); !os.IsNotExist(err) { t.Fatal("target created for invalid manifest") }
  })
 }
}


func TestRejectsSymlinkManifestBeforeRestore(t *testing.T) {
 dir := t.TempDir()
 src := filepath.Join(dir, "source")
 image := filepath.Join(dir, "image")
 target := filepath.Join(dir, "target")
 if err := os.WriteFile(src, []byte("trusted bytes"), 0600); err != nil { t.Fatal(err) }
 if err := capture(src, image); err != nil { t.Fatal(err) }
 manifest := image + ".json"
 originalManifest := filepath.Join(dir, "saved-manifest.json")
 if err := os.Rename(manifest, originalManifest); err != nil { t.Fatal(err) }
 if err := os.Symlink(originalManifest, manifest); err != nil { t.Skip(err) }
 if err := restore(image, target); err == nil { t.Fatal("symlinked manifest accepted") }
 if _, err := os.Lstat(target); !os.IsNotExist(err) { t.Fatal("target created for symlinked manifest") }
}

func TestExistingRestoreTargetPreservesContents(t *testing.T) {
 dir := t.TempDir()
 src := filepath.Join(dir, "source")
 image := filepath.Join(dir, "image")
 target := filepath.Join(dir, "existing-target")
 if err := os.WriteFile(src, []byte("image bytes"), 0600); err != nil { t.Fatal(err) }
 if err := capture(src, image); err != nil { t.Fatal(err) }
 original := []byte("do not overwrite")
 if err := os.WriteFile(target, original, 0600); err != nil { t.Fatal(err) }
 if err := restore(image, target); err == nil { t.Fatal("existing target accepted") }
 after, err := os.ReadFile(target)
 if err != nil { t.Fatal(err) }
 if !bytes.Equal(after, original) { t.Fatal("existing target contents changed") }
}
