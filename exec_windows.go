package exec

import (
	"strings"
	"syscall"
)

func (c *Cmd) KillAll() (err error) {
	return c.Process.Kill()
}
