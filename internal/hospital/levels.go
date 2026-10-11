package hospital

import (
	"fmt"
	"strings"
)

func ParseNarcolepsyLevel(s string) (NarcolepsyLevel, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "mild":
		return NarcolepsyLevelMild, nil
	case "moderate":
		return NarcolepsyLevelModerate, nil
	case "severe":
		return NarcolepsyLevelSevere, nil
	default:
		return 0, fmt.Errorf("narcolepsy level %q: %w", s, ErrInvalidInput)
	}
}
