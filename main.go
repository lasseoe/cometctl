package main

/*

Rudimentary tool for listing backup jobs on Comet Backup servers.
This is a work in progress, expect changes.

License: MIT
Copyright (c) 2023 Lasse Østerild

*/

import (
	"flag"
	"log"
	"os"
	"time"

	sdk "github.com/CometBackup/comet-go-sdk"
	"github.com/olekukonko/tablewriter"
)

type Client struct {
	client *sdk.CometAPIClient
	users  map[string]*sdk.UserProfileConfig
}

func NewClient(url, username, password string) (*Client, error) {
	client, err := sdk.NewCometAPIClient(url, username, password)
	if err != nil {
		return nil, err
	}

	// initialize map used for caching User Profiles
	users := make(map[string]*sdk.UserProfileConfig)

	return &Client{
		client: client,
		users:  users,
	}, nil
}

// GetUserProfile retrieves user profile from Comet server
// unless it is already present in map
func (c *Client) GetUserProfile(user string) error {
	if _, exists := c.users[user]; exists {
		return nil
	} else {
		up, err := c.client.AdminGetUserProfile(user)
		if err != nil {
			return err
		}

		c.users[user] = up
	}

	return nil
}

func (c *Client) GetProtectedItemName(user, sourceguid string) string {
	err := c.GetUserProfile(user)
	if err != nil {
		log.Fatal("failed to get user profile", err)
	}

	// only return if it exists, this should always be the case though
	if _, exists := c.users[user].Sources[sourceguid]; exists {
		return c.users[user].Sources[sourceguid].Description
	}

	// we really should also return an error
	return "<<unknown>>"
}

// ListJobs prints jobs to stdout given start and end times
func (c *Client) ListJobs(start, end int64) error {
	var lines = [][]string{}

	jobs, err := c.client.AdminGetJobsForDateRange(int(start), int(end))
	if err != nil {
		return err
	}

	for _, j := range jobs {
		var e, pitem string
		if j.EndTime == 0 {
			e = "incomplete"
		} else {
			e = time.Unix(j.EndTime, 0).Format("2006-01-02 15:04:05")
		}
		if j.SourceGUID == "" {
			pitem = ""
		} else {
			pitem = c.GetProtectedItemName(j.Username, j.SourceGUID)
		}

		line := []string{
			j.Username,
			pitem,
			jobClassificationText(j.Classification),
			jobStatusText(j.Status),
			time.Unix(j.StartTime, 0).Format("2006-01-02 15:04:05"),
			e,
		}
		lines = append(lines, line)

		// fmt.Printf("%-30s  %-40.40s  %-14s  %-10s  %s  %s\n", j.Username, pitem, jobClassificationText(j.Classification), jobStatusText(j.Status), time.Unix(j.StartTime, 0).Format("2006-01-02 15:04:05"), e)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"User", "Protected Item", "Type", "Status", "Start", "End"})
	table.SetAutoWrapText(false)
	table.AppendBulk(lines)
	table.Render()

	return nil
}

func main() {
	url := flag.String("s", "https://localhost:8060", "URL for the Comet Server API")
	username := flag.String("u", "", "Username to authenticate with")
	password := flag.String("p", "", "Password to authenticate with")

	since := flag.String("since", "8 hours ago", "List jobs since given date")
	until := flag.String("until", "now", "List jobs until given date")

	//incomplete := flag.Bool("i", false, "List only incomplete jobs")

	flag.Parse()

	client, err := NewClient(*url, *username, *password)
	if err != nil {
		log.Fatal("Error creating client: ", err)
	}

	s, err := parseDate(*since)
	if err != nil {
		log.Fatal("failed to parse 'since' date", err)
	}
	e, err := parseDate(*until)
	if err != nil {
		log.Fatal("failed to parse 'until' date", err)
	}

	client.ListJobs(s, e)
}
