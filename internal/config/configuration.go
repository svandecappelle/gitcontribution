package config

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/go-git/go-git/v5"
)

func isRepo(path string) bool {
	_, err := git.PlainOpen(path)
	return err == nil
}

type Configuration struct {
	Dotfile string
}

func (c *Configuration) AddFilesToScan(files []string) error {
	return addNewSliceElementsToFile(c.Dotfile, files)
}

func (c *Configuration) GetRepositories() []string {
	repositories := parseFileLinesToSlice(c.Dotfile)
	return repositories
}

func Get() (*Configuration, error) {
	fp, err := GetDotFilePath()
	if err != nil {
		return nil, err
	}
	return &Configuration{
		Dotfile: *fp,
	}, nil
}

// GetFolders returns all the folders needs to be scanned saved in dotfile
func GetFolders() ([]string, error) {
	var repos []string
	filePath, err := GetDotFilePath()
	if err != nil {
		return []string{}, err
	}
	repos = parseFileLinesToSlice(*filePath)
	if len(repos) == 0 || isRepo(".") {
		repos = []string{"."}
	}

	return repos, nil
}

// GetDotFilePath returns the dot file for the repos list.
// Creates it and the enclosing folder if it does not exist.
func GetDotFilePath() (*string, error) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	dotFile := usr.HomeDir + "/.gogitstats"

	return &dotFile, nil
}

// openFile opens the file located at `filePath`. Creates it if not existing.
func openFile(filePath string) *os.File {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_RDWR, 0755)
	if err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			_, err = os.Create(filePath)
			if err != nil {
				panic(err)
			}
			f, err = os.OpenFile(filePath, os.O_APPEND|os.O_RDWR, 0755)
			if err != nil {
				panic(err)
			}
		} else {
			// other error
			panic(err)
		}
	}

	return f
}

// parseFileLinesToSlice given a file path string, gets the content
// of each line and parses it to a slice of strings.
func parseFileLinesToSlice(configFilePath string) []string {
	f := openFile(configFilePath)
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			panic(err)
		}
	}

	return lines
}

// sliceContains returns true if `slice` contains `value`
func sliceContains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// joinSlices adds the element of the `new` slice
// into the `existing` slice, only if not already there
func joinSlices(new []string, existing []string) []string {
	for _, i := range new {
		if !sliceContains(existing, i) {
			existing = append(existing, i)
		}
	}
	return existing
}

// dumpStringsSliceToFile writes content to the file in path `filePath` (overwriting existing content)
func dumpStringsSliceToFile(repos []string, filePath string) error {
	content := strings.Join(repos, "\n")
	return os.WriteFile(filePath, []byte(content), 0755)
}

// addNewSliceElementsToFile given a slice of strings representing paths, stores them
// to the filesystem
func addNewSliceElementsToFile(filePath string, newRepos []string) error {
	existingRepos := parseFileLinesToSlice(filePath)
	repos := joinSlices(newRepos, existingRepos)
	return dumpStringsSliceToFile(repos, filePath)
}
