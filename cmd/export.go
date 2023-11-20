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
	"sort"
	"time"

	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	aacPath      *string
	outputFormat *string
	verbose      *bool
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
	verbose = exportCmd.Flags().Bool("v", false, "Verbose all responses")

	outputFormat = exportCmd.Flags().StringP("format", "f", "yaml", "Format of access-config. By default yaml. Option json yaml")

	exportCmd.MarkFlagFilename("path", "yaml")
}

func exportData() {
	jenkinsURL := viper.GetString("jenkins_url")
	jenkinsToken := viper.GetString("admin_token")
	jenkinsUser := viper.GetString("admin_user")

	var (
		accessConfig ExportData
		err          error
	)

	accessConfig.ExtractDate = time.Now().Format("2006-01-02")
	accessConfig.JenkinsURL = jenkinsURL

	if err != nil {
		log.Panic("Error getting users", err)
	}

	for _, user := range getUsers(jenkinsURL, jenkinsUser, jenkinsToken).Users {
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

	for roleName, members := range getRoles("globalRoles", jenkinsURL, jenkinsUser, jenkinsToken) {

		accessConfig.GlobalRoles = append(accessConfig.GlobalRoles, Role{
			Name: roleName,
		})
		newMembership := Membership{
			RoleName: roleName,
			RoleType: "global",
		}

		for _, member := range members {
			newMembership.Members = append(newMembership.Members, member.SID)
		}

		accessConfig.Membership = append(accessConfig.Membership, newMembership)
	}

	for roleName, members := range getRoles("projectRoles", jenkinsURL, jenkinsUser, jenkinsToken) {

		accessConfig.ItemsRoles = append(accessConfig.ItemsRoles, Role{
			Name: roleName,
		})
		newMembership := Membership{
			RoleName: roleName,
			RoleType: "item",
		}

		for _, member := range members {
			newMembership.Members = append(newMembership.Members, member.SID)
		}

		accessConfig.Membership = append(accessConfig.Membership, newMembership)
	}

	for roleName, members := range getRoles("slaveRoles", jenkinsURL, jenkinsUser, jenkinsToken) {

		newRole := Role{
			Name: roleName,
		}

		accessConfig.SlaveRoles = append(accessConfig.SlaveRoles, newRole)
		newMembership := Membership{
			RoleName: newRole.Name,
			RoleType: "item",
		}

		for _, member := range members {
			newMembership.Members = append(newMembership.Members, member.SID)
		}

		accessConfig.Membership = append(accessConfig.Membership, newMembership)
	}

	getRolesPermissions("globalRoles", jenkinsURL, jenkinsUser, jenkinsToken, &(accessConfig.GlobalRoles))
	getRolesPermissions("projectRoles", jenkinsURL, jenkinsUser, jenkinsToken, &(accessConfig.ItemsRoles))
	getRolesPermissions("slaveRoles", jenkinsURL, jenkinsUser, jenkinsToken, &(accessConfig.SlaveRoles))

	saveConfig("test-agus", &accessConfig)
}

func getUsers(jenkinsURL, jenkinsUser, apiToken string) JenkinsUsersResponse {
	client := &http.Client{}
	var users JenkinsUsersResponse

	req, err := http.NewRequest("GET", jenkinsURL+"asynchPeople/api/json?tree=users[user[fullName,id,mail,property[address]]]", nil)
	if err != nil {
		// return users, err
		log.Panic(err)

	}

	req.SetBasicAuth(jenkinsUser, apiToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err)

	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}
	err = json.Unmarshal(body, &users)
	if err != nil {
		log.Panic(err)

	}

	return users
}

func getRoles(roleScope, jenkinsURL, jenkinsUser, apiToken string) JenkinsRolesResponse {
	client := &http.Client{}
	var roles JenkinsRolesResponse

	req, err := http.NewRequest("GET", jenkinsURL+"role-strategy/strategy/getAllRoles?type="+roleScope, nil)
	if err != nil {
		log.Panic(err)
	}

	req.SetBasicAuth(jenkinsUser, apiToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}

	// fmt.Printf("%s\\n", body)

	err = json.Unmarshal(body, &roles)
	if err != nil {
		log.Panic(err)
	}

	return roles
}

func getRolesPermissions(roleScope, jenkinsURL, jenkinsUser, apiToken string, currentRoles *[]Role) {
	var (
		client           http.Client
		url              string
		rolesPermissions JenkinsRolePermissionResponse
	)

	for item, role := range *currentRoles {

		url = jenkinsURL + "role-strategy/strategy/getRole?type=" + roleScope + "&roleName=" + role.Name
		rolesPermissions = JenkinsRolePermissionResponse{}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Panic(err)
		}
		req.SetBasicAuth(jenkinsUser, apiToken)

		resp, err := client.Do(req)
		if err != nil {
			log.Panic(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Panic(err)
		}

		if *verbose {
			log.Printf("\n URL=%s \n Response=\n %s\n", url, body)
		}

		err = json.Unmarshal(body, &rolesPermissions)
		if err != nil {
			log.Panic(err)
		}

		rolePermissions := make([]string, 0)
		for permission, _ := range rolesPermissions.PermissionsIds {
			rolePermissions = append(rolePermissions, permission)
		}

		sort.Strings(rolePermissions)

		(*currentRoles)[item].Permissions = rolePermissions
	}

}

func loadConfig(filename string) (*ExportData, error) {
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
func saveConfig(filename string, config *ExportData) error {

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
