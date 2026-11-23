package main

import (
 "os"
 "path/filepath"
 "testing"
)

// A manifest is not sufficient authorization to follow a redirected image path.
func TestRestoreRejectsSymlinkedImageBeforeTargetCreation(t *testing.T) {
 dir := t.TempDir()
 src := filepath.Join(dir, "source")
 image := filepath.Join(dir, "image")
 target := filepath.Join(dir, "target")
 if err := os.WriteFile(src, []byte("trusted image bytes"), 0600); err != nil { t.Fatal(err) }
 if err := capture(src, image); err != nil { t.Fatal(err) }
 saved := filepath.Join(dir, "saved-image")
 if err := os.Rename(image, saved); err != nil { t.Fatal(err) }
 if err := os.Symlink(saved, image); err != nil { t.Skip(err) }
 if err := restore(image, target); err == nil { t.Fatal("symlinked image accepted") }
 if _, err := os.Lstat(target); !os.IsNotExist(err) {
  t.Fatalf("restore created target for symlinked image: %v", err)
 }
}
