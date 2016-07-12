package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type dataMaker struct {
	src rand.Source
}

func (r *dataMaker) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(r.src.Int63() & 0xff)
	}
	return len(p), nil
}

type Directories []string

func (a Directories) Len() int           { return len(a) }
func (a Directories) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Directories) Less(i, j int) bool { return len(a[i]) > len(a[j]) }

var (
	randomiser  *dataMaker
	fileCounter int
	baseFolder  string
	directories Directories
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println("Usage: randomiser ./path/to/folder")
		os.Exit(1)
	}

	randomiser = &dataMaker{rand.NewSource(1028890720402726901)}

	baseFolder = args[0]

	fmt.Printf("This is a very destructive action that will overwrite every file in '%s'\n", baseFolder)
	if !askForConfirmation() {
		return
	}

	fmt.Printf("\nAnonymising files\n\n")

	err := filepath.Walk(baseFolder, fileMangler)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nAnonymising folders\n\n")

	// sort folders in the order of string size so we rename "child" folder
	// first
	sort.Sort(directories)
	for i, dirPath := range directories {
		if dirPath == baseFolder {
			continue
		}
		newName := fmt.Sprintf("%s/dir_%d%s", path.Dir(dirPath), i+1, path.Ext(dirPath))
		fmt.Printf("%s renamed to %s\n", dirPath, newName)
		if err := os.Rename(dirPath, newName); err != nil {
			fmt.Printf("err: %s\n", err)
			os.Exit(1)
		}
	}
}

func fileMangler(filePath string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if f.IsDir() {
		directories = append(directories, filePath)
		return nil
	}

	mode := f.Mode()

	if !mode.IsRegular() {
		fmt.Printf("%s skipped because not regulary file", filePath)
		return nil
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	fmt.Printf("%s overwriting with %d bytes\n", filePath, f.Size())
	if _, err = io.CopyN(out, randomiser, f.Size()); err != nil {
		return err
	}

	fileCounter++
	newName := fmt.Sprintf("%s/file_%d%s", path.Dir(filePath), fileCounter, path.Ext(filePath))
	fmt.Printf("%s renamed to %s\n", filePath, newName)
	if err := os.Rename(filePath, newName); err != nil {
		return err
	}

	return nil
}

func askForConfirmation() bool {
	fmt.Printf("Would you like to continue? (y/n)? ")
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}
	if strings.ToLower(string(response[0])) == "y" {
		return true
	} else if strings.ToLower(string(response[0])) == "n" {
		return false
	}

	return askForConfirmation()
}
