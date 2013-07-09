package main

import (
    "fmt"
    "github.com/shxsun/exec"
)

func main(){
    cmd := exec.Command("/bin/bash", "-c", "sleep 20 &")
    cmd.IsClean = true // 
    err := cmd.Run()
    if err != nil{
        println(err.Error())
    }
    fmt.Println(cmd.ProcessState.Pid())
}
