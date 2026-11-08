package flatten

import (
	"fmt"
	"strings"
)

// uniqueName is name, made a valid component name, or else that with the lowest number after it that m has no entry
// for.
func uniqueName[M ~map[string]V, V any](m M, name string) string {
	name = componentName(name)
	idx := 1
	altName := name

	for {
		// check if the name already exists
		if _, ok := m[altName]; !ok {
			return altName
		}

		// increase the number next to the name
		idx++
		altName = fmt.Sprintf("%s%d", name, idx)
	}
}

// componentName is name without what a component name may not hold: anything but ASCII letters, digits, '.', '-' and
// '_', as the braces of a path template in an operation ID or the padding of a base64 property key. Where nothing is
// left, it is "Unnamed".
func componentName(name string) string {
	name = strings.Map(func(r rune) rune {
		switch {
		case 'a' <= r && r <= 'z', 'A' <= r && r <= 'Z', '0' <= r && r <= '9', r == '.', r == '-', r == '_':
			return r
		default:
			return -1
		}
	}, name)

	if name == "" {
		return "Unnamed"
	}

	return name
}
