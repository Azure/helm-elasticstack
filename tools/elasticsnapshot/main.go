// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// AzureSnapshotType snapshot type for Azure storage
const AzureSnapshotType = "azure"

// SnapshotSettings snapshot settings
type SnapshotSettings struct {
	Type string `json:"type"`
}

// BasicAuth basic authentication credentials
type BasicAuth struct {
	Username string
	Password string
}

func buildSnapshotURL(host string, port int, repository string, snapshot string) string {
	return fmt.Sprintf("http://%s:%d/_snapshot/%s/%s", host, port, repository, snapshot)
}

func buildHTTPClient() *http.Client {
	return &http.Client{
		Timeout: time.Minute * 1,
	}
}

func loadBasicAuth(authFile string) (*BasicAuth, error) {
	file, err := ioutil.ReadFile(authFile) // #nosec
	if err != nil {
		return nil, fmt.Errorf(`Failed to read the basic authentication
		credentials from the file: %v`, err)
	}

	var auth BasicAuth
	err = json.Unmarshal(file, &auth)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal the basic auth: %v", err)
	}
	return &auth, nil
}

func setBasicAuth(req *http.Request, authFile string) error {
	if authFile != "" {
		basicAuth, err := loadBasicAuth(authFile)
		if err != nil {
			return err
		}
		req.SetBasicAuth(basicAuth.Username, basicAuth.Password)
	}
	return nil
}

type createCmd struct {
	host       string
	port       int
	repository string
	snapshot   string
	verify     bool
	authFile   string
}

func (*createCmd) Name() string { return "create" }
func (*createCmd) Synopsis() string {
	return "create a new snapshot of the entire Elasticsearch cluster in an Azure storage"
}
func (*createCmd) Usage() string {
	return `create [-host] <host name> [-port] <port> [-repository] <repository-name> [-snapshot] <snapshot name> [-auth-file] <path to basic auth file> [-verify] <true/false>
        Create a new snapshot of the entire Elasticsearch cluster in an Azure storage
	`
}

func (c *createCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.host, "host", "localhost", "Host name of the Elasticsearch API")
	f.IntVar(&c.port, "port", 9200, "Port of the Elastisearch API")
	f.StringVar(&c.repository, "repository", "", "Repository name where the snapshot is created")
	f.StringVar(&c.snapshot, "snapshot", "", "Snapshot name")
	f.BoolVar(&c.verify, "verify", false, "Enable the repository verification")
	f.StringVar(&c.authFile, "auth-file", "", "Path to basic auth file")
}

func (c *createCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	snapshotURL := buildSnapshotURL(c.host, c.port, c.repository, c.snapshot)
	if !c.verify {
		snapshotURL = snapshotURL + "?verify=false"
	}

	settings := &SnapshotSettings{
		Type: AzureSnapshotType,
	}
	reqBody, err := json.Marshal(&settings)
	if err != nil {
		fmt.Printf("Failed to build the HTTP request body. Error: %v\n", err)
		return subcommands.ExitFailure
	}

	req, err := http.NewRequest(http.MethodPut, snapshotURL, bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("Failed to build the HTTP request. Error: %v\n", err)
		return subcommands.ExitFailure
	}

	err = setBasicAuth(req, c.authFile)
	if err != nil {
		fmt.Println("Failed to set the basic authentication header")
		return subcommands.ExitFailure
	}

	client := buildHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to create snapshot request. Error: %v", err)
		return subcommands.ExitFailure
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		content, _ := ioutil.ReadAll(resp.Body) // #nosec
		fmt.Printf("Failed to create the snapshot.\n Status Code: %d\n Error Message: %s\n", resp.StatusCode, string(content))
		return subcommands.ExitFailure
	}

	fmt.Printf("Start creating snapshot: %s/%s", c.repository, c.snapshot)
	return subcommands.ExitSuccess
}

type statusCmd struct {
	host       string
	port       int
	repository string
	snapshot   string
	authFile   string
}

