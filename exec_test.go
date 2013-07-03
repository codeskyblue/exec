package exec

import (
	"testing"
	"time"
)

func TestNormal(t *testing.T) {
	cmd := Command("echo", "-n", "hello")
	output, err := cmd.Output()
	if err != nil {
		t.Errorf("expect normal, but got error: %s\n", err.Error())
	}
	if string(output) != "hello" {
		t.Errorf("expect 'hello', but got: %s\n", string(output))
	}
}

// FIXME
func TestNormalTimeout(t *testing.T) {
}

func TestTimeout(t *testing.T) {
	cmd := Command("sleep", "2")
	cmd.Timeout = time.Second * 1
	_, err := cmd.Output()
	if err != ErrTimeout {
		t.Errorf("expect ErrTimeout, but got: %s\n", err.Error())
	}
}

// FIXME
func TestCleanNormal(t *testing.T) {
}

// FIXME
func TestCleanTimeout(t *testing.T) {
}
