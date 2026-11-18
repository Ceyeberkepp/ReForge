package main

// PXE requests must only create recognized tasks. An absent source is
// supported for capture, but all restore/install operations require an
// explicit source ID.
func validPXEActionType(action, sourceType, sourceID string) bool {
 switch action {
 case "install":
  if sourceID == "" { return false }
  return sourceType == "iso" || sourceType == "gold" || sourceType == "clone"
 case "capture":
  if sourceID != "" { return false }
  return sourceType == "gold" || sourceType == "clone"
 default:
  return false
 }
}

// Clones retain the original OS identity. They can only be selected for the
// enrolled host they were captured from, never generalized to another host.
func cloneMatchesHost(clone CloneImage, host Host) bool {
 return clone.SourceHostID != nil && host.ID != "" && *clone.SourceHostID == host.ID
}
