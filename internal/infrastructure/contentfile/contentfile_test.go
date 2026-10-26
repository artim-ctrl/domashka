package contentfile

import "testing"

func TestLoad(t *testing.T) {
	if _, err := Load("../../../content/page.json", "../../../content/speech.json"); err != nil {
		t.Fatal(err)
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load("missing.json", "../../../content/speech.json"); err == nil {
		t.Error("Load succeeded without a page file")
	}
}
