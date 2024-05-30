package scanner_test

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/svandecappelle/gitcontrib/internal/config"
	"github.com/svandecappelle/gitcontrib/internal/scanner"
)

func TestAddRepoToScanFromFolder(tt *testing.T) {
	t := td.NewT(tt)
	folders, err := config.GetFolders()
	t.CmpNoError(err)
	t.Cmp(folders, td.Len(1))
	t.Cmp(folders[0], ".")
}

func TestUserDotFile(tt *testing.T) {
	t := td.NewT(tt)
	file, err := config.GetDotFilePath()

	t.CmpNoError(err)
	t.Cmp(file, td.NotNil())
}

func TestIgnoreFolders(tt *testing.T) {
	t := td.NewT(tt)
	t.True(scanner.ShouldBeIgnored("node_modules"), "node_modules should be ignored from stats")
	t.True(scanner.ShouldBeIgnored("venv"), "venv should be ignored from stats")
	t.True(scanner.ShouldBeIgnored("vendor"), "vendor folder should be ignored from stats")
	t.False(scanner.ShouldBeIgnored("tests"), "tests folder should not be ignored from stats")
	t.False(scanner.ShouldBeIgnored("src"), "src folder and others should not be ignored from stats")
}

func TestListReposNoConfig(tt *testing.T) {
	t := td.NewT(tt)

	err := scanner.List()
	t.CmpNoError(err)
}

func TestLunchStatsNoConfig(tt *testing.T) {
	t := td.NewT(tt)

	folders, err := scanner.ScanGitFolders([]string{}, "..")
	t.CmpNoError(err)
	t.Cmp(folders, td.Len(1))
}
