package candidates

import (
	"os"
	"testing"

	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
)

func TestStore(t *testing.T) {
	store := New(Options{
		Dir:      t.TempDir(),
		KindDirs: clip.Dirs{Sounds: "sounds", Words: "words"},
		TextFile: "text.txt",
		Ext:      ".wav",
	})
	c := clip.Clip{Kind: clip.Word, Key: "дом"}
	voices := []voiceover.Voice{{ID: "a"}, {ID: "b"}}

	cands, err := store.List(c, voices)
	if err != nil || len(cands) != 0 {
		t.Fatalf("empty store: %v %v", cands, err)
	}
	if text, err := store.Text(c); err != nil || text != "" {
		t.Fatalf("empty text: %q %v", text, err)
	}

	if err := store.SaveText(c, "д+ом"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.Path(c, "b"), []byte("wav"), 0o644); err != nil {
		t.Fatal(err)
	}

	cands, err = store.List(c, voices)
	if err != nil || len(cands) != 1 || cands[0].Voice != "b" {
		t.Errorf("after writing b: %v %v", cands, err)
	}
	if text, err := store.Text(c); err != nil || text != "д+ом" {
		t.Errorf("text: %q %v", text, err)
	}
	if _, ok, _ := store.Get(c, "a"); ok {
		t.Error("Get found a candidate that was never written")
	}
}
