package main

import (
	"io/ioutil"
	"os"
)

func main() {
	f, err := os.Open("data.yaml")
	if err != nil {
		panic(err)
	}
	_, err = ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
}
