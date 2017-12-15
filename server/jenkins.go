package server

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/bndr/gojenkins"
)

var JENKINS_STATUS = map[string]Status{
	"aborted":        ABORTED,
	"aborted_anime":  BUILDING_FROM_ABORTED,
	"blue":           SUCCESS,
	"blue_anime":     BUILDING_FROM_SUCCESS,
	"disabled":       DISABLED,
	"disabled_anime": BUILDING_FROM_DISABLED,
	"grey":           UNKNOWN,
	"grey_anime":     BUILDING_FROM_UNKNOWN,
	"notbuilt":       NOT_BUILT,
	"notbuilt_anime": BUILDING_FROM_NOT_BUILT,
	"red":            FAILURE,
	"red_anime":      BUILDING_FROM_FAILURE,
	"yellow":         UNSTABLE,
	"yellow_anime":   BUILDING_FROM_UNSTABLE,
}

type Jenkins struct {
	Server
	jobs []*gojenkins.Job
}

// StartJenkins starts the Jenkins CI Server Polling Loop
func (j *Jenkins) Start(jenkinsConfig Server) {

	log.Println("Jenkins start")

	j.Server = jenkinsConfig
	jenkins, err := gojenkins.CreateJenkins(j.URL, j.Username, j.Password).Init()
	if err != nil {
		log.Fatal("Unable to CreateJenkins. Err:", err)
	}

	for _, jb := range j.Jobs {
		job, err := jenkins.GetJob(jb.Jobname)
		if err != nil {
			log.Fatalf("Unable to GetJob(%s). Err:", jb.Jobname, err)
		}
		j.jobs = append(j.jobs, job)
	}

}

// Poll polls the CI Server to get the latest job information
func (j *Jenkins) Poll() ServerResult {

	log.Printf("Polling Jenkins")
	var s ServerResult
	var b BuildResult
	s.Result = "SUCCESS"
	for i, jenkJob := range j.jobs {
		jenkJob.Poll()
		jobResult := fmt.Sprintf("%s", JENKINS_STATUS[jenkJob.GetDetails().Color])
		b.JobName = jenkJob.GetName()
		b.Result = jobResult
		j.Jobs[i].Result = jobResult
		if jobResult != "SUCCESS" {
			s.Result = jobResult
		}
		s.BuildResults = append(s.BuildResults, b)
		j.Result = s.Result
	}

	return s
}
