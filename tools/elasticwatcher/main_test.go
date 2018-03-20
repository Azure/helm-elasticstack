package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/google/subcommands"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("The elasticwatcher client", func() {
	var server *ghttp.Server
	var elasticHost string
	var elasticPort int
	const WatchName = "watch_test"
	const WatchBody = "{\"content\":\"test\"}\n"
	var Watches = fmt.Sprintf("{ \"watches\": [{\"name\": \"%s\", \"body\": %s }]}",
		WatchName, WatchBody)
	var WatchResponse = fmt.Sprintf("{\"%s\": %s}", WatchName, WatchBody)
	var SearchResponse = fmt.Sprintf("{\"hits\": {\"hits\": [ {\"_id\": \"%s\"}]}}", WatchName)
	var watchesFile *os.File
	const watcherEndpoint = "/_xpack/watcher/watch"
	const watchesSearchEndpoint = "/.watches/_search"
	const Username = "test"
	const Password = "test"
	var Auth = fmt.Sprintf("{\"Username\": \"%s\", \"Password\": \"%s\"}", Username, Password)

	createAuthFile := func() string {
		auth := []byte(Auth)
		authFile, err := ioutil.TempFile("", "auth")
		Expect(err).ShouldNot(HaveOccurred())

		_, err = authFile.Write(auth)
		Expect(err).ShouldNot(HaveOccurred())

		return authFile.Name()
	}

	Context("create command", func() {
		BeforeEach(func() {
			server = ghttp.NewServer()

			u, err := url.Parse(server.URL())
			Expect(err).ShouldNot(HaveOccurred())

			host, port, err := net.SplitHostPort(u.Host)
			Expect(err).ShouldNot(HaveOccurred())

			elasticHost = host
			elasticPort, err = strconv.Atoi(port)
			Expect(err).ShouldNot(HaveOccurred())

			watches := []byte(Watches)
			watchesFile, err = ioutil.TempFile("", "watches")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = watchesFile.Write(watches)
			Expect(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			if watchesFile != nil {
				os.Remove(watchesFile.Name())
			}
			server.Close()
		})
		It("should run without basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", watcherEndpoint+"/"+WatchName),
					ghttp.VerifyBody([]byte(WatchBody)),
				),
			)

			cmd := &createCmd{
				host:        elasticHost,
				port:        elasticPort,
				watchesFile: watchesFile.Name(),
				authFile:    ""}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})

		It("should run with basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", watcherEndpoint+"/"+WatchName),
					ghttp.VerifyBody([]byte(WatchBody)),
					ghttp.VerifyBasicAuth(Username, Password),
				),
			)

			authFile := createAuthFile()
			defer os.Remove(authFile)

			cmd := &createCmd{
				host:        elasticHost,
				port:        elasticPort,
				watchesFile: watchesFile.Name(),
				authFile:    authFile}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})
	})

	Context("retrieve command", func() {
		BeforeEach(func() {
			server = ghttp.NewServer()

			u, err := url.Parse(server.URL())
			Expect(err).ShouldNot(HaveOccurred())

			host, port, err := net.SplitHostPort(u.Host)
			Expect(err).ShouldNot(HaveOccurred())

			elasticHost = host
			elasticPort, err = strconv.Atoi(port)
			Expect(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			if watchesFile != nil {
				os.Remove(watchesFile.Name())
			}
			server.Close()
		})
		It("should run without basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", watcherEndpoint+"/"+WatchName),
					ghttp.RespondWith(http.StatusOK, WatchResponse),
				),
			)

			cmd := &retrieveCmd{
				host:     elasticHost,
				port:     elasticPort,
				watches:  WatchName,
				authFile: ""}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})

		It("should run with basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", watcherEndpoint+"/"+WatchName),
					ghttp.RespondWith(http.StatusOK, WatchResponse),
					ghttp.VerifyBasicAuth(Username, Password),
				),
			)

			authFile := createAuthFile()
			defer os.Remove(authFile)

			cmd := &retrieveCmd{
				host:     elasticHost,
				port:     elasticPort,
				watches:  WatchName,
				authFile: authFile}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})
	})

	Context("delete command", func() {
		BeforeEach(func() {
			server = ghttp.NewServer()

			u, err := url.Parse(server.URL())
			Expect(err).ShouldNot(HaveOccurred())

			host, port, err := net.SplitHostPort(u.Host)
			Expect(err).ShouldNot(HaveOccurred())

			elasticHost = host
			elasticPort, err = strconv.Atoi(port)
			Expect(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			if watchesFile != nil {
				os.Remove(watchesFile.Name())
			}
			server.Close()
		})
		It("should run without basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", watcherEndpoint+"/"+WatchName),
					ghttp.RespondWith(http.StatusOK, WatchResponse),
				),
			)

			cmd := &deleteCmd{
				host:     elasticHost,
				port:     elasticPort,
				watches:  WatchName,
				authFile: ""}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})

		It("should run with basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", watcherEndpoint+"/"+WatchName),
					ghttp.RespondWith(http.StatusOK, WatchResponse),
					ghttp.VerifyBasicAuth(Username, Password),
				),
			)

			authFile := createAuthFile()
			defer os.Remove(authFile)

			cmd := &deleteCmd{
				host:     elasticHost,
				port:     elasticPort,
				watches:  WatchName,
				authFile: authFile}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})
	})

	Context("activate command", func() {
		BeforeEach(func() {
			server = ghttp.NewServer()

			u, err := url.Parse(server.URL())
			Expect(err).ShouldNot(HaveOccurred())

			host, port, err := net.SplitHostPort(u.Host)
			Expect(err).ShouldNot(HaveOccurred())

			elasticHost = host
			elasticPort, err = strconv.Atoi(port)
			Expect(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			if watchesFile != nil {
				os.Remove(watchesFile.Name())
			}
			server.Close()
		})
		It("should run without basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", watcherEndpoint+"/"+WatchName+"/"+"_activate"),
					ghttp.RespondWith(http.StatusOK, WatchResponse),
				),
			)

			cmd := &activateCmd{
				host:     elasticHost,
				port:     elasticPort,
				watches:  WatchName,
				authFile: ""}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})

		It("should run with basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", watcherEndpoint+"/"+WatchName+"/"+"_activate"),
					ghttp.RespondWith(http.StatusOK, WatchResponse),
					ghttp.VerifyBasicAuth(Username, Password),
				),
			)

			authFile := createAuthFile()
			defer os.Remove(authFile)

			cmd := &activateCmd{
				host:     elasticHost,
				port:     elasticPort,
				watches:  WatchName,
				authFile: authFile}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})
	})

	Context("deactivate command", func() {
		BeforeEach(func() {
			server = ghttp.NewServer()

			u, err := url.Parse(server.URL())
			Expect(err).ShouldNot(HaveOccurred())

			host, port, err := net.SplitHostPort(u.Host)
			Expect(err).ShouldNot(HaveOccurred())

			elasticHost = host
			elasticPort, err = strconv.Atoi(port)
			Expect(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			if watchesFile != nil {
				os.Remove(watchesFile.Name())
			}
			server.Close()
		})
		It("should run without basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", watcherEndpoint+"/"+WatchName+"/"+"_deactivate"),
					ghttp.RespondWith(http.StatusOK, WatchResponse),
				),
			)

			cmd := &deactivateCmd{
				host:     elasticHost,
				port:     elasticPort,
				watches:  WatchName,
				authFile: ""}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})

		It("should run with basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", watcherEndpoint+"/"+WatchName+"/"+"_deactivate"),
					ghttp.RespondWith(http.StatusOK, WatchResponse),
					ghttp.VerifyBasicAuth(Username, Password),
				),
			)

			authFile := createAuthFile()
			defer os.Remove(authFile)

			cmd := &deactivateCmd{
				host:     elasticHost,
				port:     elasticPort,
				watches:  WatchName,
				authFile: authFile}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})
	})
	Context("list command", func() {
		BeforeEach(func() {
			server = ghttp.NewServer()

			u, err := url.Parse(server.URL())
			Expect(err).ShouldNot(HaveOccurred())

			host, port, err := net.SplitHostPort(u.Host)
			Expect(err).ShouldNot(HaveOccurred())

			elasticHost = host
			elasticPort, err = strconv.Atoi(port)
			Expect(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			if watchesFile != nil {
				os.Remove(watchesFile.Name())
			}
			server.Close()
		})
		It("should run without basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", watchesSearchEndpoint),
					ghttp.RespondWith(http.StatusOK, SearchResponse),
				),
			)

			cmd := &listCmd{
				host:     elasticHost,
				port:     elasticPort,
				authFile: ""}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})

		It("should run with basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", watchesSearchEndpoint),
					ghttp.RespondWith(http.StatusOK, SearchResponse),
					ghttp.VerifyBasicAuth(Username, Password),
				),
			)

			authFile := createAuthFile()
			defer os.Remove(authFile)

			cmd := &listCmd{
				host:     elasticHost,
				port:     elasticPort,
				authFile: authFile}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})
	})
})
