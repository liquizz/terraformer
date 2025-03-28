// Copyright 2018 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package github

import (
	"os"
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
)

type GithubProvider struct { //nolint
	terraformutils.Provider
	owner          string
	token          string
	baseURL        string
	appID          int64
	installationID int64
	pem            string
}

func (p GithubProvider) GetResourceConnections() map[string]map[string][]string {
	return map[string]map[string][]string{}
}

func (p GithubProvider) GetProviderData(arg ...string) map[string]interface{} {
	return map[string]interface{}{
		"provider": map[string]interface{}{
			"github": map[string]interface{}{
				"owner": p.owner,
			},
		},
	}
}

func (p *GithubProvider) GetConfig() cty.Value {
	if p.appID != 0 && p.installationID != 0 && p.pem != "" {
		return cty.ObjectVal(map[string]cty.Value{
			"owner": cty.StringVal(p.owner),
			"app_auth": cty.ListVal(
				[]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"id":              cty.NumberIntVal(p.appID),
						"installation_id": cty.NumberIntVal(p.installationID),
						"pem_file":        cty.StringVal(p.pem),
					}),
				},
			),
		})
	}
	return cty.ObjectVal(map[string]cty.Value{
		"owner":    cty.StringVal(p.owner),
		"token":    cty.StringVal(p.token),
		"base_url": cty.StringVal(p.baseURL),
	})
}

// Init GithubProvider with owner
func (p *GithubProvider) Init(args []string) error {
	if appIDValue, ok := os.LookupEnv("GITHUB_APP_ID"); ok {
		appID, err := strconv.ParseInt(appIDValue, 10, 64)
		if err != nil {
			return err
		}
		p.appID = appID
	}
	if installationIDValue, ok := os.LookupEnv("GITHUB_APP_INSTALLATION_ID"); ok {
		installationID, err := strconv.ParseInt(installationIDValue, 10, 64)
		if err != nil {
			return err
		}
		p.installationID = installationID
	}
	if pem, ok := os.LookupEnv("GITHUB_APP_PEM_FILE"); ok {
		p.pem = strings.Replace(pem, `\n`, "\n", -1)
	}

	p.owner = args[0]
	if len(args) < 2 {
		if os.Getenv("GITHUB_TOKEN") == "" {
			return errors.New("token requirement")
		}
		p.token = os.Getenv("GITHUB_TOKEN")
	} else {
		p.token = args[1]
	}
	if len(args) > 2 {
		if args[2] != "" {
			p.baseURL = args[2]
		} else {
			p.baseURL = githubDefaultURL
		}
	}
	return nil
}

func (p *GithubProvider) GetName() string {
	return "github"
}

func (p *GithubProvider) InitService(serviceName string, verbose bool) error {
	var isSupported bool
	if _, isSupported = p.GetSupportedService()[serviceName]; !isSupported {
		return errors.New(p.GetName() + ": " + serviceName + " not supported service")
	}
	p.Service = p.GetSupportedService()[serviceName]
	p.Service.SetName(serviceName)
	p.Service.SetVerbose(verbose)
	p.Service.SetProviderName(p.GetName())
	p.Service.SetArgs(map[string]interface{}{
		"owner":           p.owner,
		"token":           p.token,
		"base_url":        p.baseURL,
		"app_id":          p.appID,
		"installation_id": p.installationID,
		"pem":             p.pem,
	})
	return nil
}

// GetSupportedService return map of support service for Github
func (p *GithubProvider) GetSupportedService() map[string]terraformutils.ServiceGenerator {
	return map[string]terraformutils.ServiceGenerator{
		"members":               &MembersGenerator{},
		"organization":          &OrganizationGenerator{},
		"organization_blocks":   &OrganizationBlockGenerator{},
		"organization_projects": &OrganizationProjectGenerator{},
		"organization_webhooks": &OrganizationWebhooksGenerator{},
		"repositories":          &RepositoriesGenerator{},
		"teams":                 &TeamsGenerator{},
		"user_ssh_keys":         &UserSSHKeyGenerator{},
	}
}
