package exec

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
    "github.com/shxsun/beelog"
)

var ErrInvalid = errors.New("error invalid")
var ErrTimeout = errors.New("error timeout")

type SuperCmd struct {
	*exec.Cmd
	Timeout time.Duration
    EnvKey string
    EnvVal string
	IsClean bool
}

// create a new instance
func Command(name string, args ...string) *SuperCmd {
	return &SuperCmd{exec.Command(name, args...), 0, "", "", false}
}

// run command and wait until program exits or timeout
func (c *SuperCmd) Run() (err error) {
    beelog.Info("start run:", c.Args)
	ch := make(chan error)
	// set env flag
	if c.IsClean {
        if c.EnvKey == "" || c.EnvVal == ""  {
			return ErrInvalid
		}
        c.Env = append(c.Env, fmt.Sprintf("%s=%s", c.EnvKey, c.EnvVal))
	}
	// start program
	if err = c.Start(); err != nil {
		return
	}

	go func() {
		err := c.Wait()
		ch <- err
	}()

	// for timeout help
	if c.Timeout > time.Duration(0) {
		log.Println("start timeout process monitor")
		select {
		case err = <-ch:
			fmt.Println("ret", err)
			return err
		case <-time.After(c.Timeout):
			log.Println("running into timeout, send SIGKILL")

			err = c.Process.Kill()
			if err != nil {
				log.Println(err)
			}

			return ErrTimeout
		}
	}
	if c.IsClean {
		log.Printf("Do clean")
        c.KillAll()
	}
	return <-ch
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
	log.Printf("Start kill by env %s=%s %s\n", key, value, sig)
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
					fmt.Println("KILL CATCH ENV:", x, "PID:", pid)
					syscall.Kill(pid, sig)
				}
			}
		}
	}
}
