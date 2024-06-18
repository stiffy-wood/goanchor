package compile

import (
	"goanchor/src/collections"
	"strings"
)

func getInstructions() *collections.Set[string] {
	return collections.NewSet[string](
		"add",
		"arg",
		"cmd",
		"copy",
		"entrypoint",
		"env",
		"expose",
		"from",
		"healthcheck",
		"label",
		"maintainer",
		"onbuild",
		"run",
		"shell",
		"stopsignal",
		"user",
		"volume",
		"workdir",
		"as",
	)
}

func getKeywords() *collections.Set[string] {
	return collections.NewSet[string](
		"add",
		"arg",
		"cmd",
		"copy",
		"entrypoint",
		"env",
		"expose",
		"from",
		"healthcheck",
		"label",
		"maintainer",
		"onbuild",
		"run",
		"shell",
		"stopsignal",
		"user",
		"volume",
		"workdir",
		"as",
		"->|",
		"#",
		",",
		"=",
		"--",
		"[",
		"]",
		"\\",
		"\"",
		"'",
		"`",
	)
}

func TokenizeLayer(layer string) []string {
	keywords := getKeywords()
	var tokens []string
	curToken := ""
	quotes := make(map[string]bool)

	addCurToken := func() {
		normTk := strings.TrimSpace(curToken)
		if normTk != "" {
			tokens = append(tokens, normTk)
			curToken = ""
		}
	}

	for i, r := range layer {
		sVal := string(r)

		// Check if it s inside a quote
		if sVal == "\"" || sVal == "'" || sVal == "`" {
			if i == 0 {
				quotes[sVal] = true
			} else {
				quotes[sVal] = !quotes[sVal]
			}
		}

		isQuoted := false
		for _, isOpenQuote := range quotes {
			if isOpenQuote {
				isQuoted = true
				break
			}
		}

		// if it is not in in a quote, do some logic to decide if it's a separate token or not
		if !isQuoted {
			sVal = strings.TrimSpace(sVal)
			if keywords.Has(strings.ToLower(curToken)) {
				addCurToken()
			}
		}

		curToken += sVal
	}

	addCurToken()

	return tokens
}
