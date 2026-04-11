package observability

import "fmt"

const truncationSuffix = "...(truncated)"

func TruncateString(s string, maxBytes int) string {
	if maxBytes <= 0 || len(s) <= maxBytes {
		return s
	}
	if maxBytes <= len(truncationSuffix) {
		return truncationSuffix[:maxBytes]
	}

	return s[:maxBytes-len(truncationSuffix)] + truncationSuffix
}

func SummarizePayload(v interface{}, maxBytes int) string {
	if v == nil {
		return "<nil>"
	}

	return TruncateString(fmt.Sprintf("%T %v", v, v), maxBytes)
}
