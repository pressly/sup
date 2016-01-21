package sup

import (
	"fmt"
	"io"
	"strings"

	"os/exec"
)

// Copying dirs/files over SSH using TAR.
// tar -C . -cvzf - <dirs/files> | ssh <host> "tar -C <dst_dir> -xvzf -"

// RemoteTarCommand returns command to be run on remote SSH host
// to properly receive the created TAR stream.
// TODO: Check for relative directory.
func RemoteTarCommand(dir string) string {
	return fmt.Sprintf("tar -C \"%s\" -xzf -", dir)
}

func localTarOptions(path, exclude string) []string {

	// Added pattens to exclude from tar compress
	var excludes []string

	result := strings.Split(exclude, ",")

	for _, exclude := range result {
		if exclude != "" {
			excludes = append(excludes, "--exclude=" + strings.TrimSpace(exclude))
		}
	}

	tarOptions := append([]string{"-czf", "-"}, excludes...)
	return append(tarOptions, path)
}

// NewTarStreamReader creates a tar stream reader from a local path.
// TODO: Refactor. Use "archive/tar" instead.
func NewTarStreamReader(path, exclude string) io.Reader {
	cmd := exec.Command("tar", append(localTarOptions(path, exclude))...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil
	}

	output := io.MultiReader(stdout, stderr)

	if err := cmd.Start(); err != nil {
		return nil
	}

	return output
}
