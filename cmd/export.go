/*
Copyright © 2023 Agustin Larreinegabe

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	aacPath      *string
	outputFormat *string
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		exportData()
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	aacPath = exportCmd.Flags().StringP("path", "p", "access-config.yaml", "Path to the output YAML file")

	outputFormat = exportCmd.Flags().StringP("format", "f", "yaml", "Format of access-config. By default yaml. Option json yaml")

	exportCmd.MarkFlagFilename("path", "yaml")
}

func exportData() {
	jenkinsURL := viper.GetString("jenkins_url")
	jenkinsToken := viper.GetString("adm_url")

	var (
		accessConfig ExportData
		users        JenkinsUsersResponse
		globalRoles  JenkinsRolesResponse
		// itemRoles    JenkinsRolesResponse
		err error
	)

	accessConfig.ExtractDate = time.Now().Format("2006-01-02")
	accessConfig.JenkinsURL = jenkinsURL

	users, err = getUsers(jenkinsURL, jenkinsToken)

	if err != nil {
		log.Panic("Error getting users", err)
	}

	for _, user := range users.Users {
		var newUser User = User{
			FullName: user.User.FullName,
			Login:    user.User.ID,
		}
		for _, mail := range user.User.Property {
			if mail.Class == "hudson.tasks.Mailer$UserProperty" {
				newUser.Mail = mail.Address
				break
			}
		}
		accessConfig.Users = append(accessConfig.Users, newUser)
	}

	globalRoles, err = getRoles("globalRoles", jenkinsURL, jenkinsToken)

	for roleName, members := range globalRoles {

		newRole := Role{
			Name: roleName,
		}

		accessConfig.GlobalRoles = append(accessConfig.GlobalRoles, newRole)
		newMembership := Membership{
			RoleName: newRole.Name,
			RoleType: "global",
		}

		for _, member := range members {
			newMembership.Members = append(newMembership.Members, member.SID)
		}

		accessConfig.Membership = append(accessConfig.Membership, newMembership)
	}

	// itemRoles, err = getRoles("projectRoles", jenkinsURL, jenkinsToken)
}

func getUsers(jenkinsURL, apiToken string) (JenkinsUsersResponse, error) {
	client := &http.Client{}
	var users JenkinsUsersResponse

	req, err := http.NewRequest("GET", jenkinsURL+"asynchPeople/api/json?tree=users[user[fullName,id,mail,property[address]]]", nil)
	if err != nil {
		return users, err
	}

	req.Header.Add("Authorization", "Bearer "+apiToken)

	resp, err := client.Do(req)
	if err != nil {
		return users, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return users, err
	}

	err = json.Unmarshal(body, &users)
	if err != nil {
		return users, err
	}

	return users, nil
}

func getRoles(scope, jenkinsURL, apiToken string) (JenkinsRolesResponse, error) {
	client := &http.Client{}
	var roles JenkinsRolesResponse

	req, err := http.NewRequest("GET", jenkinsURL+"asynchPeople/api/json?tree=users[user[fullName,id,mail,property[address]]]", nil)
	if err != nil {
		return roles, err
	}

	req.Header.Add("Authorization", "Bearer "+apiToken)

	resp, err := client.Do(req)
	if err != nil {
		return roles, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return roles, err
	}

	err = json.Unmarshal(body, &roles)
	if err != nil {
		return roles, err
	}

	return roles, nil
}

func LoadConfig(filename string) (*ExportData, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config ExportData
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// SaveConfig saves the configuration to a YAML file.
func SaveConfig(filename string, config *ExportData) error {

	var data []byte
	var err error

	if *outputFormat == "json" {
		filename = filename + ".json"
		data, err = json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("error al convertir a JSON: %w", err)
		}
	} else {
		filename = filename + ".yaml"
		data, err = yaml.Marshal(config)
		if err != nil {
			return err
		}
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
