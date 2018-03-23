package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/subcommands"
	"github.com/oliveagle/jsonpath"
)

// BasicAuth basic authentication credentials
type BasicAuth struct {
	Username string
	Password string
}

// WatcherConfig holds all watch configurations
type WatcherConfig struct {
	Watches []Watch
}

// Watch watch configuration
type Watch struct {
	Name string
	Body interface{}
}

func buildWatcherURL(host string, port int, watchID string) string {
	return fmt.Sprintf("http://%s:%d/_xpack/watcher/watch/%s", host, port, watchID)
}

func buildWatcherSearchURL(host string, port int) string {
	return fmt.Sprintf("http://%s:%d/.watches/_search", host, port)
}

func buildHTTPClient() *http.Client {
	return &http.Client{
		Timeout: time.Minute * 1,
	}
}

func loadBasicAuth(authFile string) (*BasicAuth, error) {
	file, err := ioutil.ReadFile(authFile)
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

func loadWatches(watchesFile string) (*WatcherConfig, error) {
	file, err := ioutil.ReadFile(watchesFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read the watches from file: %v", err)
	}
	var w WatcherConfig
	err = json.Unmarshal(file, &w)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal the watches: %v", err)
	}
	return &w, nil
}

func parseWatchNames(watches string) []string {
	watchNames := strings.Split(watches, ",")
	for i, name := range watchNames {
		watchNames[i] = strings.TrimSpace(name)
	}
	return watchNames
}

func switchWatch(host string, port int, authFile string, watch string, action string) subcommands.ExitStatus {
	fragment := fmt.Sprintf("%s/_%s", watch, action)
	watcherURL := buildWatcherURL(host, port, fragment)

	client := buildHTTPClient()
	req, err := http.NewRequest(http.MethodPut, watcherURL, nil)
	if err != nil {
		fmt.Printf("Failed to build the HTTP request for watch '%s'. Error: %v\n", watch, err)
		return subcommands.ExitFailure
	}

	err = setBasicAuth(req, authFile)
	if err != nil {
		fmt.Printf("Failed to set the Basic Auth Header for watch '%s'\n", watch)
		return subcommands.ExitFailure
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to %s the watch '%s'. Error: %v\n", action, watch, err)
		return subcommands.ExitFailure
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || err != nil {
		fmt.Printf("Failed to %s the watch '%s'. Status Code: %d. Error: %v\n", action, watch, resp.StatusCode, string(content))
		return subcommands.ExitFailure
	}

	var prettyContent bytes.Buffer
	err = json.Indent(&prettyContent, content, "", "    ")
	if err != nil {
		fmt.Printf("Failed to indent the content of %s response. Error: %v", action, err)
		return subcommands.ExitFailure
	}

	fmt.Printf("%s watch '%s':\n", action, watch)
	fmt.Print(string(prettyContent.Bytes()))
	fmt.Println()
	return subcommands.ExitSuccess
}

type listCmd struct {
	host     string
	port     int
	authFile string
}

func (*listCmd) Name() string { return "list" }
func (*listCmd) Synopsis() string {
	return "List all watches installed in Elasticsearch Watcher"
}

func (*listCmd) Usage() string {
	return `list [-host] <host name> [-port] <port> [-auth-file] <path to basic auth file>
        List the watches which are installed in Elasticsearch Watcher
	`
}

func (l *listCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&l.host, "host", "localhost", "Host name of the Elasticsearch API")
	f.IntVar(&l.port, "port", 9200, "Port of the Elastisearch API")
	f.StringVar(&l.authFile, "auth-file", "", "Path to basic auth file")
}

func (l *listCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	client := buildHTTPClient()

	watcherSearchURL := buildWatcherSearchURL(l.host, l.port)

	req, err := http.NewRequest(http.MethodGet, watcherSearchURL, nil)
	if err != nil {
		fmt.Printf("Failed to build the HTTP request. Error: %v\n", err)
		return subcommands.ExitFailure
	}

	err = setBasicAuth(req, l.authFile)
	if err != nil {
		fmt.Println("Failed to set the Basic Auth Header")
		return subcommands.ExitFailure
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to list the watches. Error: %v\n", err)
		return subcommands.ExitFailure
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || err != nil {
		fmt.Printf("Failed to list the watches. Status Code: %d. Error: %v\n", resp.StatusCode, string(content))
		return subcommands.ExitFailure
	}

	var watches interface{}
	err = json.Unmarshal(content, &watches)
	if err != nil {
		fmt.Printf("Failed to parse the response. Error: %v\n", err)
		return subcommands.ExitFailure
	}

	res, err := jsonpath.JsonPathLookup(watches, "$.hits.hits[*]._id")
	if err != nil {
		fmt.Printf("Failed to find the watch IDs in the response. Error: %v\n", err)
		return subcommands.ExitFailure
	}

	fmt.Println("Installed Watches:")
	if watchIDs, ok := res.([]interface{}); ok == true {
		for _, watchID := range watchIDs {
			fmt.Printf("  %s\n", watchID)
		}
	}

	return subcommands.ExitSuccess
}

type deactivateCmd struct {
	host     string
	port     int
	authFile string
	watches  string
}

func (*deactivateCmd) Name() string { return "deactivate" }
func (*deactivateCmd) Synopsis() string {
	return "Deactivate a list of watches from Elasicsearch Watcher"
}

func (*deactivateCmd) Usage() string {
	return `deactivate [-host] <host name> [-port] <port> [-watches] <comma separated list of watchers> [-auth-file] <path to basic auth file>
        Deactivate a list of watches from Elasticsearch Watcher
	`
}

func (da *deactivateCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&da.host, "host", "localhost", "Host name of the Elasticsearch API")
	f.IntVar(&da.port, "port", 9200, "Port of the Elastisearch API")
	f.StringVar(&da.watches, "watches", "", "Comma separated list with watches names")
	f.StringVar(&da.authFile, "auth-file", "", "Path to basic auth file")
}

