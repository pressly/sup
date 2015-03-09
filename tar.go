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
	return fmt.Sprintf("tar -C \"%s\" -xvzf -", dir)
}

func LocalTarCommand(path string) string {
	return fmt.Sprintf("tar -C '.' -cvzf - %s", path)
}

// NewTarStreamReader creates a tar stream reader from a local path.
// TODO: Refactor. Use "archive/tar" instead.
func NewTarStreamReader(path, env string) io.Reader {
	// // Dumb way how to check if the "path" exist
	// _, err := os.Stat(path)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	cmd := exec.Command("bash", "-c", env+LocalTarCommand(path))
	//cmd := exec.Command("tar", "-C", ".", "-cvzf", "-", path)

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
