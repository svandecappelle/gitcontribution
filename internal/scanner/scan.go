package scanner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/svandecappelle/gitcontrib/internal/config"
)

// recursiveScanFolder starts the recursive search of git repositories
// living in the `folder` subtree
func recursiveScanFolder(folder string) ([]string, error) {
	return ScanGitFolders(make([]string, 0), folder)
}

// Scan scans a new folder for Git repositories
func Scan(folder string) error {
	repositories, err := recursiveScanFolder(folder)
	if err != nil {
		return err
	}
	conf, err := config.Get()
	if err != nil {
		return err
	}
	return conf.AddFilesToScan(repositories)
}

// List list all repositories wich saved to scan
func List() error {
	conf, err := config.Get()
	if err != nil {
		return err
	}
	repositories := conf.GetRepositories()
	fmt.Printf("Git folders:\n\n")
	for _, repository := range repositories {
		fmt.Printf("- %s\n", repository)
	}
	return nil
}

func ShouldBeIgnored(folderName string) bool {
	return folderName == "vendor" || folderName == "node_modules" || folderName == "venv"
}

// ScanGitFolders returns a list of subfolders of `folder` ending with `.git`.
// Returns the base folder of the repo, the .git folder parent.
// Recursively searches in the subfolders by passing an existing `folders` slice.
func ScanGitFolders(folders []string, folder string) ([]string, error) {
	// trim the last `/`
	folder = strings.TrimSuffix(folder, "/")

	f, err := os.Open(folder)
	if err != nil {
		log.Fatal(err)
		return folders, err
	}

	pathFrom, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return folders, err
	}
	pathFrom = filepath.Join(pathFrom, folder)
	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Fatal(err)
		return folders, err
	}
	for _, file := range files {
		if file.IsDir() {
			pathRelative := folder + "/" + file.Name()
			pathAbsolute := filepath.Join(pathFrom, file.Name())
			if file.Name() == ".git" {
				pathAbsolute = strings.TrimSuffix(pathAbsolute, "/.git")
				fmt.Printf("Folder %s added to scan list\n", pathAbsolute)
				folders = append(folders, pathAbsolute)
				continue
			}
			if ShouldBeIgnored(file.Name()) {
				continue
			}
			folders, err = ScanGitFolders(folders, pathRelative)
			if err != nil {
				return folders, err
			}
		}
	}

	return folders, nil
}
