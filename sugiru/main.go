package main

import (
	"fmt"
	"os"
	"os/user"
	"sugiru/repl"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("[ SUGIRU REPL MODE : USER {%s} ]\n", user.Username)
	repl.Start(os.Stdin, os.Stdout)
}
