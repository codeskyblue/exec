package exec

import (
	"github.com/shxsun/beelog"
	"github.com/shxsun/monitor"
	"strings"
	"syscall"
)

func (c *Cmd) KillAll() (err error) {
	if c.Process != nil {
		c.Process.Kill()
	}
	sig := syscall.SIGTERM
	pids, err := monitor.Pids()
	if err != nil {
		return
	}
	for _, pid := range pids {
		envs, err := procEnv(pid)
		if err != nil {
			continue
		}
		flag := ""
		for _, e := range envs {
			if strings.HasPrefix(e, defaultEnvName+"=") {
				flag = e
				break
			}
		}
		if flag == defaultEnvName+"="+c.UniqID {
			beelog.Trace("kill", sig, pid)
			syscall.Kill(pid, sig)
		}
	}
	return
}