func (da *deactivateCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	watchNames := parseWatchNames(da.watches)

	for _, watch := range watchNames {
		rc := switchWatch(da.host, da.port, da.authFile, watch, "deactivate")
		if rc != subcommands.ExitSuccess {
			return rc
		}
	}

	return subcommands.ExitSuccess
}

type activateCmd struct {
	host     string
	port     int
	authFile string
	watches  string
}

func (*activateCmd) Name() string { return "activate" }
func (*activateCmd) Synopsis() string {
	return "Activate a list of watches from Elasicsearch Watcher"
}

func (*activateCmd) Usage() string {
	return `activate [-host] <host name> [-port] <port> [-watches] <comma separated list of watchers> [-auth-file] <path to basic auth file>
        Activate a list of watches from Elasticsearch Watcher
	`
}

func (a *activateCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&a.host, "host", "localhost", "Host name of the Elasticsearch API")
	f.IntVar(&a.port, "port", 9200, "Port of the Elastisearch API")
	f.StringVar(&a.watches, "watches", "", "Comma separated list with watches names")
	f.StringVar(&a.authFile, "auth-file", "", "Path to basic auth file")
}

func (a *activateCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	watchNames := parseWatchNames(a.watches)

	for _, watch := range watchNames {
		rc := switchWatch(a.host, a.port, a.authFile, watch, "activate")
		if rc != subcommands.ExitSuccess {
			return rc
		}
	}

	return subcommands.ExitSuccess
}

type deleteCmd struct {
	host     string
	port     int
	authFile string
	watches  string
}

func (*deleteCmd) Name() string { return "delete" }
func (*deleteCmd) Synopsis() string {
	return "Delete a list of watches from Elasicsearch Watcher"
}

func (*deleteCmd) Usage() string {
	return `delete [-host] <host name> [-port] <port> [-watches] <comma separated list of watchers> [-auth-file] <path to basic auth file>
        Delete a list of watches from Elasticsearch Watcher
	`
}

func (d *deleteCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&d.host, "host", "localhost", "Host name of the Elasticsearch API")
	f.IntVar(&d.port, "port", 9200, "Port of the Elastisearch API")
	f.StringVar(&d.watches, "watches", "", "Comma separated list with watches names")
	f.StringVar(&d.authFile, "auth-file", "", "Path to basic auth file")
}

func (d *deleteCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	watchNames := parseWatchNames(d.watches)
	client := buildHTTPClient()

	for _, watch := range watchNames {
		watcherURL := buildWatcherURL(d.host, d.port, watch)

		req, err := http.NewRequest(http.MethodDelete, watcherURL, nil)
		if err != nil {
			fmt.Printf("Failed to build the HTTP request for watch '%s'. Error: %v\n", watch, err)
			return subcommands.ExitFailure
		}

		err = setBasicAuth(req, d.authFile)
		if err != nil {
			fmt.Printf("Failed to set the Basic Auth Header for watch '%s'\n", watch)
			return subcommands.ExitFailure
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Failed to delete the watch '%s'. Error: %v\n", watch, err)
			return subcommands.ExitFailure
		}
		defer resp.Body.Close()

		content, err := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK || err != nil {
			fmt.Printf("Failed to delete the watch '%s'. Status Code: %d. Error: %v\n", watch, resp.StatusCode, string(content))
			return subcommands.ExitFailure
		}

		var prettyContent bytes.Buffer
		err = json.Indent(&prettyContent, content, "", "    ")
		if err != nil {
			fmt.Printf("Failed to indent the content of watch '%s'. Error: %v", watch, err)
			return subcommands.ExitFailure
		}

		fmt.Printf("Deleted watch '%s':\n", watch)
		fmt.Print(string(prettyContent.Bytes()))
		fmt.Println()
	}

	return subcommands.ExitSuccess
}

type retrieveCmd struct {
	host     string
	port     int
	authFile string
	watches  string
}

func (*retrieveCmd) Name() string { return "retrieve" }
func (*retrieveCmd) Synopsis() string {
	return "Retrieve a list of watches from Elasicsearch Watcher by their name"
}

