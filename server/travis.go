package server

import (
	"context"
	"fmt"
	"log"
	"time"

	gotravis "github.com/Ableton/go-travis"
)

var TRAVIS_STATUS = map[string]Status{
	"created":  UNKNOWN,
	"queued":   UNKNOWN,
	"received": UNKNOWN,
	"started":  BUILDING_FROM_PREVIOUS_STATE,
	"passed":   SUCCESS,
	"failed":   FAILURE,
	"errored":  FAILURE,
	"canceled": ABORTED,
	"ready":    UNKNOWN,
}

type Travis struct {
	serverConfig Server
}

// Start starts the Travis CI Server Polling Loop
func (t *Travis) Start(ctx context.Context, travisConfig Server, ch chan string) {

	log.Println("Travis started")
	defer log.Println("Travis: caller has told us to stop")

	t.serverConfig = travisConfig
	travis := gotravis.NewClient(gotravis.TRAVIS_API_DEFAULT_URL, t.serverConfig.AccessToken)
	log.Printf("travis.IsAuthenticated() = %v \n", travis.IsAuthenticated())

	var repos []gotravis.Repository

	for _, tj := range t.serverConfig.Jobs {
		opt := &gotravis.RepositoryListOptions{Slug: tj.Jobname}
		repo, _, err := travis.Repositories.Find(opt)
		if err != nil {
			log.Printf("Error in Repo Find %v", err)
		}
		repos = append(repos, repo[0])
	}

	ticker := time.NewTicker(time.Second * time.Duration(t.serverConfig.Pollrate))
	defer ticker.Stop()
	for {
		select {
		case _ = <-ticker.C:
			msg := "SUCCESS"
			for _, travJob := range repos {
				status := TRAVIS_STATUS[travJob.LastBuildState]
				temp := fmt.Sprintf("%s", status)
				if temp != "SUCCESS" {
					msg = fmt.Sprintf("Travis-Ci: %s Status = %s", travJob.Slug, status)
				}
			}
			ch <- fmt.Sprintf("Travis: %s", msg)
		case <-ctx.Done():
			log.Println("Travis Poller: caller has told us to stop")
			return
		}
	}
}
