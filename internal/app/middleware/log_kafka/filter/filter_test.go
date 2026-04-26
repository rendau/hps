package filter

import "testing"

type filterTestCase struct {
	name       string
	method     string
	pathStr    string
	filterRule []string
	expected   bool
}

func TestFilter(t *testing.T) {
	tests := []filterTestCase{
		{
			name:    "custom url 1",
			method:  "GET",
			pathStr: "/api/v3/product/basseyn-bestway-karkasnyy-steel-pro-splash-in-shade-244-h-51-sm-1688-l-56432",
			filterRule: []string{
				"GET:/api/v2/product/*",
				"GET:/api/v3/product/*",
			},
			expected: true,
		},
		{
			name:    "custom url 2",
			method:  "GET",
			pathStr: "/api/v3/product/smart-chasy-huawei-gt-6-46mm-cvet-chernyy-atm-b19",
			filterRule: []string{
				"GET:/api/v2/product/*",
				"GET:/api/v3/product/*",
			},
			expected: true,
		},
		{
			name:       "no filter rules deny everything",
			method:     "GET",
			pathStr:    "/any/path",
			filterRule: []string{},
			expected:   false,
		},
		{
			name:    "match specific method and path",
			method:  "POST",
			pathStr: "/submit/data",
			filterRule: []string{
				"POST:/submit/data",
			},
			expected: true,
		},
		{
			name:    "match any method with path pattern",
			method:  "GET",
			pathStr: "/api/item/123",
			filterRule: []string{
				"/api/item/*",
			},
			expected: true,
		},
		{
			name:    "match specific method with wildcard path",
			method:  "DELETE",
			pathStr: "/records/456",
			filterRule: []string{
				"DELETE:/records/*",
			},
			expected: true,
		},
		{
			name:    "method mismatch",
			method:  "GET",
			pathStr: "/admin",
			filterRule: []string{
				"POST:/admin",
			},
			expected: false,
		},
		{
			name:    "no match due to path mismatch",
			method:  "GET",
			pathStr: "/different/path",
			filterRule: []string{
				"GET:/specific/path",
			},
			expected: false,
		},
		{
			name:    "case insensitivity in path",
			method:  "GET",
			pathStr: "/Mixed/Case/tEsT",
			filterRule: []string{
				"/mixed/case/test",
			},
			expected: true,
		},
		{
			name:    "case insensitivity in method",
			method:  "get",
			pathStr: "/case/test",
			filterRule: []string{
				"GET:/case/test",
			},
			expected: true,
		},
		{
			name:    "empty path matches root",
			method:  "GET",
			pathStr: "",
			filterRule: []string{
				"/",
			},
			expected: true,
		},
		{
			name:    "path trimming works",
			method:  "GET",
			pathStr: "///trim/test//",
			filterRule: []string{
				"/trim/test",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := New(tt.filterRule)
			result := f.Check(tt.method, tt.pathStr)
			if result != tt.expected {
				t.Errorf("Check(%q, %q) expected %v, got %v", tt.method, tt.pathStr, tt.expected, result)
			}
		})
	}
}
