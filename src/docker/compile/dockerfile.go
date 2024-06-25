package compile

import (
	"fmt"
	"goanchor/src/docker/file"
	"goanchor/src/shared/collections"
	"os"
	"strings"
	"sync"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

type astIterInfo struct {
	parentNode *parser.Node
	childIndex int
}

type NodeInfo struct {
	ParentNode *parser.Node
	Node       *parser.Node
	Name       string
	Value      string
	File       string
}

type AnchorPoint struct {
	NodeInfo
	AnchorPath string
	Anchorfile *Dockerfile
}

type AstIter struct {
	file     string
	iterInfo *collections.Stack[*astIterInfo]
}

func GetIter(d *Dockerfile) *AstIter {
	return &AstIter{
		file: d.File,
		iterInfo: collections.NewStack(&astIterInfo{
			parentNode: d.Ast,
			childIndex: 0,
		})}
}

func (a *AstIter) getCurNode() *parser.Node {
	info := a.iterInfo.Peek()
	if info != nil {
		return info.parentNode.Children[info.childIndex]
	}
	return nil
}

func (a *AstIter) GetNext() *NodeInfo {
	curNode := a.getCurNode()
	if curNode == nil {
		return nil
	}
	info := &NodeInfo{
		ParentNode: a.iterInfo.Peek().parentNode,
		Node:       curNode,
		Name:       curNode.Value,
		Value:      "",
		File:       a.file,
	}

	values := []string{}
	if curNode.Value != "" {
		values = append(values, curNode.Flags...)
	}

	for n := curNode.Next; n != nil; n = n.Next {
		values = append(values, n.Value)
	}
	info.Value = strings.Join(values, " ")

	for a.iterInfo.Peek() != nil {
		if a.iterInfo.Peek().childIndex < len(a.iterInfo.Peek().parentNode.Children)-1 {
			a.iterInfo.Peek().childIndex++
			break
		}
		a.iterInfo.Pop()
	}
	return info
}

func extractAnchor(comment string) (string, bool) {
	if !strings.Contains(comment, "|<-") {
		return "", false
	}
	return strings.Split(comment, "|<-")[1], true
}

type Dockerfile struct {
	File string
	Ast  *parser.Node
	Args []*NodeInfo
	Envs []*NodeInfo
	From *NodeInfo
	Cmds []*NodeInfo

	AnchorPoints []*AnchorPoint
	Parent       *Dockerfile
}

func (d *Dockerfile) findAnchorPoints() error {
	iter := GetIter(d)
	var prevNode *NodeInfo
	for n := iter.GetNext(); n != nil; n = iter.GetNext() {
		for _, c := range n.Node.PrevComment {
			if a, exists := extractAnchor(c); exists {
				a, err := file.FindDockerFile(a)
				if err != nil {
					return err
				}
				d.AnchorPoints = append(d.AnchorPoints, &AnchorPoint{
					NodeInfo:   *prevNode,
					AnchorPath: a,
				})
			}
		}
		prevNode = n
	}
	return nil
}

func NewDockerfile(dfPath string) (*Dockerfile, error) {
	f, err := os.Open(dfPath)
	if err != nil {
		return nil, err
	}
	res, err := parser.Parse(f)
	if err != nil {
		return nil, err
	}
	d := &Dockerfile{
		File: dfPath,
		Ast:  res.AST,
	}

	iter := GetIter(d)
	for n := iter.GetNext(); n != nil; n = iter.GetNext() {
		if strings.EqualFold(n.Name, "ARG") {
			d.Args = append(d.Args, n)
		} else if strings.EqualFold(n.Name, "ENV") {
			d.Envs = append(d.Envs, n)
		} else if strings.EqualFold(n.Name, "FROM") {
			d.From = n // Do not check that there is only single FROM, let docker handle that
		} else if strings.EqualFold(n.Name, "CMD") {
			d.Cmds = append(d.Cmds, n)
		}
	}

	if err = d.findAnchorPoints(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Dockerfile) RaiseAnchors() error {
	parsedAnchors := collections.NewSet[string]()
	wg := sync.WaitGroup{}
	errCh := make(chan error)
	doneCh := make(chan struct{})

	var parse func(*Dockerfile, *sync.WaitGroup)
	parse = func(df *Dockerfile, wg *sync.WaitGroup) {
		defer wg.Done()
		for _, ap := range d.AnchorPoints {
			if exists := parsedAnchors.InserIfNotExist(ap.AnchorPath); exists {
				continue
			}
			a, err := NewDockerfile(ap.AnchorPath)
			if err != nil {
				panic(err)
			}
			a.Parent = df
			ap.Anchorfile = a
			wg.Add(1)
			go parse(a, wg)
		}
	}
	wg.Add(1)
	go parse(d, &wg)

	go func() {
		wg.Wait()
		close(doneCh)
	}()

	var errToReturn error
	for {
		select {
		case errToReturn = <-errCh:
		case <-doneCh:
			return errToReturn
		}
	}
}

func (d *Dockerfile) isDockerfile() bool {
	return d.Parent == nil
}

func (d *Dockerfile) getDockerfile() *Dockerfile {
	if d.isDockerfile() {
		return d
	} else {
		return d.Parent.getDockerfile()
	}
}

func (d *Dockerfile) CleanAnchorfiles() {
	wg := sync.WaitGroup{}

	wg.Add(1)
	go d.cleanAnchorfiles(&wg)

	wg.Wait()
}

func (d *Dockerfile) cleanAnchorfiles(wg *sync.WaitGroup) {
	defer wg.Done()
	for _, ap := range d.AnchorPoints {
		wg.Add(1)
		go ap.Anchorfile.cleanAnchorfiles(wg)
	}

	if d.isDockerfile() {
		return
	}

	d.removeNode(d.From)
}

func (d *Dockerfile) removeNode(node *NodeInfo) {
	pNode := node.ParentNode

	for i, n := range node.ParentNode.Children {
		if n == node.Node {
			pNode.Children = append(pNode.Children[:i], pNode.Children[i+1:]...)
		}
	}
}

func (d *Dockerfile) ToString() string {
	df := ""
	iter := GetIter(d)
	for n := iter.GetNext(); n != nil; n = iter.GetNext() {
		df += fmt.Sprintf("%s %s\n", n.Name, n.Value)
		for _, ap := range d.AnchorPoints {
			if n.Node == ap.Node {
				df += ap.Anchorfile.ToString()
			}
		}
	}
	return df
}
