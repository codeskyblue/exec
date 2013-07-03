exec
====
[![Build Status](https://drone.io/github.com/shxsun/exec/status.png)](https://drone.io/github.com/shxsun/exec/latest)

An implement of os/exec, but add timeout option

## Example

    package main

    import (
        "fmt"
        "github.com/shxsun/exec"
        "time"
    )

    func main() {
        cmd := exec.Command("sleep", "2")
        cmd.Timeout = time.Second * 1
        output, err := cmd.Output()
        if err != nil {
            fmt.Println(err)
            return
        }
        fmt.Println("Output: ", string(output))
    }
