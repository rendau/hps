package middleware

import (
	"testing"
)

type filterTestCase struct {
	name       string
	method     string
	pathStr    string
	filterRule []FilterRule
	expected   bool
}

func TestFilter(t *testing.T) {
	tests := []filterTestCase{
		{
			name:    "custom url",
			method:  "GET",
			pathStr: "/api/v3/product/basseyn-bestway-karkasnyy-steel-pro-splash-in-shade-244-h-51-sm-1688-l-56432",
			filterRule: []FilterRule{
				{Method: "GET", Pattern: "/api/v2/product/*"},
				{Method: "GET", Pattern: "/api/v3/product/*"},
			},
			expected: true,
		},
		{
			name:    "custom url",
			method:  "GET",
			pathStr: "/api/v3/product/smart-chasy-huawei-gt-6-46mm-cvet-chernyy-atm-b19",
			filterRule: []FilterRule{
				{Method: "GET", Pattern: "/api/v2/product/*"},
				{Method: "GET", Pattern: "/api/v3/product/*"},
			},
			expected: true,
		},
		{
			name:       "no filter rules allow everything",
			method:     "GET",
			pathStr:    "/any/path",
			filterRule: []FilterRule{},
			expected:   true,
		},
		{
			name:    "match specific method and path",
			method:  "POST",
			pathStr: "/submit/data",
			filterRule: []FilterRule{
				{Method: "POST", Pattern: "/submit/data"},
			},
			expected: true,
		},
		{
			name:    "match any method with path pattern",
			method:  "GET",
			pathStr: "/api/item/123",
			filterRule: []FilterRule{
				{Pattern: "/api/item/*"},
			},
			expected: true,
		},
		{
			name:    "match specific method with wildcard path",
			method:  "DELETE",
			pathStr: "/records/456",
			filterRule: []FilterRule{
				{Method: "DELETE", Pattern: "/records/*"},
			},
			expected: true,
		},
		{
			name:    "method mismatch",
			method:  "GET",
			pathStr: "/admin",
			filterRule: []FilterRule{
				{Method: "POST", Pattern: "/admin"},
			},
			expected: false,
		},
		{
			name:    "no match due to path mismatch",
			method:  "GET",
			pathStr: "/different/path",
			filterRule: []FilterRule{
				{Method: "GET", Pattern: "/specific/path"},
			},
			expected: false,
		},
		{
			name:    "case insensitivity in path",
			method:  "GET",
			pathStr: "/Mixed/Case/tEsT",
			filterRule: []FilterRule{
				{Pattern: "/mixed/case/test"},
			},
			expected: true,
		},
		{
			name:    "case insensitivity in method",
			method:  "get",
			pathStr: "/case/test",
			filterRule: []FilterRule{
				{Method: "GET", Pattern: "/case/test"},
			},
			expected: true,
		},
		{
			name:    "empty path matches root",
			method:  "GET",
			pathStr: "",
			filterRule: []FilterRule{
				{Pattern: "/"},
			},
			expected: true,
		},
		{
			name:    "path trimming works",
			method:  "GET",
			pathStr: "///trim/test//",
			filterRule: []FilterRule{
				{Pattern: "/trim/test"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logKafka := &LogKafka{filterRules: tt.filterRule}
			result := logKafka.filter(tt.method, tt.pathStr)
			if result != tt.expected {
				t.Errorf("filter(%q, %q) expected %v, got %v", tt.method, tt.pathStr, tt.expected, result)
			}
		})
	}
}