func (*statusCmd) Name() string { return "status" }
func (*statusCmd) Synopsis() string {
	return "retrieves the status of an Elasticsearch snapshot"
}
func (*statusCmd) Usage() string {
	return `status [-host] <host name> [-port] <port> [-repository] <repository-name> [-snapshot] <snapshot name> [-auth-file] <path to basic auth file>
        Retrieves the status of an Elasticsearch snapshot
	`
}

func (s *statusCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&s.host, "host", "localhost", "Host name of the Elasticsearch API")
	f.IntVar(&s.port, "port", 9200, "Port of the Elastisearch API")
	f.StringVar(&s.repository, "repository", "", "Repository name where the snapshot is created")
	f.StringVar(&s.snapshot, "snapshot", "", "Snapshot name")
	f.StringVar(&s.authFile, "auth-file", "", "Path to basic auth file")
}

func (s *statusCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	snapshotURL := buildSnapshotURL(s.host, s.port, s.repository, s.snapshot)
	snapshotURL = snapshotURL + "/_status"

	req, err := http.NewRequest(http.MethodGet, snapshotURL, nil)
	if err != nil {
		fmt.Printf("Failed to build the HTTP request. Error: %v\n", err)
		return subcommands.ExitFailure
	}

	err = setBasicAuth(req, s.authFile)
	if err != nil {
		fmt.Println("Failed to set the basic authentication header")
		return subcommands.ExitFailure
	}

	client := buildHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to get the snapshot status. Error: %v", err)
		return subcommands.ExitFailure
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || err != nil {
		fmt.Printf("Failed to read the snapshot status.\n Status Code: %d\n Error Message: %s\n", resp.StatusCode, string(content))
		return subcommands.ExitFailure
	}

	var prettyStatus bytes.Buffer
	err = json.Indent(&prettyStatus, content, "", "    ")
	if err != nil {
		fmt.Printf("Failed to indent the status information. Error: %v\n", err)
		return subcommands.ExitFailure

	}
	fmt.Println("Status:")
	fmt.Print(string(prettyStatus.Bytes()))
	fmt.Println()

	return subcommands.ExitSuccess
}

type restoreCmd struct {
	host       string
	port       int
	repository string
	snapshot   string
	authFile   string
}

func (*restoreCmd) Name() string { return "restore" }
func (*restoreCmd) Synopsis() string {
	return "restore an entire Elasticsearch cluster snapshot from Azure storage"
}
func (*restoreCmd) Usage() string {
	return `status [-host] <host name> [-port] <port> [-repository] <repository-name> [-snapshot] <snapshot name> [-auth-file] <path to basic auth file>
        Restore an entire Elasticsearch cluster snapshot from Azure storage
	`
}

func (r *restoreCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.host, "host", "localhost", "Host name of the Elasticsearch API")
	f.IntVar(&r.port, "port", 9200, "Port of the Elastisearch API")
	f.StringVar(&r.repository, "repository", "", "Repository name where the snapshot is created")
	f.StringVar(&r.snapshot, "snapshot", "", "Snapshot name")
	f.StringVar(&r.authFile, "auth-file", "", "Path to basic auth file")
}

func (r *restoreCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	snapshotURL := buildSnapshotURL(r.host, r.port, r.repository, r.snapshot)
	snapshotURL = snapshotURL + "/_restore"

	req, err := http.NewRequest(http.MethodPost, snapshotURL, nil)
	if err != nil {
		fmt.Printf("Failed to build the HTTP request. Error: %v\n", err)
		return subcommands.ExitFailure
	}

	err = setBasicAuth(req, r.authFile)
	if err != nil {
		fmt.Println("Failed to set the basic authentication header")
		return subcommands.ExitFailure
	}

	client := buildHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to restore snapshot. Error: %v", err)
		return subcommands.ExitFailure
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		content, _ := ioutil.ReadAll(resp.Body) // #nosec
		fmt.Printf("Failed to restore the snapshot.\n Status Code: %d\n Error Message: %s\n", resp.StatusCode, string(content))
		return subcommands.ExitFailure
	}

	fmt.Printf("Start restoring snapshot: %s/%s", r.repository, r.snapshot)
	return subcommands.ExitSuccess
}

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&createCmd{}, "")
	subcommands.Register(&statusCmd{}, "")
	subcommands.Register(&restoreCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
