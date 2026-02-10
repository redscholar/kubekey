package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kubesphere/kubekey/v4/cmd/apiserver/app"
)

func main() {
	if err := app.NewAPIServerCommand().Execute(); err != nil {
		vFlag := flag.Lookup("v")
		if vFlag != nil {
			fmt.Printf("%+v", err)
		} else {
			fmt.Printf("%v", err)
		}
		os.Exit(1)
	}
}
