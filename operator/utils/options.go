package utils

import (
	"encoding/json"
	"os"
)

// Opts options struct for the operator to autopopulate help templates, autogenerate options, and ensure consistency between env and cli.
type Opts struct {
	MetricsAddr          string `arg:"--metrics-bind-address,env" default:":8080" json:"metricsAddr,omitempty" help:"The address the metric endpoint binds to."`
	LeaderElectionID     string `arg:"--leader-election-id,env" default:"d460f2c2.goauthentik.io" json:"leaderElectionID,omitempty" help:"Lease name to use for leader election."`
	WatchesPath          string `arg:"--watches-file,env" default:"watches.yaml" json:"watchesPath,omitempty" help:"Path to watches file."`
	ProbeAddr            string `arg:"--health-probe-bind-address,env" default:":8081" json:"probeAddr,omitempty" help:"The address the probe endpoint binds to."`
	EnableLeaderElection bool   `arg:"--leader-elect,env" json:"enableLeaderElection,omitempty" help:"To elect a leader to be active else all active."`
	OperatorNamespace    string `arg:"--operator-namespace,env" default:"auth" json:"operatorNamespace,omitempty" help:"The operators namespace for leader election."`
	WatchedNamespace     string `arg:"--watched-namespace,env" default:"" json:"watchedNamespace,omitempty" help:"The operators watched namespace. Defaults to empty (which watches all)."`
	Debug                bool   `arg:"-d,--debug,env" json:"debug,omitempty" help:"We should run in debug mode."`
	Port                 int    `arg:"-p,--port,env" default:"9443" json:"port,omitempty" help:"What port should the controller bind to."`
	AppVersion           string `arg:"--app-version,required,env:APP_VERSION" json:"appVersion,omitempty" help:"version of the operated on app."`
	SrcVersion           string `arg:"--source-version,required,env:SRC_VERSION" json:"srcVersion,omitempty" help:"version of the operator."`
}

func PrettyPrint(i interface{}) (string, error) {
	s, err := json.MarshalIndent(i, "", "\t")
	if err != nil {
		return "", err
	}
	return string(s), nil
}

// exists returns whether the given file or directory exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
