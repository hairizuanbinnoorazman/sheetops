// Package appmgr provide a higher level abstraction to make it easier to fetch values from googlesheets
package appmgr

type AppSetting struct {
	Name     string
	Image    string
	Replicas int
}
