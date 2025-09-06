package main

/*

Rudimentary tool for listing backup jobs on Comet Backup servers.
This is a work in progress, expect changes.

License: MIT
Copyright (c) 2025 Lasse Østerild

*/

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	sdk "github.com/CometBackup/comet-go-sdk/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
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
func (c *Client) GetUserProfile(ctx context.Context, user string) error {
	if _, exists := c.users[user]; exists {
		return nil
	} else {
		up, err := c.client.AdminGetUserProfile(ctx, user)
		if err != nil {
			return err
		}

		c.users[user] = up
	}

	return nil
}

func (c *Client) GetProtectedItemName(ctx context.Context, user, sourceguid string) string {
	err := c.GetUserProfile(ctx, user)
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
func (c *Client) ListJobs(ctx context.Context, start, end int64) error {
	var lines = [][]string{}

	jobs, err := c.client.AdminGetJobsForDateRange(ctx, int(start), int(end))
	if err != nil {
		return err
	}

	for _, j := range jobs {
		var e, pitem string

		// jobs with EndTime 0 are "incomplete"
		if j.EndTime == 0 {
			e = "incomplete"
		} else {
			e = time.Unix(j.EndTime, 0).Format("2006-01-02 15:04:05")
		}
		// if job doesn't have a protected item GUID, set it to blank
		if j.SourceGUID == "" {
			pitem = ""
		} else {
			pitem = c.GetProtectedItemName(ctx, j.Username, j.SourceGUID)
		}

		// build table slice
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

	// https://github.com/olekukonko/tablewriter/blob/master/MIGRATION.md
	cfg := tablewriter.Config{
		Header: tw.CellConfig{
			Alignment:  tw.CellAlignment{Global: tw.AlignCenter},
			Formatting: tw.CellFormatting{AutoFormat: tw.Off},
		},
		Row: tw.CellConfig{
			Alignment:  tw.CellAlignment{Global: tw.AlignLeft},
			Formatting: tw.CellFormatting{AutoFormat: tw.Off},
		},
		//MaxWidth: 80,
		Behavior: tw.Behavior{TrimSpace: tw.On},
	}

	table := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(cfg))
	table.Header("User", "Protected Item", "Type", "Status", "Start", "End")
	table.Bulk(lines)
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

	// required for Comet SDK v2.x
	ctx := context.Background()

	client.ListJobs(ctx, s, e)
}
