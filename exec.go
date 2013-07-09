package exec

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/shxsun/beelog"
	"io/ioutil"
	"math/rand"
	"os/exec"
	"strings"
	"time"
)

var ErrInvalid = errors.New("error invalid")
var ErrTimeout = errors.New("error timeout")

var defaultEnvName = "TIMEOUT_EXEC_ID"

type Cmd struct {
	Timeout time.Duration
	UniqID  string
	IsClean bool
	*exec.Cmd
}

func init() {
	beelog.SetLevel(beelog.LevelWarning)
}

// create a new instance
func Command(name string, args ...string) *Cmd {
	cmd := &Cmd{}
	cmd.Cmd = exec.Command(name, args...)
	return cmd
}

func (c *Cmd) WaitTimeout(timeout time.Duration) error {
	done := make(chan error)
	go func() {
		done <- c.Wait()
	}()

	select {
	case <-time.After(timeout):
		return ErrTimeout
	case err := <-done:
		return err
	}
}

// run command and wait until program exits or timeout
func (c *Cmd) Run() (err error) {
	beelog.Info("start run:", c.Args)
	// set env flag
	if c.IsClean {
		if c.UniqID == "" {
			c.UniqID = fmt.Sprintf("%d:%d", time.Now().UnixNano(), rand.Int())
		}
		c.Env = append(c.Env, defaultEnvName+"="+c.UniqID)
	}
	// start program
	if err = c.Start(); err != nil {
		return
	}

	if c.Timeout > time.Duration(0) {
		err = c.WaitTimeout(c.Timeout)
	} else {
		err = c.Wait()
	}
	if c.IsClean {
		c.KillAll()
		return
	} else if err == ErrTimeout {
		c.Process.Kill()
	}
	return
}

// Output runs the command and returns its standard output.
func (c *Cmd) Output() ([]byte, error) {
	if c.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	var b bytes.Buffer
	c.Stdout = &b
	err := c.Run()
	return b.Bytes(), err
}

// get spectified pid of environ
func procEnv(pid int) ([]string, error) {
	data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/environ", pid))
	if err != nil {
		return nil, err
	}
	envs := strings.Split(string(data), "\x00")
	return envs, nil
}
