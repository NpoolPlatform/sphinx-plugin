package main

import "fmt"

var mmp = map[string]map[string]string{}

func main() {
	mmp["x"] = map[string]string{"xx": "x"}
	fmt.Println(mmp)

	v, ok := mmp["x"]["xx"]
	fmt.Println(v, ok)
}
