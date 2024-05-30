package scanner

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/schollz/progressbar/v3"

	"github.com/svandecappelle/gitcontrib/internal/dashboard"
	"github.com/svandecappelle/gitcontrib/internal/date"
	"github.com/svandecappelle/gitcontrib/internal/interfaces"
)

var DefaultDurationInDays = 365

// TODO use an interface object in order to refacto in same place the statistic run logic and then print results

func Launch(opts interfaces.LaunchOptions) []*interfaces.StatsResult {
	var results []*interfaces.StatsResult = []*interfaces.StatsResult{}
	var wg sync.WaitGroup
	bar := progressbar.Default(-1, "Analyzing commits")

	if opts.Merge {
		options := interfaces.StatsOptions{
			EmailOrUsername:      opts.User,
			DurationParamInWeeks: opts.DurationInWeeks,
			Folders:              opts.Folders,
			Delta:                opts.Delta,
			Silent:               opts.Dashboard,
			PatternToExclude:     opts.PatternToExclude,
			PatternToInclude:     opts.PatternToInclude,
		}

		r := &interfaces.StatsResult{
			Options: options,
		}
		populateDurationInDays(opts, r)

		results = append(results, r)
		wg.Add(1)
		go Stats(r, &wg, bar)
	} else {
		for _, folder := range opts.Folders {
			options := interfaces.StatsOptions{
				EmailOrUsername:      opts.User,
				DurationParamInWeeks: opts.DurationInWeeks,
				Folders:              []string{folder},
				Delta:                opts.Delta,
				Silent:               opts.Dashboard,
				PatternToExclude:     opts.PatternToExclude,
				PatternToInclude:     opts.PatternToInclude,
			}
			r := &interfaces.StatsResult{
				Options: options,
			}
			populateDurationInDays(opts, r)
			results = append(results, r)
			wg.Add(1)
			go Stats(r, &wg, bar)
		}
	}
	wg.Wait()

	for _, r := range results {
		if !opts.Dashboard {
			fmt.Println()
			dashboard.PrintResult(r)
		}
	}

	return results
}

func populateDurationInDays(options interfaces.LaunchOptions, r *interfaces.StatsResult) {
	nowDate := time.Now()
	end := nowDate

	delta := options.Delta
	switch {
	case strings.Contains(delta, "y"):
		value, err := strconv.Atoi(strings.Split(delta, "y")[0])
		if err != nil {
			r.Error = errors.New("error delta is not a number")
			return
		}
		if value > 0 {
			value = -value
		}
		end = nowDate.AddDate(value, 0, 0)
	case strings.Contains(delta, "m"):
		value, err := strconv.Atoi(strings.Split(delta, "m")[0])
		if err != nil {
			r.Error = errors.New("error delta is not a number")
			return
		}
		if value > 0 {
			value = -value
		}
		end = nowDate.AddDate(0, value, 0)
	case strings.Contains(delta, "w"):
		value, err := strconv.Atoi(strings.Split(delta, "w")[0])
		if err != nil {
			r.Error = errors.New("error delta is not a number")
			return
		}
		if value > 0 {
			value = -value
		}
		end = nowDate.AddDate(0, 0, value*7)
	case strings.Contains(delta, "d"):
		value, err := strconv.Atoi(strings.Split(delta, "d")[0])
		if err != nil {
			r.Error = errors.New("error delta is not a number")
			return
		}
		if value > 0 {
			value = -value
		}
		end = nowDate.AddDate(0, 0, value)
	default:
		if delta != "" {
			r.Error = errors.New("invalid delta value use the format: <int>[y/m/w/d]")
			return
		}
	}
	durationInDays := DefaultDurationInDays
	if options.DurationInWeeks > 0 {
		durationInDays = options.DurationInWeeks * 7
	}
	r.DurationInDays = durationInDays
	r.EndOfScan = end
	r.BeginOfScan = end.AddDate(0, 0, -durationInDays)
	if int(r.BeginOfScan.Weekday()) != 1 {
		// Not a monday
		// offset := math.Max(0, float64(6-int(r.BeginOfScan.Weekday())))
		offset := -1 * (int(r.BeginOfScan.Weekday()) - 1)
		r.BeginOfScan = date.GetBeginningOfDay(r.BeginOfScan.AddDate(0, 0, offset))

		r.EndOfScan = date.GetEndOfDay(r.EndOfScan.AddDate(0, 0, offset+6))
		daysBetween := r.EndOfScan.Sub(r.BeginOfScan).Hours() / 24
		r.DurationInDays = int(daysBetween)
	}
}

