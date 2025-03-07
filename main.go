package main

import (
	"os"

	"github.com/arisu-archive/assets-dumper/cmd/root"
)

const Version = "0.0.1"

func main() {
	root.Execute(root.ExecuteConfig{
		Version: Version,
		Exit:    os.Exit,
		In:      os.Stdin,
		Out:     os.Stdout,
		Err:     os.Stderr,
		Args:    os.Args[1:],
	})
}
