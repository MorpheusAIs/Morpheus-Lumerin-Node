package main

import (
    "encoding/json"
    "path"
    "log"
    "io"
    "os"
    "os/exec"
    "sync"
    "github.com/google/shlex"
)

type Config struct {
    Run []string `json:"run"`
}

func findConfig(fileName string) (string, error) {
    homeFile := path.Join(os.Getenv("HOME"), fileName)
    if _, err := os.Stat(fileName); err == nil {
        return fileName, nil
    } else if _, err := os.Stat(homeFile); err == nil {
        return homeFile, nil
    } else {
        return "", os.ErrNotExist
    }
}

func main() {
    confName := "mor-launch.json"
    confFile, err := findConfig(confName)
    if err != nil {
        log.Fatalf("Error finding %s: %v", confName, err)
    }
    data, err := os.ReadFile(confFile)
    if err != nil {
        log.Fatalf("Error reading %s: %v", confName, err)
    }
    var conf Config
    err = json.Unmarshal(data, &conf)
    if err != nil {
        log.Fatal("Error during Unmarshal(): ", err)
    }
    var wg sync.WaitGroup
    for _, v := range conf.Run {
        wg.Add(1)
        go func(v string) {
            defer wg.Done()
            var args []string
            var err error
            args, err = shlex.Split(v)
            if err != nil {
                log.Print(err)
                return
            }
            if len(args) == 0 {
                log.Print("No args, bailing")
                return
            }
            cmd := exec.Command(args[0], args[1:]...)
            stdout, err := cmd.StdoutPipe()
            if err != nil {
                log.Print(err)
                return
            }
            if err := cmd.Start(); err != nil {
                log.Print(err)
                return
            }
            if _, err := io.Copy(os.Stdout, stdout); err != nil {
                log.Print(err)
                return
            }
            if err := cmd.Wait(); err != nil {
                log.Print(err)
            }
        }(v)
    }
    wg.Wait()
}

