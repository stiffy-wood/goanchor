package compile

type node struct {
	name     string
	value    string
	children []node
}

type syntaxTree struct {
	topNode *node
}


