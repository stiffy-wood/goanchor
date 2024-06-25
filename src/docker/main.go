package main

import (
	"fmt"
	"goanchor/src/docker/compile"
	"goanchor/src/docker/file"
	"os"
)

func main() {
    dfPath, err := file.FindDockerFile(os.Args[1])
    if err != nil {
        panic(err)
    }

    df, err := compile.NewDockerfile(dfPath)
    if err != nil {
        panic(err)
    }
    
    err = df.RaiseAnchors()
    if err != nil {
        panic(err)
    }

    df.CleanAnchorfiles()

    fmt.Println(df.ToString())
}
