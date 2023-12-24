package sup_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adamwasila/sup"
)

// TestNewSupfile reads content of testdata and run separate test of NewSupfile for each file found there.
// Files prefixed with "invalid_" have to have invalid content and are requiree to result in an error.
// Upon succesful NewSupfile call resulting struct is then validated.
// Each file can have an optional remark that will be used as descriptive test name so filenames
// can remain short and compact.
func TestNewSupfile(t *testing.T) {
	testRemarks := map[string]string{
		"Supfile_empty": "empty file is valid file",
		"Supfile_full":  "supfile that has every possible option and feature used",
	}

	baseDir := filepath.Join(".", "testdata")

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range entries {
		description := ""
		ok := false
		if description, ok = testRemarks[f.Name()]; !ok {
			description = fmt.Sprintf("Supfile: %s", f.Name())
		}

		wantErr := strings.HasPrefix(f.Name(), "invalid_")

		t.Run(description, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join(baseDir, f.Name()))
			if err != nil {
				t.Fatal(err)
			}

			got, err := sup.NewSupfile(b)
			if (err != nil) != wantErr {
				t.Errorf("NewSupfile() error = %v, wantErr %v", err, wantErr)
				return
			}
			if err := simpleValidator(got); err != nil {
				t.Errorf("NewSupfile() result is invalid because of: %v", err)
			}
		})
	}
}

func simpleValidator(s *sup.Supfile) error {
	if s == nil {
		// empty file is valid but require no further validation
		return nil
	}
	if s.Version == "" {
		return fmt.Errorf("unknown version")
	}
	return nil
}
