package main

import (
	"context"
	"fmt"
	"goanchor/src/compile"
	"goanchor/src/file"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	r := file.NewReader(context.Background())

	ch := make(chan string)
	go r.ReadLayers("./Dockerfile.example", false, ch)
    for {
        select{
        case <- r.Done():
            return
        case layer := <- ch:
            //fmt.Printf("########\n\n%s\n\n#########", layer)
            fmt.Printf("[%s]\n", strings.Join( compile.TokenizeLayer(layer), ", "))
        }
    }
}

func startDocker() {
	exe := filepath.Join(filepath.Dir(os.Args[0]), "original.docker.exe")
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	cmd.Wait()
}
