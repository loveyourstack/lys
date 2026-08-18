package lysstring

import "fmt"

func FormatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	_, exp := int64(unit), 0
	for n >= unit && exp < 5 {
		n /= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%d %s", n, units[exp-1])
}
