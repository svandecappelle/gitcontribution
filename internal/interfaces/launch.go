package interfaces

import (
	"sync"
	"time"

	syncx "github.com/svandecappelle/gitcontrib/internal/types"
)

type LaunchOptions struct {
	User             *string
	DurationInWeeks  int
	Folders          []string
	Merge            bool
	Delta            string
	Dashboard        bool
	PatternToExclude []string
	PatternToInclude []string
}

type StatsResult struct {
	Options         StatsOptions
	BeginOfScan     time.Time
	EndOfScan       time.Time
	DurationInDays  int
	Folder          string
	Commits         map[int]int
	HoursCommits    [24]int
	DayCommits      [7]int
	AuthorsEditions AuthorsStats
	Error           error
}

type AuthorsStats struct {
	syncx.Map[string, Contribution]
}

type Contribution struct {
	sync.Mutex
	Additions int
	Deletions int
}

func (a *AuthorsStats) AddContributions(user string, additions int, deletions int) Contribution {
	contributions, _ := a.LoadOrStore(user, Contribution{Additions: 0, Deletions: 0})
	contributions.Additions += additions
	contributions.Deletions += deletions
	a.Store(user, contributions)
	return contributions
}

type StatsOptions struct {
	EmailOrUsername      *string
	DurationParamInWeeks int
	Folders              []string
	Delta                string
	Silent               bool
	PatternToExclude     []string
	PatternToInclude     []string
}
