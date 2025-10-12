package interactive

import (
	"os"
	"strconv"

	"ztictl/pkg/logging"

	"github.com/ktr0731/go-fuzzyfinder"
)

// getDisplayItemCount returns the number of items to display in the fuzzy finder
func getDisplayItemCount() int {
	heightStr := os.Getenv("ZTICTL_SELECTOR_HEIGHT")
	if heightStr == "" {
		return 10 // Default to 10 items
	}

	height, err := strconv.Atoi(heightStr)
	if err != nil || height < 1 {
		logging.LogWarn("Invalid ZTICTL_SELECTOR_HEIGHT value '%s', using default of 10", heightStr)
		return 10
	}

	if height > 20 {
		logging.LogWarn("ZTICTL_SELECTOR_HEIGHT of %d is too large, limiting to 20", height)
		return 20
	}

	return height
}

// FuzzyFind is a generic fuzzy finder function.
func FuzzyFind(items interface{}, itemFunc func(i int) string, header string, previewFunc func(i, w, h int) string) (int, error) {
	maxDisplayItems := getDisplayItemCount()
	totalHeight := maxDisplayItems + 5

	return fuzzyfinder.Find(items,
		itemFunc,
		fuzzyfinder.WithCursorPosition(fuzzyfinder.CursorPositionBottom),
		fuzzyfinder.WithPromptString("🔍 Type to search > "),
		fuzzyfinder.WithHeader(header),
		fuzzyfinder.WithMode(fuzzyfinder.ModeSmart),
		fuzzyfinder.WithHeight(totalHeight),
		fuzzyfinder.WithHorizontalAlignment(fuzzyfinder.AlignLeft),
		fuzzyfinder.WithBorder(),
		fuzzyfinder.WithPreviewWindow(previewFunc),
	)
}
