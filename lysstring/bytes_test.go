package lysstring

import "testing"

func TestFormatBytes(t *testing.T) {
	const unit int64 = 1024

	tests := []struct {
		name  string
		input int64
		want  string
	}{
		{name: "zero", input: 0, want: "0 B"},
		{name: "bytes", input: 512, want: "512 B"},
		{name: "kilobytes", input: unit, want: "1 KB"},
		{name: "truncates kilobytes", input: unit + unit/2, want: "1 KB"},
		{name: "megabytes", input: unit * unit, want: "1 MB"},
		{name: "gigabytes", input: unit * unit * unit, want: "1 GB"},
		{name: "terabytes", input: unit * unit * unit * unit, want: "1 TB"},
		{name: "petabytes", input: unit * unit * unit * unit * unit, want: "1 PB"},
		{name: "larger than petabyte", input: unit * unit * unit * unit * unit * 2, want: "2 PB"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := FormatBytes(test.input); got != test.want {
				t.Errorf("FormatBytes(%d) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
