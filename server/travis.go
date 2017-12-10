package server

import (
	"fmt"
	"log"

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
	jobs         []gotravis.Repository
}

// Start starts the Travis CI Server Polling Loop
func (t *Travis) Start(travisConfig Server) {

	log.Println("Travis started")
	defer log.Println("Travis: caller has told us to stop")

	t.serverConfig = travisConfig
	travis := gotravis.NewClient(gotravis.TRAVIS_API_DEFAULT_URL, t.serverConfig.AccessToken)
	log.Printf("travis.IsAuthenticated() = %v \n", travis.IsAuthenticated())

	for _, tj := range t.serverConfig.Jobs {
		opt := &gotravis.RepositoryListOptions{Slug: tj.Jobname}
		job, _, err := travis.Repositories.Find(opt)
		if err != nil {
			log.Printf("Error in Repo Find %v", err)
		}
		t.jobs = append(t.jobs, job[0])
	}
}

func (t *Travis) Poll() string {

	msg := "SUCCESS"
	for _, travJob := range t.jobs {
		status := TRAVIS_STATUS[travJob.LastBuildState]
		temp := fmt.Sprintf("%s", status)
		if temp != "SUCCESS" {
			msg = fmt.Sprintf("Travis-Ci: %s Status = %s", travJob.Slug, status)
		}
	}
	return msg
}
