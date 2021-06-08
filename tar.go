package sup

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// Copying dirs/files over SSH using TAR.
// upload:
// tar -C . -cvzf - $SRC | ssh $HOST "tar -C $DST -xvzf -"
// download:
// ssh $HOST "tar -C $SRC_DIR -czvf - $SRC_FILE" | tar -C $DST -xzvf -

// RemoteTarCommand returns command to be run on remote SSH host
// to properly receive the created TAR stream.
// TODO: Check for relative directory.
func RemoteTarCommand(dir string) string {
	return fmt.Sprintf("tar -C \"%s\" -xzf -", dir)
}

func LocalTarCmdArgs(path, exclude string) []string {
	args := []string{}

	// Added pattens to exclude from tar compress
	excludes := strings.Split(exclude, ",")
	for _, exclude := range excludes {
		trimmed := strings.TrimSpace(exclude)
		if trimmed != "" {
			args = append(args, `--exclude=`+trimmed)
		}
	}

	args = append(args, "-C", ".", "-czf", "-", path)
	return args
}

// NewTarStreamReader creates a tar stream reader from a local path.
// TODO: Refactor. Use "archive/tar" instead.
func NewTarStreamReader(cwd, path, exclude string) (io.Reader, error) {
	cmd := exec.Command("tar", LocalTarCmdArgs(path, exclude)...)
	cmd.Dir = cwd
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "tar: stdout pipe failed")
	}

	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(err, "tar: starting cmd failed")
	}

	return stdout, nil
}

// RemoteTarCreateCommand forms "tar -C $SRC_DIR -czvf - $SRC_FILE"
// which is the remote part of download task
func RemoteTarCreateCommand(dir, src string) string {
	return fmt.Sprintf("tar -C \"%s\" -czvf - \"%s\"", dir, src)
}

// NewTarStreamWriter creates a tar stream writer to local path
// by calling tar -C $DST -xzvf -
// which is the local part of download task
func NewTarStreamWriter(dst string) (io.Writer, error) {
	cmd := exec.Command("tar", "-C", dst, "-xzvf", "-")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.Wrap(err, "tar: stdin pipe failed")
	}

	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(err, "tar: starting cmd failed")
	}

	return stdin, nil
}
