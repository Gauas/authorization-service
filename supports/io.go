package supports

import "io"
func ReadBody(r io.Reader) string {
	raw, err := io.ReadAll(r)
	if err != nil {
		return ""
	}
	return string(raw)
}
