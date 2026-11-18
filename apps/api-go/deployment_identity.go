package main

import (
 "errors"
 "strings"
)

// Fresh Windows deployments must come from a generalized reference image.
// Sysprep /generalize removes machine-specific identity; Windows specializes
// the deployed instance on its first boot. Clone backups are NOT gold images.
func validateFreshDeploymentImage(image GoldImage) error {
 osName := strings.ToLower(strings.TrimSpace(image.OSName))
 if strings.Contains(osName, "windows") && !image.SysprepReady {
  return errors.New("Windows reimaging requires a Sysprep /generalize prepared gold image; raw clones preserve the original device identity and must only be used for same-device recovery")
 }
 return nil
}
