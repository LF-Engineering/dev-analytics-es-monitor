package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fatalf("%s: you must secify path to fixtures diectory as an argument\n", os.Args[0])
	}
	fixtures := getFixtures(os.Args[1])
	fmt.Printf("%v\n", fixtures)
}
