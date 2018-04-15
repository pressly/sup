package mock

import (
	"fmt"
)

type MockEnv struct {
	data map[string]string
}

func New() *MockEnv {
	e := &MockEnv{}
	e.Clearenv()
	return e
}

func (e *MockEnv) Environ() []string {
	out := []string{}

	for k, v := range e.data {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}

	return out
}

func (e *MockEnv) Getenv(key string) string {
	return e.data[key]
}

func (e *MockEnv) Setenv(key, value string) error {
	e.data[key] = value
	return nil
}

func (e *MockEnv) Clearenv() {
	e.data = make(map[string]string)
}