func (*retrieveCmd) Usage() string {
	return `retrieve [-host] <host name> [-port] <port> [-watches] <comma separated list of watchers> [-auth-file] <path to basic auth file>
        Retrieve a list of watches from Elasticsearch Watcher by their name
	`
}

func (r *retrieveCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.host, "host", "localhost", "Host name of the Elasticsearch API")
	f.IntVar(&r.port, "port", 9200, "Port of the Elastisearch API")
	f.StringVar(&r.watches, "watches", "", "Comma separated list with watches names")
	f.StringVar(&r.authFile, "auth-file", "", "Path to basic auth file")
}

func (r *retrieveCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	watchNames := parseWatchNames(r.watches)
	client := buildHTTPClient()

	for _, watch := range watchNames {
		watcherURL := buildWatcherURL(r.host, r.port, watch)

		req, err := http.NewRequest(http.MethodGet, watcherURL, nil)
		if err != nil {
			fmt.Printf("Failed to build the HTTP request for watch '%s'. Error: %v\n", watch, err)
			return subcommands.ExitFailure
		}

		err = setBasicAuth(req, r.authFile)
		if err != nil {
			fmt.Printf("Failed to set the Basic Auth Header for watch '%s'\n", watch)
			return subcommands.ExitFailure
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Failed to retrieve the watch '%s'. Error: %v\n", watch, err)
			return subcommands.ExitFailure
		}
		defer resp.Body.Close()

		content, err := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK || err != nil {
			fmt.Printf("Failed to retrieve the watch '%s'. Status Code: %d. Error: %v\n", watch, resp.StatusCode, string(content))
			return subcommands.ExitFailure
		}

		var prettyContent bytes.Buffer
		err = json.Indent(&prettyContent, content, "", "    ")
		if err != nil {
			fmt.Printf("Failed to indent the content of watch '%s'. Error: %v", watch, err)
			return subcommands.ExitFailure
		}

		fmt.Printf("Watch: %s\n", watch)
		fmt.Print(string(prettyContent.Bytes()))
		fmt.Println()
	}

	return subcommands.ExitSuccess
}

type createCmd struct {
	host        string
	port        int
	watchesFile string
	authFile    string
}

func (*createCmd) Name() string { return "create" }
func (*createCmd) Synopsis() string {
	return "Register a list of watches in Elasicsearch Watcher or update them"
}
func (*createCmd) Usage() string {
	return `create [-host] <host name> [-port] <port> [-watches-file] <path to watches file> [-auth-file] <path to basic auth file>
        Register a list of watches in Elasticsearch Watcher or update them
	`
}

func (c *createCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.host, "host", "localhost", "Host name of the Elasticsearch API")
	f.IntVar(&c.port, "port", 9200, "Port of the Elastisearch API")
	f.StringVar(&c.watchesFile, "watches-file", "", "Path to watches file")
	f.StringVar(&c.authFile, "auth-file", "", "Path to basic auth file")
}

func (c *createCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if _, err := os.Stat(c.watchesFile); os.IsNotExist(err) {
		fmt.Printf("Watches file '%s' not found\n", c.watchesFile)
		return subcommands.ExitFailure
	}
	cfg, err := loadWatches(c.watchesFile)
	if err != nil {
		fmt.Print(err)
		return subcommands.ExitFailure
	}

	for _, watch := range cfg.Watches {
		reader, writer := io.Pipe()
		wg := sync.WaitGroup{}
		wg.Add(2)
		errc := make(chan error, 1)
		go func() {
			defer wg.Done()
			defer writer.Close()
			enc := json.NewEncoder(writer)
			enc.Encode(watch.Body) // #nosec
		}()

		go func() {
			defer wg.Done()
			defer reader.Close()
			watcherURL := buildWatcherURL(c.host, c.port, watch.Name)
			req, err := http.NewRequest(http.MethodPut, watcherURL, reader)
			if err != nil {
				errc <- fmt.Errorf("Failed to build the request to create the watch: %v", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			err = setBasicAuth(req, c.authFile)
			if err != nil {
				errc <- fmt.Errorf("Failed to set Basic Auth header: %v", err)
				return
			}

			client := buildHTTPClient()
			resp, err := client.Do(req)
			if err != nil {
				errc <- fmt.Errorf("Failed to execute the watch create/update request: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode >= http.StatusBadRequest {
				content, _ := ioutil.ReadAll(resp.Body) // #nosec
				errc <- fmt.Errorf(
					"Failed to create/update the watch:\n  Status Code: %d.\n  Error Message: %s",
					resp.StatusCode, string(content))
				return
			}
			errc <- nil
		}()

		wg.Wait()

		if err := <-errc; err != nil {
			fmt.Println(err)
			return subcommands.ExitFailure
		}
		fmt.Printf("Successfully created/updated the watch '%s'.\n", watch.Name)
	}
	return subcommands.ExitSuccess
}

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&createCmd{}, "")
	subcommands.Register(&retrieveCmd{}, "")
	subcommands.Register(&deleteCmd{}, "")
	subcommands.Register(&activateCmd{}, "")
	subcommands.Register(&deactivateCmd{}, "")
	subcommands.Register(&listCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
