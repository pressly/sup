package os

import (
	"os"
)

type OsEnv struct {
}

func New() *OsEnv {
	return &OsEnv{}
}

func (e *OsEnv) Getenv(key string) string {
	return os.Getenv(key)
}

func (e *OsEnv) Setenv(key, value string) error {
	return os.Setenv(key, value)
}

func (e *OsEnv) Environ() []string {
	return os.Environ()
}

func (e *OsEnv) Clearenv() {
	os.Clearenv()
}
