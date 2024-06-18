package file

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"goanchor/src/utils"
	"io"
	"os"
	"strings"
)

type DockerfileReader struct {
	log    utils.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

func NewReader(ctx context.Context) *DockerfileReader {
	newCtx, cancel := context.WithCancel(ctx)
	return &DockerfileReader{
		log:    *utils.NewLogger("DockerfileReader"),
		ctx:    newCtx,
		cancel: cancel,
	}
}

func (r *DockerfileReader) Done() <-chan struct{} {
	return r.ctx.Done()
}

func (r *DockerfileReader) ReadLayers(path string, isAnchor bool, outLayers chan string) {
	defer func() {
		r.cancel()
	}()

	if valid, err := isValid(path, isAnchor); !valid {
		r.log.Error(err.Error(), path)
		return
	}

	f, err := os.Open(path)
	if err != nil {
		r.log.Error(err.Error(), path)
	}
	defer f.Close()

	buf := make([]byte, 32*1024)

	var curLayer []byte

	for _, err = f.Read(buf); err == nil; {
		fmt.Println(err)
		for _, b := range buf {
			curLayer = append(curLayer, b)
			if b == byte('\n') {
				trimmed := bytes.TrimSpace(curLayer)
				if bytes.HasSuffix(trimmed, []byte("\\")) || len(trimmed) == 0 {
					continue
				}
				select {
				case outLayers <- string(trimmed):
					curLayer = []byte{}
				case <-r.ctx.Done():
					return
				}
			}
		}
		return
	}
	if err != io.EOF {
		r.log.Error(err.Error(), path)
	}
}

func isValid(path string, isAnchor bool) (bool, error) {
	f, err := os.Stat(path)
	if err == os.ErrNotExist {
		return false, err
	}

	if isAnchor && !strings.HasPrefix(strings.ToLower(f.Name()), "anchorfile") {
		return false, errors.New("is not an anchorfile")
	}

	return true, nil
}
