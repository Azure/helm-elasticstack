// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"strconv"

	"github.com/google/subcommands"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"net/http"
)

var _ = Describe("The elasticsnapshot client", func() {
	var server *ghttp.Server
	var elasticHost string
	var elasticPort int
	const snapshotName = "test"
	const snapshotRepository = "repository"
	const snapshotEndpoint = "/_snapshot"
	const snapshotQuery = "verify=false"
	const snapshotBody = "{\"type\":\"azure\"}"
	const statusResponse = "{\"state\": \"IN_PROGRESS\"}"
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
		})

		AfterEach(func() {
			server.Close()
		})
		It("should run without basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", snapshotEndpoint+"/"+snapshotRepository+"/"+snapshotName, snapshotQuery),
					ghttp.VerifyBody([]byte(snapshotBody)),
					ghttp.RespondWith(http.StatusCreated, nil),
				),
			)

			cmd := &createCmd{
				host:       elasticHost,
				port:       elasticPort,
				snapshot:   snapshotName,
				repository: snapshotRepository,
				authFile:   ""}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})

		It("should run with basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", snapshotEndpoint+"/"+snapshotRepository+"/"+snapshotName, snapshotQuery),
					ghttp.VerifyBody([]byte(snapshotBody)),
					ghttp.RespondWith(http.StatusCreated, nil),
					ghttp.VerifyBasicAuth(Username, Password),
				),
			)

			authFile := createAuthFile()
			defer os.Remove(authFile)

			cmd := &createCmd{
				host:       elasticHost,
				port:       elasticPort,
				snapshot:   snapshotName,
				repository: snapshotRepository,
				authFile:   authFile}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})
	})

	Context("status command", func() {
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
					ghttp.VerifyRequest("GET", snapshotEndpoint+"/"+snapshotRepository+"/"+snapshotName+"/_status"),
					ghttp.RespondWith(http.StatusOK, statusResponse),
				),
			)

			cmd := &statusCmd{
				host:       elasticHost,
				port:       elasticPort,
				snapshot:   snapshotName,
				repository: snapshotRepository,
				authFile:   ""}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})

		It("should run with basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", snapshotEndpoint+"/"+snapshotRepository+"/"+snapshotName+"/_status"),
					ghttp.RespondWith(http.StatusOK, statusResponse),
				),
			)

			authFile := createAuthFile()
			defer os.Remove(authFile)

			cmd := &statusCmd{
				host:       elasticHost,
				port:       elasticPort,
				snapshot:   snapshotName,
				repository: snapshotRepository,
				authFile:   authFile}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})
	})

	Context("restore command", func() {
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
					ghttp.VerifyRequest("POST", snapshotEndpoint+"/"+snapshotRepository+"/"+snapshotName+"/_restore"),
					ghttp.RespondWith(http.StatusOK, ""),
				),
			)

			cmd := &restoreCmd{
				host:       elasticHost,
				port:       elasticPort,
				snapshot:   snapshotName,
				repository: snapshotRepository,
				authFile:   ""}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})

		It("should run with basic auth", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", snapshotEndpoint+"/"+snapshotRepository+"/"+snapshotName+"/_restore"),
					ghttp.RespondWith(http.StatusOK, ""),
				),
			)

			authFile := createAuthFile()
			defer os.Remove(authFile)

			cmd := &restoreCmd{
				host:       elasticHost,
				port:       elasticPort,
				snapshot:   snapshotName,
				repository: snapshotRepository,
				authFile:   authFile}

			exitStatus := cmd.Execute(nil, nil)

			Expect(exitStatus).Should(Equal(subcommands.ExitSuccess))
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
		})
	})
})
