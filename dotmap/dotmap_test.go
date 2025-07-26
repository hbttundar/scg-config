package dotmap_test

import (
	"reflect"
	"testing"

	"github.com/hbttundar/scg-config/dotmap"
)

func TestResolver(t *testing.T) {
	t.Parallel()

	settings := map[string]interface{}{
		"App": map[string]interface{}{
			"Name": "scg-config",
			"Port": 8080,
			"List": []string{"a", "b"},
			"Deep": map[string]interface{}{
				"X": "Y",
			},
		},
		"NUMS": []interface{}{1, 2, 3},
		"MapStr": map[string]string{
			"Hello": "World",
		},
		"iface": map[interface{}]interface{}{
			"X": 10,
		},
		"Mixed": []interface{}{
			map[string]interface{}{"Key": "Value"},
		},
	}

	tests := []struct {
		name     string
		settings map[string]interface{}
		path     string
		want     interface{}
	}{
		{"nil settings", nil, "foo.bar", nil},
		{"empty path", settings, "", nil},
		{"missing path", settings, "not.exist", nil},

		// --- Case-sensitive tests ---
		{"case-sensitive: exact match", settings, "App.Name", "scg-config"},
		{"case-sensitive: list", settings, "App.List.1", "b"},
		{"case-sensitive: nested map", settings, "App.Deep.X", "Y"},
		{"case-sensitive: wrong case fails", settings, "app.name", "scg-config"}, // should fallback to ci
		{"case-sensitive: deep wrong case", settings, "app.deep.x", "Y"},         // should fallback to ci

		// --- Case-insensitive fallback ---
		{"ci: all lower", settings, "app.name", "scg-config"},
		{"ci: all upper", settings, "APP.PORT", 8080},
		{"ci: array element", settings, "nums.2", 3},
		{"ci: string array", settings, "app.list.0", "a"},
		{"ci: partial ci", settings, "App.LIST.1", "b"},
		{"ci: deep nested ci", settings, "app.deep.x", "Y"},
		{"ci: map[string]string", settings, "mapstr.hello", "World"},
		{"ci: map[interface{}]interface{}", settings, "iface.x", 10},
		{"ci: []interface{} of map", settings, "mixed.0.key", "Value"},

		// --- Fails ---
		{"fail: missing nested", settings, "App.Nope.Value", nil},
		{"fail: missing root", settings, "nope", nil},
		{"fail: missing index", settings, "App.List.99", nil},
		{"fail: partial path", settings, "App.", nil},
	}

	for _, testCase := range tests {
		// capture range var
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := dotmap.Resolve(testCase.settings, testCase.path)
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("Resolve(%q) = %v (%T), want %v (%T)", testCase.path, got, got, testCase.want, testCase.want)
			}
		})
	}
}
