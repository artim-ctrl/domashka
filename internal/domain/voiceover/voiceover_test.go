package voiceover

import "testing"

func TestNewCatalog(t *testing.T) {
	cases := map[string]struct {
		voices []Voice
		ok     bool
	}{
		"valid":          {[]Voice{{ID: "a", Engine: EngineSay}, {ID: "b", Engine: EngineSilero}}, true},
		"empty":          {nil, false},
		"unknown engine": {[]Voice{{ID: "a", Engine: "espeak"}}, false},
		"duplicate":      {[]Voice{{ID: "a", Engine: EngineSay}, {ID: "a", Engine: EngineSay}}, false},
	}
	for name := range cases {
		_, err := NewCatalog(cases[name].voices)
		if (err == nil) != cases[name].ok {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestCatalogByID(t *testing.T) {
	catalog, err := NewCatalog([]Voice{{ID: "a", Engine: EngineSay}})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := catalog.ByID("a"); !ok {
		t.Error("ByID did not find a listed voice")
	}
	if _, ok := catalog.ByID("b"); ok {
		t.Error("ByID found an unknown voice")
	}
}
