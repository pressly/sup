package sup

import (
	"fmt"
	"strings"
)

func NewRSyncCommand(path, dst, user, host string) string {
	hostPort := strings.Split(host, ":")
	return fmt.Sprintf("rsync -ac --port %s %s %s@%s:%s", hostPort[1], path, user, hostPort[0], dst)
}
