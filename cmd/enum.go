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

type ExportData struct {
	ExtractDate string       `json:"extract_date" yaml:"extract_date"`
	JenkinsURL  string       `json:"jenkins_url" yaml:"jenkins_url"`
	Users       []User       `json:"users" yaml:"users"`
	GlobalRoles []Role       `json:"global_roles" yaml:"global_roles"`
	ItemsRoles  []Role       `json:"items_roles" yaml:"items_roles"`
	Membership  []Membership `json:"membership" yaml:"membership"`
}

type User struct {
	Login    string `json:"login" yaml:"login"`
	FullName string `json:"full_name" yaml:"full_name"`
	Mail     string
}

type Role struct {
	Name        string       `json:"name" yaml:"name"`
	Permissions []Permission `json:"permissions" yaml:"permissions"`
}

type Permission struct {
	Name string `json:"name" yaml:"name"`
}

type Membership struct {
	RoleName string   `json:"role_name" yaml:"role_name"`
	Members  []string `json:"members" yaml:"members"`
	RoleType string   `json:"role_type" yaml:"role_type"`
}

// Jenkins
type JenkinsUsersResponse struct {
	Class string      `json:"_class"`
	Users []UserEntry `json:"users"`
}

type UserEntry struct {
	User UserDetails `json:"user"`
}

type UserDetails struct {
	FullName string     `json:"fullName"`
	ID       string     `json:"id"`
	Property []Property `json:"property"`
}

type Property struct {
	Class   string `json:"_class"`
	Address string `json:"address,omitempty"` // omitempty para ignorar si está vacío
}

type RoleMember struct {
	Type string `json:"type"`
	SID  string `json:"sid"`
}

type JenkinsRolesResponse map[string][]RoleMember
