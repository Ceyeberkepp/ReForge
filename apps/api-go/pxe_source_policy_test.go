package main

import "testing"

func TestValidPXEActionType(t *testing.T) {
 cases := []struct{a, typ, id string; ok bool}{
  {"install","gold","img1",true},
  {"install","iso","iso1",true},
  {"install","clone","clone1",true},
  {"install","clone","",false},
  {"install","other","img1",false},
  {"capture","clone","",true},
  {"capture","gold","",true},
  {"capture","iso","",false},
  {"capture","clone","some-id",false},
  {"wipe","clone","clone1",false},
 }
 for _, c:=range cases { if got:=validPXEActionType(c.a,c.typ,c.id);got!=c.ok {t.Errorf("%q/%q/%q got %t want %t",c.a,c.typ,c.id,got,c.ok)} }
}

func TestCloneOnlyMatchesOriginalHost(t *testing.T) {
 sourceID:="host-one"
 clone:=CloneImage{SourceHostID:&sourceID}
 if !cloneMatchesHost(clone,Host{ID:"host-one"}) {t.Fatal("source host rejected")}
 if cloneMatchesHost(clone,Host{ID:"host-two"}) {t.Fatal("cross-host clone permitted")}
 if cloneMatchesHost(CloneImage{},Host{ID:"host-one"}) {t.Fatal("unbound clone permitted")}
}
