package exec

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/shxsun/beelog"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var ErrInvalid = errors.New("error invalid")
var ErrTimeout = errors.New("error timeout")

type SuperCmd struct {
	*exec.Cmd
	Timeout time.Duration
	EnvKey  string
	EnvVal  string
	IsClean bool
}

// create a new instance
func Command(name string, args ...string) *SuperCmd {
	return &SuperCmd{exec.Command(name, args...), 0, "", "", false}
}

func (c *SuperCmd) WaitTimeout(timeout time.Duration) error {
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
func (c *SuperCmd) Run() (err error) {
	beelog.Info("start run:", c.Args)
	// set env flag
	if c.IsClean {
		if c.EnvKey == "" || c.EnvVal == "" {
			return ErrInvalid
		}
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", c.EnvKey, c.EnvVal))
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
	}
	if err == ErrTimeout {
		c.Process.Kill()
	}
	return
}

func (c *SuperCmd) KillAll() {
	killAllEnv(c.EnvKey, c.EnvVal, syscall.SIGTERM)
	killAllEnv(c.EnvKey, c.EnvVal, syscall.SIGKILL)
}

// Output runs the command and returns its standard output.
func (c *SuperCmd) Output() ([]byte, error) {
	if c.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	var b bytes.Buffer
	c.Stdout = &b
	err := c.Run()
	return b.Bytes(), err
}

// Read Process Env
func readProcessEnv(pid int) (m map[string]string, e error) {
	m = make(map[string]string)
	fd, e := os.Open("/proc/" + strconv.Itoa(pid) + "/environ")
	if e != nil {
		return nil, e
	}
	bi := bufio.NewReader(fd)
	for {
		line, err := bi.ReadString(byte(0))
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key, value := kv[0], kv[1]
		m[key] = value[:len(value)-1] // remove last \0
	}
	return
}

// Walk all proc and kill by Env which contails key=value
func killAllEnv(key string, value string, sig syscall.Signal) {
	beelog.Debug("kill by env, send ", sig)
	files, _ := ioutil.ReadDir("/proc")
	for _, file := range files {
		if file.IsDir() {
			pid, e := strconv.Atoi(file.Name())
			if e != nil {
				continue
			}
			env, err := readProcessEnv(pid)
			if err != nil {
				continue
			}
			if x, ok := env[key]; ok {
				//fmt.Println("CATCH ENV:", x, "PID:", pid)
				if x == value {
					beelog.Trace("kill", sig, pid)
					syscall.Kill(pid, sig)
				}
			}
		}
	}
}
