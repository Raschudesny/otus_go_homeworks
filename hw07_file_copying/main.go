package main

import (
	"flag"
	"log"
)

var (
	from, to      string
	limit, offset int64
)

func init() {
	flag.StringVar(&from, "from", "", "file to read from")
	flag.StringVar(&to, "to", "", "file to write to")
	flag.Int64Var(&limit, "limit", 0, "limit of bytes to copy")
	flag.Int64Var(&offset, "offset", 0, "offset in input file")
}

func main() {
	flag.Parse()
	if len(from) == 0 {
		log.Fatalf("Error, -from flag is empty ")
	}
	if len(to) == 0 {
		log.Fatalf("Error, -to flag is empty ")
	}
	if err := Copy(from, to, offset, limit); err != nil {
		log.Fatalf("Error during file copying: %v", err)
	}
	log.Println("Done!")
}
