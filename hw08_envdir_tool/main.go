package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalln("Expected at least two arguments")
	}
	envs, err := ReadDirEnvs(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	result := RunCmd(os.Args[2:], envs, CmdIOStreams{os.Stdin, os.Stdout, os.Stderr})
	os.Exit(result)
}
