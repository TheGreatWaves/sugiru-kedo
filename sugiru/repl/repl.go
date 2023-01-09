package repl

import (
	"bufio"
	"fmt"
	"io"
	"sugiru/lexer"
	"sugiru/token"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()

		// If nothing is scanned, we simply end
		if !scanned {
			return
		}

		// Retrieve the text scanned
		line := scanner.Text()

		l := lexer.New(line)

		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			fmt.Printf("%+v\n", tok)
		}
	}
}
