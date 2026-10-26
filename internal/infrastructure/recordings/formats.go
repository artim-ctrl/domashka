package recordings

import (
	"path/filepath"
	"strings"
)

type Format struct {
	Ext         string
	ContentType string
}

const fallbackContentType = "application/octet-stream"

func (l *Store) formatRank(name string) int {
	ext := strings.ToLower(filepath.Ext(name))
	for i := range l.Formats {
		if l.Formats[i].Ext == ext {
			return i
		}
	}
	return -1
}

func (l *Store) ContentType(path string) string {
	if r := l.formatRank(path); r >= 0 {
		return l.Formats[r].ContentType
	}
	return fallbackContentType
}
