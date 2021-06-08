package sup

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestEnvListUnmarshalYAML(t *testing.T) {
	type holder struct {
		Env EnvList `yaml:"env"`
	}

	testCases := []struct {
		input  string
		expect holder
	}{
		{

			input: `
env:
  MY_KEY: abc123
`,
			expect: holder{
				Env: EnvList{
					&EnvVar{Key: "MY_KEY", Value: "abc123"},
				},
			},
		},
		{

			input: `
env:
  MY_KEY: $(echo abc123)
`,
			expect: holder{
				Env: EnvList{
					&EnvVar{Key: "MY_KEY", Value: "abc123"},
				},
			},
		},
	}

	for _, tc := range testCases {
		h := holder{}
		yaml.Unmarshal([]byte(tc.input), &h)
		if !reflect.DeepEqual(h, tc.expect) {
			t.Errorf("Unmarshalling yaml did not produce the expected result. Got:\n%#v\nExpected: %#v\n", h, tc.expect)
		}
	}
}
