package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"sort"
	. "strings"
	"time"
)

var (
	flagForce   = flag.Bool("f", false, "force, don't ask for confirmation, assume yes")
	flagVerbose = flag.Bool("v", false, "verbose, show actions for all files")
	flagDryRun  = flag.Bool("d", false, "dry run, displays the operations that would be performed without actually running them")
)

var (
	randomReader  *randByteMaker
	fileCounter   int
	baseDirectory string
	directories   Directories
)

// randByteMaker wraps a rand.Source with Read() so that is can be used as a io.Reader
type randByteMaker struct {
	src rand.Source
}

func (r *randByteMaker) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(r.src.Int63() & 0xff) // mask to only the first 255 byte values
	}
	return len(p), nil
}

// Directories is a named type of []string that is a Sortable so that we can
// sort this in the order of length of the string, with the longest strings first
type Directories []string

func (a Directories) Len() int           { return len(a) }
func (a Directories) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Directories) Less(i, j int) bool { return len(a[i]) > len(a[j]) }

func main() {

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "usage: %s [-f | -v | -d ] ./path/to/directory\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	baseDirectory = flag.Arg(0)
	if info, err := os.Stat(baseDirectory); err != nil || !info.IsDir() {
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n\n", err)
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "%s is not a directory\n\n", baseDirectory)
		}
		flag.Usage()
		os.Exit(1)
	}

	randomReader = &randByteMaker{
		rand.NewSource(time.Now().Unix()),
	}

	if !*flagForce {
		fmt.Printf("This is a very destructive action that will overwrite every file in '%s'\n", baseDirectory)
		if !askForConfirmation() {
			return
		}
	}

	fmt.Printf("Scrubbing files\n")
	if err := filepath.Walk(baseDirectory, fileScrubber); err != nil {
		fmt.Printf("\n%s\n", err)
		os.Exit(1)
	}
	fmt.Printf("\nScrubbed %d files\n", fileCounter)

	fmt.Printf("\nScrubbing directories\n")
	// sort folders in the order of string size so we rename the leaf "child"
	// folders first
	sort.Sort(directories)
	for i, currentPath := range directories {
		if currentPath == baseDirectory {
			continue
		}
		newPath := fmt.Sprintf("%s/dir_%d", path.Dir(currentPath), i+1)
		if *flagVerbose {
			fmt.Printf("renaming %s to %s\n", currentPath, newPath)
		} else {
			fmt.Print(".")
		}
		if !*flagDryRun {
			if err := os.Rename(currentPath, newPath); err != nil {
				fmt.Printf("err: %s\n", err)
			}
		}
	}
	fmt.Printf("\nScrubbed %d directories\n", len(directories))
}

func fileScrubber(filePath string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if f.IsDir() {
		directories = append(directories, filePath)
		return nil
	}

	mode := f.Mode()

	if !mode.IsRegular() {
		fmt.Printf("\n%s skipped because not regular file\n", filePath)
		return nil
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	if *flagVerbose {
		fmt.Printf("scrubbing %s with %d bytes\n", filePath, f.Size())
	}
	if !*flagDryRun {
		if _, err = io.CopyN(out, randomReader, f.Size()); err != nil {
			return err
		}
	}

	fileCounter++
	newName := fmt.Sprintf("%s/file_%d%s", path.Dir(filePath), fileCounter, path.Ext(filePath))
	if *flagVerbose {
		fmt.Printf("renaming %s to %s\n", filePath, newName)
	} else {
		fmt.Print(".")
	}
	if !*flagDryRun {
		if err := os.Rename(filePath, newName); err != nil {
			return err
		}
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
	if HasPrefix(ToLower(response), "y") {
		return true
	} else if HasPrefix(ToLower(response), "n") {
		return false
	}
	return askForConfirmation()
}
