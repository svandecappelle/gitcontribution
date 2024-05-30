package services

import (
	"log"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	"github.com/svandecappelle/gitcontrib/internal/interfaces"
	"github.com/svandecappelle/gitcontrib/internal/scanner"

	"github.com/svandecappelle/gitcontrib/pkg/swagger/server/models"
	"github.com/svandecappelle/gitcontrib/pkg/swagger/server/restapi"
	"github.com/svandecappelle/gitcontrib/pkg/swagger/server/restapi/operations"
)

func Start() error {
	// Initialize Swagger
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewHelloAPIAPI(swaggerSpec)
	server := restapi.NewServer(api)

	defer func() {
		if err := server.Shutdown(); err != nil {
			// error handle
			log.Fatalln(err)
		}
	}()

	server.Port = 8080

	api.CheckHealthHandler = operations.CheckHealthHandlerFunc(Health)
	api.GetHelloUserHandler = operations.GetHelloUserHandlerFunc(GetHelloUser)
	api.GetAPIReportHandler = operations.GetAPIReportHandlerFunc(GetApiReport)

	// Start server which listening
	if err := server.Serve(); err != nil {
		log.Fatalln(err)
		return err
	}
	return nil
}

// Health route returns OK
func Health(operations.CheckHealthParams) middleware.Responder {
	return operations.NewCheckHealthOK().WithPayload("Yes")
}

// GetHelloUser returns Hello + your name
func GetHelloUser(user operations.GetHelloUserParams) middleware.Responder {
	return operations.NewGetHelloUserOK().WithPayload("Hello " + user.User + "!")
}

// GetApiReport returns a statistic report
func GetApiReport(params operations.GetAPIReportParams) middleware.Responder {
	var payload = models.GitStatisticsReport{}
	user := params.Username
	opts := interfaces.LaunchOptions{
		User:             user,
		DurationInWeeks:  52,
		Folders:          []string{"."},
		Merge:            false,
		Delta:            "",
		Dashboard:        true,
		PatternToExclude: nil,
		PatternToInclude: nil,
	}
	rLaunch := scanner.Launch(opts)
	payload.DateFrom = strfmt.DateTime(rLaunch[0].BeginOfScan)
	payload.DateTo = strfmt.DateTime(rLaunch[0].EndOfScan)
	c := models.GitCommits{}
	commits := []*models.GitCommits{&c}
	payload.Commits = commits

	for _, l := range rLaunch {
		if l.Error != nil {
			continue
		}
		var additions, deletions, i int64
		authors := make([]*models.GitAuthors, l.AuthorsEditions.Len())
		l.AuthorsEditions.Range(func(user string, c interfaces.Contribution) bool {
			adds := int64(c.Additions)
			dels := int64(c.Deletions)
			additions += adds
			deletions += dels
			authors[i] = &models.GitAuthors{
				Name:      user,
				Additions: adds,
				Deletions: dels,
			}
			i += 1
			return true
		})
		for _, commit := range l.Commits {
			contrib := models.GitAuthorContributions{
				Additions: int64(additions),
				Deletions: int64(deletions),
			}
			c.Contributions = &contrib
			c.Count += int64(commit)
			c.Authors = authors
		}
	}

	return operations.NewGetAPIReportOK().WithPayload(&payload)
}
