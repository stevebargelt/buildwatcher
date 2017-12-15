package server

import (
	"fmt"

	log "github.com/sirupsen/logrus"

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
	Server
	jobs   []gotravis.Repository
	travis *gotravis.Client
}

// Start starts the Travis CI Server Polling Loop
func (t *Travis) Start(travisConfig Server) {

	log.Println("Travis start")

	t.Server = travisConfig
	t.travis = gotravis.NewClient(gotravis.TRAVIS_API_DEFAULT_URL, t.AccessToken)
	log.Printf("travis.IsAuthenticated() = %v \n", t.travis.IsAuthenticated())

	log.Printf("Adding %d jobs", len(t.Jobs))
	for _, tj := range t.Jobs {
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
func (t *Travis) Poll() bool {

	log.Printf("Polling Travis")
	t.Result = "SUCCESS"
	for i, travJob := range t.jobs {
		job, _, err := t.travis.Repositories.Get(travJob.Id)
		if err != nil {
			log.Printf("Error in Repo Get %v", err)
		}
		if job == nil {
			log.Printf("Could not find Repo %v", travJob.Id)
		}
		jobResult := fmt.Sprintf("%s", TRAVIS_STATUS[job.LastBuildState])
		t.Jobs[i].Result = jobResult
		if jobResult != "SUCCESS" {
			t.Result = jobResult
		}

	}

	return true
}

// Status returns the Status of the entire server
func (t *Travis) Status() string {
	return t.Result
}

// JobStatus returns a string array of all the Job results form last Poll
func (t *Travis) JobStatus() []string {
	var jobResults []string
	for _, job := range t.Jobs {
		jobResults = append(jobResults, job.Result)
	}
	return jobResults
}
