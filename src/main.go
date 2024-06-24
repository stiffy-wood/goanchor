package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("fuuuuuck no")
	}

	dockerfile, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := parser.Parse(dockerfile)
	if err != nil {
		fmt.Println(err)
		return
	}

	var p func(node *parser.Node, indent string)
	p = func(node *parser.Node, indent string) {
		fmt.Printf("%s%s\n", indent, node.Value)
		for _, child := range node.Children {
			p(child, indent+" ")
		}
	}

    p(res.AST, "")
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
