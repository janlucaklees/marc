package main

import (
	"reflect"
	"testing"
)

func TestPermuteArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "no arguments",
			args: []string{},
			want: nil,
		},
		{
			name: "positional only, no flags",
			args: []string{"input.md"},
			want: []string{"input.md"},
		},
		{
			name: "flag before positional, short form, unchanged",
			args: []string{"-t", "work", "input.md"},
			want: []string{"-t", "work", "input.md"},
		},
		{
			name: "flag before positional, long form, unchanged",
			args: []string{"--template", "work", "input.md"},
			want: []string{"--template", "work", "input.md"},
		},
		{
			name: "flag after positional, short form, reordered",
			args: []string{"input.md", "-t", "work"},
			want: []string{"-t", "work", "input.md"},
		},
		{
			name: "flag after positional, long form, reordered",
			args: []string{"input.md", "--template", "work"},
			want: []string{"--template", "work", "input.md"},
		},
		{
			name: "output flag after positional, short form, reordered",
			args: []string{"input.md", "-o", "out.pdf"},
			want: []string{"-o", "out.pdf", "input.md"},
		},
		{
			name: "output flag after positional, long form, reordered",
			args: []string{"input.md", "--output", "out.pdf"},
			want: []string{"--output", "out.pdf", "input.md"},
		},
		{
			name: "both flags after positional, mixed forms, reordered and kept paired",
			args: []string{"input.md", "-t", "work", "--output", "out.pdf"},
			want: []string{"-t", "work", "--output", "out.pdf", "input.md"},
		},
		{
			name: "both flags before positional, unchanged",
			args: []string{"-t", "work", "-o", "out.pdf", "input.md"},
			want: []string{"-t", "work", "-o", "out.pdf", "input.md"},
		},
		{
			name: "flag value with = form after positional",
			args: []string{"input.md", "-t=work"},
			want: []string{"-t=work", "input.md"},
		},
		{
			name: "long flag value with = form after positional",
			args: []string{"input.md", "--template=work"},
			want: []string{"--template=work", "input.md"},
		},
		{
			name: "flag value does not consume a following positional beyond what it needs",
			args: []string{"-t", "work", "input.md", "-o", "out.pdf"},
			want: []string{"-t", "work", "-o", "out.pdf", "input.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := permuteArgs(tt.args)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("permuteArgs(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
