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
	travis       *gotravis.Client
}

// Start starts the Travis CI Server Polling Loop
func (t *Travis) Start(travisConfig Server) {

	log.Println("Travis start")

	t.serverConfig = travisConfig
	t.travis = gotravis.NewClient(gotravis.TRAVIS_API_DEFAULT_URL, t.serverConfig.AccessToken)
	log.Printf("travis.IsAuthenticated() = %v \n", t.travis.IsAuthenticated())

	log.Printf("Adding %d jobs", len(t.serverConfig.Jobs))
	for _, tj := range t.serverConfig.Jobs {
		opt := &gotravis.RepositoryListOptions{Slug: tj.Jobname}
		job, _, err := t.travis.Repositories.Find(opt)
		t.travis.Repositories.Get(job[0].Id)
		if err != nil {
			log.Printf("Error in Repo Find %v", err)
		}
		if len(job) > 0 {
			t.jobs = append(t.jobs, job[0])
		} else {
			log.Printf("Travis: Could not find Repo %s", tj.Jobname)
		}
	}
}

// Poll polls the CI Server to get the latest job information
func (t *Travis) Poll() ServerResult {

	log.Printf("Polling Travis")
	var s ServerResult
	var b BuildResult
	s.Result = "SUCCESS"
	for _, travJob := range t.jobs {
		job, _, err := t.travis.Repositories.Get(travJob.Id)
		if err != nil {
			log.Printf("Error in Repo Get %v", err)
		}
		if job == nil {
			log.Printf("Could not find Repo %v", travJob.Id)
		}
		jobResult := fmt.Sprintf("%s", TRAVIS_STATUS[job.LastBuildState])
		b.JobName = job.Slug
		b.Result = jobResult
		if jobResult != "SUCCESS" {
			s.Result = jobResult
		}
		s.BuildResults = append(s.BuildResults, b)
	}
	return s
}
