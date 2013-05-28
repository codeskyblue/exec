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
)

type JfCmd struct {
	*exec.Cmd
	Timeout time.Duration
	EnvFlag string
	IsClean bool
}

// create a new instance
func Command(name string, args ...string) *JfCmd {
	return &JfCmd{exec.Command(name, args...), 0, "", false}
}

// run command and wait until program exits or timeout
func (c *JfCmd) Run() (err error) {
	log.Println("Start Run:", c.Args)
	ch := make(chan error)
	var key, val string
	// set env flag
	if c.IsClean {
		if v := strings.SplitN(c.EnvFlag, "=", 2); len(v) != 2 {
			return errors.New("EnvFlag not set as KEY=VALUE")
		} else {
			key, val = v[0], v[1]
			log.Println(v)
			c.Env = append(c.Env, c.EnvFlag)
		}
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

			return errors.New("timeout")
		}
	}
	if c.IsClean {
		log.Printf("Do clean")
		KillallByEnv(key, val, syscall.SIGTERM)
		KillallByEnv(key, val, syscall.SIGKILL)
	}
	return <-ch
}

// Output runs the command and returns its standard output.
func (c *JfCmd) Output() ([]byte, error) {
	if c.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	var b bytes.Buffer
	c.Stdout = &b
	err := c.Run()
	return b.Bytes(), err
}

// Read Process Env
func ReadProcessEnv(pid int) (m map[string]string, e error) {
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
func KillallByEnv(key string, value string, sig syscall.Signal) {
	log.Printf("Start kill by env %s=%s %s\n", key, value, sig)
	files, _ := ioutil.ReadDir("/proc")
	for _, file := range files {
		if file.IsDir() {
			pid, e := strconv.Atoi(file.Name())
			if e != nil {
				continue
			}
			env, err := ReadProcessEnv(pid)
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
