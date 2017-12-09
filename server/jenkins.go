package server

import (
	"context"
	"fmt"
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
	serverConfig Server
}

// StartJenkins starts the Jenkins CI Server Polling Loop
func (j *Jenkins) Start(ctx context.Context, jenkinsConfig Server) {

	log.Println(ctx, "Jenkins started")
	defer log.Println("Jenkins: caller has told us to stop")

	j.serverConfig = jenkinsConfig
	//j.stopCh = make(chan struct{})
	//log.Println("Starting Jenkins")
	jenkins, err := gojenkins.CreateJenkins(j.serverConfig.URL, j.serverConfig.Username, j.serverConfig.Password).Init()
	if err != nil {
		log.Fatal("Unable to CreateJenkins. Err:", err)
	}

	var jenkinsJobs []*gojenkins.Job
	for _, jb := range j.serverConfig.Jobs {
		job, err := jenkins.GetJob(jb.Jobname)
		if err != nil {
			log.Fatalf("Unable to GetJob(%s). Err:", jb.Jobname, err)
		}
		jenkinsJobs = append(jenkinsJobs, job)
	}
	ticker := time.NewTicker(time.Second * time.Duration(j.serverConfig.Pollrate))
	defer ticker.Stop()

	for {
		select {
		case _ = <-ticker.C:
			for _, jenkJob := range jenkinsJobs {
				jenkJob.Poll()
				status := JENKINS_STATUS[jenkJob.GetDetails().Color]
				log.Printf("Jenkins: %s Status = %s", jenkJob.GetName(), status)
			}
		case <-ctx.Done():
			fmt.Println("Jenkins Poller: caller has told us to stop")
			return
		}
	}
}
