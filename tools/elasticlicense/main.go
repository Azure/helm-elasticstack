package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/google/subcommands"
)

func buildLicenseURL(host string, port int) string {
	return fmt.Sprintf("http://%s:%d/_xpack/license", host, port)
}

func buildHTTPClient() *http.Client {
	return &http.Client{
		Timeout: time.Minute * 1,
	}
}

func loadBasicAuth(authFile string) (*basicAuth, error) {
	file, err := ioutil.ReadFile(authFile)
	if err != nil {
		return nil, fmt.Errorf(`Failed to read the basic authentication
		credentials from the file: %v`, err)
	}

	var auth basicAuth
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

type basicAuth struct {
	Username string
	Password string
}

type viewCmd struct {
	host     string
	port     int
	authFile string
}

func (*viewCmd) Name() string     { return "view" }
func (*viewCmd) Synopsis() string { return "Display the installed license in Elasticsearch" }
func (*viewCmd) Usage() string {
	return `view [-host] <host name> [-port] <port> [-auth-file] <basic auth file path>:
	Display the installed license to stdout
	`
}

func (v *viewCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&v.host, "host", "localhost", "Host name of the Elasticsearch API")
	f.IntVar(&v.port, "port", 9200, "Port of the Elastisearch API")
	f.StringVar(&v.authFile, "auth-file", "", "File with basic authentication credentials")
}

func (v *viewCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	client := buildHTTPClient()
	licenseURL := buildLicenseURL(v.host, v.port)

	req, err := http.NewRequest(http.MethodGet, licenseURL, nil)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	err = setBasicAuth(req, v.authFile)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to get the license information with status code [%d]",
			resp.StatusCode)
		return subcommands.ExitFailure
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read the license from the HTTP response body")
		return subcommands.ExitFailure
	}

	var prettyContent bytes.Buffer
	err = json.Indent(&prettyContent, content, "", "    ")
	if err != nil {
		fmt.Printf("Failed to indent the license content: %s\n", string(content))
		return subcommands.ExitFailure
	}

	fmt.Print(string(prettyContent.Bytes()))

	return subcommands.ExitSuccess
}

type installCmd struct {
	host        string
	port        int
	licenseFile string
	authFile    string
}

func (*installCmd) Name() string     { return "install" }
func (*installCmd) Synopsis() string { return "Install a new Elasticsearch license" }
func (*installCmd) Usage() string {
	return `install [-host] <host name> [-port] <port> [-license-file] <path to license file> [-auth-file] <path to basic auth file>
	Install a new license into Elasticsearch cluster
	`
}

func (i *installCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&i.host, "host", "localhost", "Host name of the Elasticsearch API")
	f.IntVar(&i.port, "port", 9200, "Port of the Elastisearch API")
	f.StringVar(&i.licenseFile, "license-file", "", "Path to license file")
	f.StringVar(&i.authFile, "auth-file", "", "Path to basic auth file")
}

func (i *installCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if _, err := os.Stat(i.licenseFile); os.IsNotExist(err) {
		fmt.Printf("License file '%s' not found\n", i.licenseFile)
		return subcommands.ExitFailure
	}

	file, err := os.Open(i.licenseFile)
	if err != nil {
		fmt.Printf("Failed to open the license file '%s'\n", i.licenseFile)
		return subcommands.ExitFailure
	}

	licenseURL := buildLicenseURL(i.host, i.port)
	req, err := http.NewRequest(http.MethodPut, licenseURL, file)
	if err != nil {
		fmt.Printf("Failed to build the request to install the license: %v\n", err)
		return subcommands.ExitFailure
	}
	req.Header.Set("Content-Type", "application/json")
	err = setBasicAuth(req, i.authFile)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	client := buildHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to install the license with status code [%d]\n", resp.StatusCode)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&viewCmd{}, "")
	subcommands.Register(&installCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
