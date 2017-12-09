package server

import (
	"log"
	"time"

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
	stopCh       chan struct{}
	serverConfig Server
}

func NewJenkins(jenkinsConfig Server) *Jenkins {
	return &Jenkins{
		serverConfig: jenkinsConfig,
	}
}

// StartJenkins starts the Jenkins CI Server Polling Loop
func (j *Jenkins) StartJenkins() {
	j.stopCh = make(chan struct{})
	log.Println("Starting Jenkins")
	jenkins, err := gojenkins.CreateJenkins(j.serverConfig.URL, j.serverConfig.Username, j.serverConfig.Password).Init()
	if err != nil {
		log.Fatal("Unable to CreateJenkins. Err:", err)
	}

	var JenkinsJobs []*gojenkins.Job
	for _, jb := range j.serverConfig.Jobs {
		job, err := jenkins.GetJob(jb.Jobname)
		if err != nil {
			log.Fatalf("Unable to GetJob(%s). Err:", jb.Jobname, err)
		}
		JenkinsJobs = append(JenkinsJobs, job)
	}

	ticker := time.NewTicker(time.Second * time.Duration(j.serverConfig.Pollrate))
	go func() {
		for _ = range ticker.C {
			for _, jenkJob := range JenkinsJobs {
				jenkJob.Poll()
				status := JENKINS_STATUS[jenkJob.GetDetails().Color]
				log.Printf("Jeankis: %s Status = %s", jenkJob.GetName(), status)
			}
		}
	}()

	select {
	case <-j.stopCh:
		log.Println("Stopping Jenkins polling")
		return
	}

}

func (j *Jenkins) Stop() {
	if j.stopCh == nil {
		log.Println("WARNING: stop channel is not initialized.")
		return
	}
	j.stopCh <- struct{}{}
	log.Println("Stopped Jenkins")
}
