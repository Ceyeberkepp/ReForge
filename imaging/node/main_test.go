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
