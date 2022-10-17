package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"

	"github.com/ktr0731/go-fuzzyfinder"
)

var (
	projectId   = flag.String("p", "", "Google Cloud Project Id")
	updateCache = flag.Bool("u", false, "Update dashboards list from gcloud command")
)

var (
	homeDir    string
	baseDir    string
	dashboards []*Dashboard
	// find dashboardid from projects/${projectnum}/dashboards/${dashboardid}
	r = regexp.MustCompile(`\/[a-zA-Z\d\-]+$`)
)

type Dashboard struct {
	DisplayName      string `json:"displayName"`
	Name             string `json:"name"`
	defaultDashboard bool
}

var defaultDashboards = []*Dashboard{
	{
		DisplayName:      "App Engine",
		Name:             "gae_application",
		defaultDashboard: true,
	},
	{
		DisplayName:      "BigQuery",
		Name:             "bigquery_dataset",
		defaultDashboard: true,
	},
	{
		DisplayName:      "Cloud Spanner",
		Name:             "spanner_instance",
		defaultDashboard: true,
	},
	{
		DisplayName:      "Cloud SQL",
		Name:             "cloudsql_database",
		defaultDashboard: true,
	},
	{
		DisplayName:      "Cloud Strage",
		Name:             "gcs_bucket",
		defaultDashboard: true,
	},
	{
		DisplayName:      "Dataflow",
		Name:             "dataflow_job",
		defaultDashboard: true,
	},
	{
		DisplayName:      "Disks",
		Name:             "gce_disk",
		defaultDashboard: true,
	},
	{
		DisplayName:      "External HTTP(S) Load Balancers",
		Name:             "l7_lb_rule",
		defaultDashboard: true,
	},
	{
		DisplayName:      "Firewalls",
		Name:             "compute_firewall",
		defaultDashboard: true,
	},
	{
		DisplayName:      "GKE",
		Name:             "kubernetes",
		defaultDashboard: true,
	},
	{
		DisplayName:      "Google Cloud Load Balancers",
		Name:             "loadbalancing",
		defaultDashboard: true,
	},
	{
		DisplayName:      "Infrastructure Summary",
		Name:             "infrastructure",
		defaultDashboard: true,
	},
	{
		DisplayName:      "Network Security Policies",
		Name:             "network_security_policy",
		defaultDashboard: true,
	},
	{
		DisplayName:      "Pub/Sub",
		Name:             "pubsub_topic",
		defaultDashboard: true,
	},
	{
		DisplayName:      "VM Instances",
		Name:             "gce_instance",
		defaultDashboard: true,
	},
}

func init() {
	flag.Parse()
	if *projectId == "" {
		log.Println("require flag '-p'. project id must be set")
		os.Exit(0)
	}
	homeDir, _ = os.UserHomeDir()
	baseDir = homeDir + "/.cloudmonitoring_dig"
}

func main() {
	if callGetCache() {
		getCache()
	}
	dashboards = getDashboards(*projectId)
	dashboards = append(dashboards, defaultDashboards...)
	i, _ := fuzzyfinder.Find(
		dashboards,
		func(i int) string {
			return dashboards[i].DisplayName
		},
	)
	dashboards[i].Open(*projectId)
}

func getDashboards(selectedProject string) []*Dashboard {
	dashboards := []*Dashboard{}
	filename := getCacheFilePath()
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(f, &dashboards)
	return dashboards
}

func getCache() {
	log.Println("get dashbaords from gcloud")
	if _, err := os.Stat(getCacheDir()); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(getCacheDir(), os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	f, err := os.Create(getCacheFilePath())
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	cmd := exec.Command("gcloud", "monitoring", "dashboards", "list", "--project", *projectId, "--format", "json")
	cmd.Stdout = f
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func getCacheDir() string {
	return baseDir + "/" + *projectId
}

func getCacheFilePath() string {
	cacheDir := getCacheDir()
	return cacheDir + "/cache.json"
}

func callGetCache() bool {
	if *updateCache {
		log.Println("update dashbaords from gcloud")
		return true
	}
	if _, err := os.Stat(getCacheFilePath()); err != nil {
		if os.IsNotExist(err) {
			return true
		}
	}
	return false
}

func (d *Dashboard) String() string {
	return d.DisplayName
}

func (d *Dashboard) Open(projectId string) {
	url := d.buildUrl(projectId)
	log.Printf("%s", url)
	cmd := exec.Command("open", url)
	cmd.Run()
	os.Exit(cmd.ProcessState.ExitCode())
}

// default pattern: https://console.cloud.google.com/monitoring/dashboards/resourceList/spanner_instance?project=$selected_project
// pattern: https://console.cloud.google.com/monitoring/dashboards/builder/$name\?project=$selected_project
func (d *Dashboard) buildUrl(projectId string) string {
	if d.defaultDashboard {
		return fmt.Sprintf("https://console.cloud.google.com/monitoring/dashboards/resourceList/%s?project=%s", d.Name, projectId)
	} else {

		ms := r.FindString(d.Name)
		return fmt.Sprintf("https://console.cloud.google.com/monitoring/dashboards/builder%s?project=%s", ms, projectId)
	}
}
