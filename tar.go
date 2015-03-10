package stackup

import (
	"fmt"
	"io"

	"os/exec"
)

// Copying dirs/files over SSH using TAR.
// tar -C . -cvzf - <dirs/files> | ssh <host> "tar -C <dst_dir> -xvzf -"

// RemoteTarCommand returns command to be run on remote SSH host
// to properly receive the created TAR stream.
// TODO: Check for relative directory.
func RemoteTarCommand(dir string) string {
	return fmt.Sprintf("tar -C \"%s\" --checkpoint=100 --totals -xzf -", dir)
}

func LocalTarCommand(path string) string {
	return fmt.Sprintf("tar -C '.' -czf - %s", path)
}

// NewTarStreamReader creates a tar stream reader from a local path.
// TODO: Refactor. Use "archive/tar" instead.
func NewTarStreamReader(path, env string) io.Reader {
	cmd := exec.Command("bash", "-c", env+LocalTarCommand(path))

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
