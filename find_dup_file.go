package main

import (
	"crypto/sha512"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var files = make(map[[sha512.Size]byte]string)

func checkDuplicate(path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if info.IsDir() { // skip directory
		return nil
	}

	data, err := ioutil.ReadFile(path)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	hash := sha512.Sum512(data) // get the file sha512 hash

	if v, ok := files[hash]; ok {
		fmt.Printf("%q is a duplicate of %q\n", path, v)
	} else {
		files[hash] = path // store in map for comparison
	}

	return nil
}

func main() {

	if len(os.Args) != 2 {
		fmt.Printf("USAGE : %s <target_directory> \n", os.Args[0])
		os.Exit(0)
	}

	dir := os.Args[1] // get the target directory

	err := filepath.Walk(dir, checkDuplicate)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
