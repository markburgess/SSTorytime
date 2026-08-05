package main

import (
	"reflect"
	"testing"
)

func TestApplyMultiCall(t *testing.T) {
	old := multiCall
	multiCall = map[string]string{
		"N4L":       "n4l",
		"searchN4L": "search",
	}
	t.Cleanup(func() { multiCall = old })

	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "busybox n4l",
			in:   []string{"/usr/bin/N4L", "-u", "doors.n4l"},
			want: []string{"/usr/bin/N4L", "n4l", "-u", "doors.n4l"},
		},
		{
			name: "already has subcommand",
			in:   []string{"/usr/bin/N4L", "n4l", "-u", "x.n4l"},
			want: []string{"/usr/bin/N4L", "n4l", "-u", "x.n4l"},
		},
		{
			name: "canonical name unchanged",
			in:   []string{"sstorytime", "n4l", "-u", "x.n4l"},
			want: []string{"sstorytime", "n4l", "-u", "x.n4l"},
		},
		{
			name: "searchN4L",
			in:   []string{"searchN4L", "door"},
			want: []string{"searchN4L", "search", "door"},
		},
		{
			name: "exe suffix on basename",
			in:   []string{"N4L.exe", "-u", "a.n4l"},
			want: []string{"N4L.exe", "n4l", "-u", "a.n4l"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := applyMultiCall(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("applyMultiCall(%v) = %v; want %v", tc.in, got, tc.want)
			}
		})
	}
}
