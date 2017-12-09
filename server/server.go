package server

import (
	"context"
)

type Job struct {
	Name    string `json:"name"`
	Jobname string `json:"jobname"`
	Branch  string `json:"branch"`
}

type Server struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Username    string `json:"username"`
	AccessToken string `json:"accesstoken"`
	Password    string `json:"password"`
	URL         string `json:"url"`
	Pollrate    int    `json:"pollrate"`
	Jobs        []Job  `yaml:"jobs"`
}

// Status is the integer rep of the build status
type Status int

// Statuses
const (
	UNKNOWN Status = iota
	SUCCESS
	FAILURE
	ABORTED
	DISABLED
	UNSTABLE
	NOT_BUILT
	BUILDING_FROM_UNKNOWN
	BUILDING_FROM_SUCCESS
	BUILDING_FROM_FAILURE
	BUILDING_FROM_ABORTED
	BUILDING_FROM_DISABLED
	BUILDING_FROM_UNSTABLE
	BUILDING_FROM_NOT_BUILT
	BUILDING_FROM_PREVIOUS_STATE
	POLL_ERROR
	INVALID_STATUS
)

var statuses = [...]string{
	"UNKNOWN",
	"SUCCESS",
	"FAILURE",
	"ABORTED",
	"DISABLED",
	"UNSTABLE",
	"NOT_BUILT",
	"BUILDING_FROM_UNKNOWN",
	"BUILDING_FROM_SUCCESS",
	"BUILDING_FROM_FAILURE",
	"BUILDING_FROM_ABORTED",
	"BUILDING_FROM_DISABLED",
	"BUILDING_FROM_UNSTABLE",
	"BUILDING_FROM_NOT_BUILT",
	"BUILDING_FROM_PREVIOUS_STATE",
	"POLL_ERROR",
	"INVALID_STATUS",
}

func (s Status) String() string {
	return statuses[s]
}

type server interface {
	Start(context.Context, Server)
	Stop()
}
