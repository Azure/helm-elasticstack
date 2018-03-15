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
)

// BasicAuth credentials for HTTP basic authentication
type BasicAuth struct {
	Username string
	Password string
}

// TemplatesConfig templates configuration
type TemplatesConfig struct {
	Templates []Template
}

// Template define a template
type Template struct {
	Name string
	Body interface{}
}

func buildTemplateURL(host string, port int, templateID string) string {
	return fmt.Sprintf("http://%s:%d/_template/%s", host, port, templateID)
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
		credetials from the file: %v`, err)
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

func loadTemplates(templatesFile string) (*TemplatesConfig, error) {
	file, err := ioutil.ReadFile(templatesFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read the templates from file: %v", err)
	}
	var t TemplatesConfig
	err = json.Unmarshal(file, &t)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal the templates: %v", err)
	}
	return &t, nil
}

func parseTemplateNames(templates string) []string {
	templateNames := strings.Split(templates, ",")
	for i, name := range templateNames {
		templateNames[i] = strings.TrimSpace(name)
	}
	return templateNames
}

type retrieveCmd struct {
	host      string
	port      int
	authFile  string
	templates string
}

func (*retrieveCmd) Name() string { return "retrieve" }
func (*retrieveCmd) Synopsis() string {
	return "Retrieve the content of Elasicsearch Index Templates"
}

func (*retrieveCmd) Usage() string {
	return `retrieve [-host] <host name> [-port] <port> [-templates] <comma separated list of templates> [-auth-file] <path to basic auth file>
        Retrieve the content of Elasticsearch Index Templates

	`
}

func (r *retrieveCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.host, "host", "localhost", "Host name of the Elasticsearch API")
	f.IntVar(&r.port, "port", 9200, "Port of the Elastisearch API")
	f.StringVar(&r.templates, "templates", "", "Comma separated list with template names")
	f.StringVar(&r.authFile, "auth-file", "", "Path to basic auth file")
}

func (r *retrieveCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	templateNames := parseTemplateNames(r.templates)
	client := buildHTTPClient()

	for _, template := range templateNames {
		templateURL := buildTemplateURL(r.host, r.port, template)

		req, err := http.NewRequest(http.MethodGet, templateURL, nil)
		if err != nil {
			fmt.Printf("Failed to build the HTTP request for template '%s'. Error: %v\n", template, err)
			return subcommands.ExitFailure
		}

		err = setBasicAuth(req, r.authFile)
		if err != nil {
			fmt.Printf("Failed to set the Basic Auth Header for template '%s'\n", template)
			return subcommands.ExitFailure
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Failed to retrieve the template '%s'. Error: %v\n", template, err)
			return subcommands.ExitFailure
		}
		defer resp.Body.Close()

		content, err := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK || err == nil {
			fmt.Printf("Failed to retrieve the template '%s'. Status Code: %d. Error: %v\n",
				template, resp.StatusCode, string(content))
			return subcommands.ExitFailure
		}

		var prettyContent bytes.Buffer
		err = json.Indent(&prettyContent, content, "", "    ")
		if err != nil {
			fmt.Printf("Failed to indent the content of the template '%s'. Error: %v", template, err)
			return subcommands.ExitFailure
		}

		fmt.Printf("Template: %s\n", template)
		fmt.Print(string(prettyContent.Bytes()))
		fmt.Println()
	}

	return subcommands.ExitSuccess
}

type deleteCmd struct {
	host      string
	port      int
	authFile  string
	templates string
}

func (*deleteCmd) Name() string { return "delete" }
func (*deleteCmd) Synopsis() string {
	return "Delete the templates from Elasicsearch"
}

func (*deleteCmd) Usage() string {
	return `delete [-host] <host name> [-port] <port> [-templates] <comma separated list of templates> [-auth-file] <path to basic auth file>
        Delete the templates from Elasticsearch
	`
}

func (d *deleteCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	templateNames := parseTemplateNames(d.templates)
	client := buildHTTPClient()

	for _, template := range templateNames {
		templateURL := buildTemplateURL(d.host, d.port, template)

		req, err := http.NewRequest(http.MethodDelete, templateURL, nil)
		if err != nil {
			fmt.Printf("Failed to build the HTTP request for template '%s'. Error: %v\n", template, err)
			return subcommands.ExitFailure
		}

		err = setBasicAuth(req, d.authFile)
		if err != nil {
			fmt.Printf("Failed to set the Basic Auth Header for template '%s'\n", template)
			return subcommands.ExitFailure
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Failed to delete the template '%s'. Error: %v\n", template, err)
			return subcommands.ExitFailure
		}
		defer resp.Body.Close()

		content, err := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK || err != nil {
			fmt.Printf("Failed to delete the template '%s'. Status Code: %d. Error: %v\n", template, resp.StatusCode, string(content))
			return subcommands.ExitFailure
		}

		var prettyContent bytes.Buffer
		err = json.Indent(&prettyContent, content, "", "    ")
		if err != nil {
			fmt.Printf("Failed to indent the content of template '%s' delete response. Error: %v", template, err)
			return subcommands.ExitFailure
		}

		fmt.Printf("Deleted template '%s':\n", template)
		fmt.Print(string(prettyContent.Bytes()))
		fmt.Println()
	}

	return subcommands.ExitSuccess
}

func (d *deleteCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&d.host, "host", "localhost", "Host name of the Elasticsearch API")
	f.IntVar(&d.port, "port", 9200, "Port of the Elastisearch API")
	f.StringVar(&d.templates, "templates", "", "Comma separated list of template names")
	f.StringVar(&d.authFile, "auth-file", "", "Path to basic auth file")
}

type listCmd struct {
	host     string
	port     int
	authFile string
}

func (*listCmd) Name() string { return "list" }
func (*listCmd) Synopsis() string {
	return "List all Elasticsearch Index Templates"
}
func (*listCmd) Usage() string {
	return `list [-host] <host name> [-port] <port> [-auth-file] <path to basic auth file>
        List the Elasticsearch Index Templates
	`
}

func (l *listCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&l.host, "host", "localhost", "Host name of the Elasticsearch API")
	f.IntVar(&l.port, "port", 9200, "Port of the Elastisearch API")
	f.StringVar(&l.authFile, "auth-file", "", "Path to basic auth file")
}

func (l *listCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	client := buildHTTPClient()
	templateURL := buildTemplateURL(l.host, l.port, "*")

	req, err := http.NewRequest(http.MethodGet, templateURL, nil)
	if err != nil {
		fmt.Printf("Failed to build the HTTP request. Error: %v\n", err)
		return subcommands.ExitFailure
	}

	err = setBasicAuth(req, l.authFile)
	if err != nil {
		fmt.Printf("Failed to set the Basic Auth Header\n")
		return subcommands.ExitFailure
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to retrieve the templates. Error: %v\n", err)
		return subcommands.ExitFailure
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || err != nil {
		fmt.Printf("Failed to retrieve the templates. Status Code: %d. Error: %v\n", resp.StatusCode, string(content))
		return subcommands.ExitFailure
	}

	var templates map[string]interface{}
	err = json.Unmarshal(content, &templates)
	if err != nil {
		fmt.Printf("Failed to unmarshal the templates. Error: %v\n", err)
		return subcommands.ExitFailure
	}

	fmt.Println("Templates:")
	for key := range templates {
		fmt.Printf("%s\n", key)
	}
	return subcommands.ExitSuccess
}

type createCmd struct {
	host          string
	port          int
	templatesFile string
	authFile      string
}

func (*createCmd) Name() string { return "create" }
func (*createCmd) Synopsis() string {
	return "Create an Elasticsearch Index Template or update an existing one"
}
func (*createCmd) Usage() string {
	return `create [-host] <host name> [-port] <port> [-templates-file] <path to template file> [-auth-file] <path to basic auth file>
        Create/Update an Elasticsearch Index Template
	`
}

func (c *createCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.host, "host", "localhost", "Host name of the Elasticsearch API")
	f.IntVar(&c.port, "port", 9200, "Port of the Elastisearch API")
	f.StringVar(&c.templatesFile, "templates-file", "", "Path to templates file")
	f.StringVar(&c.authFile, "auth-file", "", "Path to basic auth file")
}

func (c *createCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if _, err := os.Stat(c.templatesFile); os.IsNotExist(err) {
		fmt.Printf("Templates file '%s' not found\n", c.templatesFile)
		return subcommands.ExitFailure
	}

	cfg, err := loadTemplates(c.templatesFile)
	if err != nil {
		fmt.Print(err)
		return subcommands.ExitFailure
	}

	for _, template := range cfg.Templates {
		reader, writer := io.Pipe()
		wg := sync.WaitGroup{}
		wg.Add(2)
		errc := make(chan error, 1)
		go func() {
			defer wg.Done()
			defer writer.Close()
			enc := json.NewEncoder(writer)
			enc.Encode(template.Body) // #nosec
		}()

		go func() {
			defer wg.Done()
			defer reader.Close()
			templateURL := buildTemplateURL(c.host, c.port, template.Name)
			req, err := http.NewRequest(http.MethodPut, templateURL, reader)
			if err != nil {
				errc <- fmt.Errorf("Failed to build the request to create the template: %v", err)
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
				errc <- fmt.Errorf("Failed to execute the template create/update request: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode >= http.StatusBadRequest {
				content, _ := ioutil.ReadAll(resp.Body) // #nosec
				errc <- fmt.Errorf(
					"Failed to create/update the template:\n  Status Code: %d.\n  Error Message: %s",
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
		fmt.Printf("Successfully created/updated the template '%s'.\n", template.Name)
	}
	return subcommands.ExitSuccess
}

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&createCmd{}, "")
	subcommands.Register(&listCmd{}, "")
	subcommands.Register(&deleteCmd{}, "")
	subcommands.Register(&retrieveCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
