package main

import (
 "strings"
 "testing"
)

func TestValidateFreshDeploymentImage(t *testing.T) {
 tests := []struct{
  name string
  image GoldImage
  allowed bool
 }{
  {"generalized Windows 11", GoldImage{OSName:"Windows 11", SysprepReady:true}, true},
  {"ungeneralized Windows 11", GoldImage{OSName:"Windows 11"}, false},
  {"ungeneralized Windows Server", GoldImage{OSName:"Microsoft Windows Server 2025"}, false},
  {"generalized Windows Server", GoldImage{OSName:"Windows Server 2025", SysprepReady:true}, true},
  {"Linux image", GoldImage{OSName:"Ubuntu Linux"}, true},
 }
 for _, tt := range tests {
  t.Run(tt.name,func(t *testing.T) {
   err:=validateFreshDeploymentImage(tt.image)
   if (err==nil)!=tt.allowed {t.Fatalf("allowed=%t, error=%v",tt.allowed,err)}
   if err!=nil && !strings.Contains(err.Error(),"Sysprep") {t.Fatal(err)}
  })
 }
}
