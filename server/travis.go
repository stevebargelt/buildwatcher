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
func (t *Travis) Start(ctx context.Context, travisConfig Server) {

	log.Println(ctx, "Travis started")
	defer log.Println("Travis: caller has told us to stop")

	t.serverConfig = travisConfig
	//t.stopCh = make(chan struct{})
	//log.Println("Starting Travis")
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
	log.Printf("Found %d repos", len(repos))

	ticker := time.NewTicker(time.Second * time.Duration(t.serverConfig.Pollrate))
	defer ticker.Stop()

	for {
		select {
		case _ = <-ticker.C:
			for _, travJob := range repos {
				status := TRAVIS_STATUS[travJob.LastBuildState]
				log.Printf("Travis-Ci: %s Status = %s", travJob.Slug, status)
			}
		case <-ctx.Done():
			fmt.Println("Travis Poller: caller has told us to stop")
			return
		}
	}
}
