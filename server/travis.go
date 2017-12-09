package server

import (
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
	stopCh       chan struct{}
	serverConfig Server
}

func NewTravis(travisConfig Server) *Travis {
	return &Travis{
		serverConfig: travisConfig,
	}
}

// StartTravis starts the Travis CI Server Polling Loop
func (t *Travis) StartTravis() {
	t.stopCh = make(chan struct{})
	log.Println("Starting Travis")
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
	go func() {
		for _ = range ticker.C {
			for _, travJob := range repos {
				status := TRAVIS_STATUS[travJob.LastBuildState]
				log.Printf("Travis-Ci: %s Status = %s", travJob.Slug, status)
			}
		}
	}()

	select {
	case <-t.stopCh:
		log.Println("Stopping Travis polling")
		return
	}

}

func (t *Travis) Stop() {
	if t.stopCh == nil {
		log.Println("WARNING: stop channel is not initialized.")
		return
	}
	t.stopCh <- struct{}{}
	log.Println("Stopped Travis")
}
