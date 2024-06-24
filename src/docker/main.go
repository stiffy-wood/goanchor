package main

import (
	"fmt"
	"goanchor/src/shared"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

const (
	INSTRUCT_ITER_STEP = 0
	NEXT_ITER_STEP     = 1
)

type varInstruct struct {
	name  string
	value string
}

func newArgVarInstruct(instruction string) *varInstruct {
	instruction = strings.TrimPrefix(instruction, "ARG")
	splits := strings.Split(instruction, "=")
	return &varInstruct{
		name:  splits[0],
		value: strings.Join(splits[1:], "="),
	}
}

func newEnvVarInstruct(instruction string) * varInstruct {
    instruction = strings.TrimPrefix(instruction, "ENV")
}
type astInfo struct {
	args        []*varInstruct
	envs        []*varInstruct
	labels      []*varInstruct
	lastWorkdir string
	ports       []string
	cmds        []string
}

func extractDfPath(args []string) (string, error) {
	if len(os.Args) < 2 {
		return "", os.ErrNotExist
	}
	return args[1], nil
}

func runDocker(logger *shared.Logger, args []string) {
	exe := filepath.Join(filepath.Dir(args[0]), shared.DockerHiddenName)
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		logger.Error(err.Error())
	}

	cmd.Wait()
}

func buildAst(dfPath string) (*parser.Node, error) {
	df, err := os.Open(dfPath)
	if err != nil {
		return &parser.Node{}, err
	}

	res, err := parser.Parse(df)
	return res.AST, err
}

func extractAnchorPath(comment string) string {
	if !strings.Contains(comment, "|<-") {
		return ""
	}
	splits := strings.Split(comment, "|<-")
	return splits[len(splits)-1]
}

func extractInfoFromAst(ast *parser.Node) *astInfo {
	info := &astInfo{
		args:        []*varInstruct{},
		envs:        []*varInstruct{},
		labels:      []*varInstruct{},
		lastWorkdir: "",
		ports:       []string{},
		cmds:        []string{},
	}

	for _, c := range ast.Children {
		if strings.EqualFold(c.Value, "ARG") {
			arg := ""
			for n := c.Next; n != nil; n = n.Next {
				arg += n.Value
			}
			info.args = append(info.args, newArgVarInstruct(arg))
		} else if strings.EqualFold(c.Value, "ENV") {
            env := ""
            for n := c.Next; n != nil && n.Next != nil; n = n.Next{
                env += n.Value
            }
            info.args = append(info.envs, newArgVarInstruct(env))
        }
	}
	return info
}

func pullAnchors(ast *parser.Node) error {
	pulledAnchors := shared.NewSet[string]()
	var prevNode *parser.Node
	var err error
	applyToAst(ast, func(step int, node *parser.Node) bool {
		if step != INSTRUCT_ITER_STEP {
			return true
		}

		for _, c := range node.PrevComment {
			ap := extractAnchorPath(c)
			if ap == "" || pulledAnchors.Has(ap) {
				continue
			}
			pulledAnchors.Insert(ap)
			aAst, e := buildAst(ap)
			if e != nil {
				err = e
				return false
			}
			if prevNode == nil {
				err = fmt.Errorf("cannot have anchors before 'from'; offending anchor: %s", ap)
				return false
			}

			e = pullAnchors(aAst)
			if e != nil {
				err = e
				return false
			}

			prevNode.AddChild(aAst, 0, 0)
		}
		prevNode = node
		return true
	}, false)
	return err
}

func applyToAst(node *parser.Node, fn func(int, *parser.Node) bool, reverse bool) bool {
	if !fn(INSTRUCT_ITER_STEP, node) {
		return false
	}
	for n := node.Next; n != nil; n = n.Next {
		if !fn(NEXT_ITER_STEP, n) {
			return false
		}
	}

	if !reverse {
		for _, c := range node.Children {
			if !applyToAst(c, fn, reverse) {
				return false
			}
		}
	} else {
		for i := len(node.Children) - 1; i >= 0; i-- {
			if !applyToAst(node.Children[i], fn, reverse) {
				return false
			}
		}
	}

	return true
}

func astToDockerfile(ast *parser.Node) (string, error) {
	df := ""

	curInstruct := []string{}
	applyToAst(ast, func(step int, node *parser.Node) bool {
		normV := strings.TrimSpace(node.Value)
		switch step {

		case INSTRUCT_ITER_STEP:
			if len(curInstruct) > 0 {
				df += strings.Join(curInstruct, " ") + "\n"
			}
			curInstruct = []string{normV}

		case NEXT_ITER_STEP:
			curInstruct = append(curInstruct, normV)
			if strings.EqualFold(normV, "=") && strings.EqualFold(curInstruct[0], "ENV") {
				curInstruct[len(curInstruct)-1], curInstruct[len(curInstruct)-2] = curInstruct[len(curInstruct)-2], curInstruct[len(curInstruct)-1]
			}
		}
		return true
	}, false)

	df += strings.Join(curInstruct, " ") + "\n"
	return df, nil
}

func main() {
	logger := shared.NewLogger("main")
	defer runDocker(logger, os.Args)

	dfPath, err := extractDfPath(os.Args)
	if err != nil {
		logger.Error("does not contain a path to dockerfile", os.Args...)
		return
	}

	ast, err := buildAst(dfPath)
	if err != nil {
		logger.Error("failed to parse dockerfile", err.Error())
		return
	}

	err = pullAnchors(ast)
	if err != nil {
		logger.Error("failed to pull anchors", err.Error())
		return
	}
	df, _ := astToDockerfile(ast)
	fmt.Println(df)
	fmt.Println(ast.Dump())
}
