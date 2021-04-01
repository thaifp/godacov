// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/jainpiyush19/godacov/coverage"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

const url = "%s/2.0/coverage/%s/go"
const defaultApiBase = "https://api.codacy.com"

var coverageFile string
var codacyToken string
var commitHash string
var codacyAPIBase string
var allowInsecure bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "godacov",
	Short: "Allows to publish go coverage reports to codacy",
	Long: `codagov allows to send go test coverage reports to codacy.
It transforms the report into a json format accepted by codacy
and publishes it with codacy's api using the provided commit 
id and project token.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.Flags().StringVarP(&coverageFile, "report", "r", "", "coverage report file generated by 'go test'")
	RootCmd.Flags().StringVarP(&codacyToken, "token", "t", "", "Codacy project token")
	RootCmd.Flags().StringVarP(&commitHash, "commit", "c", "", "The hash of the commit to provide coverage for")
	RootCmd.Flags().StringVarP(&codacyAPIBase, "api-base", "a", defaultApiBase, "The base URL of the codacy API server to use")
	RootCmd.Flags().BoolVarP(&allowInsecure, "allow-insecure", "i", false, "Allow insecure connection to base URL")
}

func run(cmd *cobra.Command, args []string) {

	if coverageFile == "" {
		fmt.Println("Error: A coverage file is required.")
		cmd.Usage()
		os.Exit(1)
	}
	if codacyToken == "" {
		fmt.Println("Error: A project token is required.")
		cmd.Usage()
		os.Exit(1)
	}
	if commitHash == "" {
		fmt.Println("Error: A commit hash is required.")
		cmd.Usage()
		os.Exit(1)
	}

	json, err := coverage.GenerateCoverageJSON(coverageFile)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf(url, codacyAPIBase, commitHash), bytes.NewBuffer(json))
	if err != nil {
		panic(err)
	}
	req.Header.Set("project_token", codacyToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	if allowInsecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.Status == "200 OK" {
		fmt.Println("Successfully posted coverage to Codacy.")
	} else {
		body, _ := ioutil.ReadAll(res.Body)
		fmt.Println("Failed to post: ", string(body))
	}
}
