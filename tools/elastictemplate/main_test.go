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

var _ = Describe("The elastictemplate client", func() {
	var server *ghttp.Server
	var elasticHost string
	var elasticPort int
	const TemplateName = "template_test"
	const TemplateBody = "{\"content\":\"test\"}\n"
	var Templates = fmt.Sprintf("{ \"templates\": [{\"name\": \"%s\", \"body\": %s }]}",
		TemplateName, TemplateBody)
	var TemplateResponse = fmt.Sprintf("{\"%s\": %s}", TemplateName, TemplateBody)
	var templatesFile *os.File
	const endpoint = "/_template"
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

			templates := []byte(Templates)
			templatesFile, err = ioutil.TempFile("", "templates")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = templatesFile.Write(templates)
			Expect(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			if templatesFile != nil {
				os.Remove(templatesFile.Name())
			}
			server.Close()
		})
		It("should run without basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", endpoint+"/"+TemplateName),
					ghttp.VerifyBody([]byte(TemplateBody)),
				),
			)

			cmd := &createCmd{
				host:          elasticHost,
				port:          elasticPort,
				templatesFile: templatesFile.Name(),
				authFile:      ""}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})

		It("should run with basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", endpoint+"/"+TemplateName),
					ghttp.VerifyBody([]byte(TemplateBody)),
					ghttp.VerifyBasicAuth(Username, Password),
				),
			)

			authFile := createAuthFile()
			defer os.Remove(authFile)

			cmd := &createCmd{
				host:          elasticHost,
				port:          elasticPort,
				templatesFile: templatesFile.Name(),
				authFile:      authFile}

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
			server.Close()
		})
		It("should run without basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", endpoint+"/"+"*"),
					ghttp.RespondWith(http.StatusOK, TemplateResponse),
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
					ghttp.VerifyRequest("GET", endpoint+"/"+"*"),
					ghttp.RespondWith(http.StatusOK, TemplateResponse),
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
			server.Close()
		})
		It("should run without basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", endpoint+"/"+TemplateName),
					ghttp.RespondWith(http.StatusOK, TemplateResponse),
				),
			)

			cmd := &deleteCmd{
				host:      elasticHost,
				port:      elasticPort,
				templates: TemplateName,
				authFile:  ""}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})

		It("should run with basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", endpoint+"/"+TemplateName),
					ghttp.RespondWith(http.StatusOK, TemplateResponse),
					ghttp.VerifyBasicAuth(Username, Password),
				),
			)

			authFile := createAuthFile()
			defer os.Remove(authFile)

			cmd := &deleteCmd{
				host:      elasticHost,
				port:      elasticPort,
				templates: TemplateName,
				authFile:  authFile}

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
			server.Close()
		})
		It("should run without basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", endpoint+"/"+TemplateName),
					ghttp.RespondWith(http.StatusOK, TemplateResponse),
				),
			)

			cmd := &retrieveCmd{
				host:      elasticHost,
				port:      elasticPort,
				templates: TemplateName,
				authFile:  ""}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})

		It("should run with basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", endpoint+"/"+TemplateName),
					ghttp.RespondWith(http.StatusOK, TemplateResponse),
					ghttp.VerifyBasicAuth(Username, Password),
				),
			)

			authFile := createAuthFile()
			defer os.Remove(authFile)

			cmd := &retrieveCmd{
				host:      elasticHost,
				port:      elasticPort,
				templates: TemplateName,
				authFile:  authFile}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})
	})

})