// Stats calculates and prints the stats.
func Stats(r *interfaces.StatsResult, wg *sync.WaitGroup, bar *progressbar.ProgressBar) {
	defer wg.Done()
	err := processRepositories(r, bar)

	if err != nil {
		r.Error = err
		return
	}

	r.Folder = strings.Join(r.Options.Folders, ",")
}

// fillCommits given a repository found in `path`, gets the commits and
// puts them in the `commits` map, returning it when completed
func fillCommits(r *interfaces.StatsResult, emailOrUsername *string, path string, bar *progressbar.ProgressBar) error {
	// instantiate a git repo object from path
	repo, err := git.PlainOpen(path)
	if err != nil {
		// log.Fatalf("Cannot get stat from folder (not a repository): %s", path)
		return fmt.Errorf("cannot get stat from folder (not a repository): %s", path)
	}
	// Remove one day to end date to be sure parse today date
	// trueEndDateParse := endDate.AddDate(0, 0, 1)
	// get the commits history until endDate is not reached
	iterator, err := repo.Log(&git.LogOptions{Since: &r.BeginOfScan, Until: &r.EndOfScan})
	if err != nil {
		log.Fatalf("Cannot get repository history: %s", err)
		return err
	}
	// iterate the commits
	offset := calcOffset(r.EndOfScan)
	err = iterator.ForEach(func(c *object.Commit) error {
		daysAgo := date.CountDaysSinceDate(c.Author.When, r) + offset
		hour := c.Author.When.Hour()
		day := int(c.Author.When.Weekday())
		if daysAgo == date.OutOfRange {
			return nil
		}

		if emailOrUsername != nil {
			users := strings.Split(*emailOrUsername, ",")
			var found bool
			for _, u := range users {
				if strings.Contains(u, "@") && c.Author.Email == u {
					found = true
					break
				} else if c.Author.Name == u {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}

		// TODO find a solution for improve perf
		stats, _ := c.Stats()
		wg := sync.WaitGroup{}
		for _, stat := range stats {
			wg.Add(1)
			go func() {
				for _, pattern := range r.Options.PatternToExclude {
					pR, eRegex := regexp.Compile(pattern)
					if eRegex != nil {
						log.Fatalf("Input regex is not valid")
					}
					if pR.MatchString(stat.Name) {
						wg.Done()
						return
					}
				}
				for _, pattern := range r.Options.PatternToInclude {
					pR, eRegex := regexp.Compile(pattern)
					if eRegex != nil {
						log.Fatalf("Input regex is not valid")
					}
					if !pR.MatchString(stat.Name) {
						wg.Done()
						return
					} else {
						break
					}
				}

				r.AuthorsEditions.AddContributions(c.Author.Name, stat.Addition, stat.Deletion)
				wg.Done()
			}()
		}
		wg.Wait()

		if daysAgo <= r.DurationInDays {
			r.Commits[daysAgo] = r.Commits[daysAgo] + 1
			r.HoursCommits[hour] = r.HoursCommits[hour] + 1
			r.DayCommits[day] = r.DayCommits[day] + 1
		}
		_ = bar.Add(1)
		return nil
	})
	if err != nil {
		log.Fatalf("Error on git-log iterate: %s", err)
		return err
	}

	return nil
}

// processRepositories given an user email, returns the
// commits made in the last 6 months
func processRepositories(r *interfaces.StatsResult, bar *progressbar.ProgressBar) error {
	daysInMap := r.DurationInDays

	r.Commits = make(map[int]int, daysInMap)
	var errReturn error
	for i := daysInMap; i > 0; i-- {
		r.Commits[i] = 0
	}

	for _, path := range r.Options.Folders {
		err := fillCommits(r, r.Options.EmailOrUsername, path, bar)
		if err != nil {
			// continue for other folders
			// TODO rename printer and error
			dashboard.Print(dashboard.Error, fmt.Sprintf("\nError scanning folder repository %s: %s\n", path, err))
			errReturn = err
			continue
		}
	}
	return errReturn
}

// calcOffset determines and returns the amount of days missing to fill
// the last row of the stats graph
func calcOffset(endDate time.Time) int {
	var offset int
	weekday := endDate.Weekday()

	switch weekday {
	case time.Sunday:
		offset = 7
	case time.Monday:
		offset = 6
	case time.Tuesday:
		offset = 5
	case time.Wednesday:
		offset = 4
	case time.Thursday:
		offset = 3
	case time.Friday:
		offset = 2
	case time.Saturday:
		offset = 1
	}

	return offset
}
