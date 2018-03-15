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

var _ = Describe("The elasticlicense client", func() {
	var server *ghttp.Server
	var elasticHost string
	var elasticPort int
	const License = "license"
	var licenseFile *os.File
	const endpoint = "/_xpack/license"
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

	Context("install command", func() {
		BeforeEach(func() {
			server = ghttp.NewServer()

			u, err := url.Parse(server.URL())
			Expect(err).ShouldNot(HaveOccurred())

			host, port, err := net.SplitHostPort(u.Host)
			Expect(err).ShouldNot(HaveOccurred())

			elasticHost = host
			elasticPort, err = strconv.Atoi(port)
			Expect(err).ShouldNot(HaveOccurred())

			license := []byte(License)
			licenseFile, err = ioutil.TempFile("", "license")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = licenseFile.Write(license)
			Expect(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			if licenseFile != nil {
				os.Remove(licenseFile.Name())
			}
			server.Close()
		})
		It("should run without basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", endpoint),
					ghttp.VerifyBody([]byte(License)),
				),
			)

			cmd := &installCmd{
				host:        elasticHost,
				port:        elasticPort,
				licenseFile: licenseFile.Name(),
				authFile:    ""}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})

		It("should run with basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", endpoint),
					ghttp.VerifyBody([]byte(License)),
					ghttp.VerifyBasicAuth(Username, Password),
				),
			)

			authFile := createAuthFile()
			defer os.Remove(authFile)

			cmd := &installCmd{
				host:        elasticHost,
				port:        elasticPort,
				licenseFile: licenseFile.Name(),
				authFile:    authFile}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})
	})
	Context("view command", func() {
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
					ghttp.VerifyRequest("GET", endpoint),
					ghttp.RespondWithJSONEncoded(http.StatusOK, License),
				),
			)

			cmd := &viewCmd{
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
					ghttp.VerifyRequest("GET", endpoint),
					ghttp.RespondWithJSONEncoded(http.StatusOK, License),
					ghttp.VerifyBasicAuth(Username, Password),
				),
			)

			authFile := createAuthFile()
			defer os.Remove(authFile)

			cmd := &viewCmd{
				host:     elasticHost,
				port:     elasticPort,
				authFile: authFile}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})
	})
})
