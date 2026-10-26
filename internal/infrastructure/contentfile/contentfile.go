package contentfile

import (
	"fmt"
	"os"

	"github.com/artim-ctrl/domashka/internal/domain/content"
)

func Load(pagePath, speechPath string) (*content.Page, error) {
	pageData, err := os.ReadFile(pagePath)
	if err != nil {
		return nil, fmt.Errorf("read page: %w", err)
	}
	speechData, err := os.ReadFile(speechPath)
	if err != nil {
		return nil, fmt.Errorf("read speech: %w", err)
	}
	page, err := content.Parse(pageData, speechData)
	if err != nil {
		return nil, fmt.Errorf("%s, %s: %w", pagePath, speechPath, err)
	}
	return page, nil
}
