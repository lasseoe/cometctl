package main

/*

yes, utils.go is a terrible name, it is what it is, for now anyway.

License: MIT
Copyright (c) 2023 Lasse Østerild

*/

import (
	"time"

	sdk "github.com/CometBackup/comet-go-sdk"
	"github.com/ijt/go-anytime"
)

func jobStatusText(status sdk.JobStatus) string {
	switch {
	case status == sdk.JOB_STATUS_FAILED_ABANDONED:
		return "Abandoned"
	case status == sdk.JOB_STATUS_FAILED_CANCELLED:
		return "Cancelled"
	case status == sdk.JOB_STATUS_FAILED_ERROR:
		return "Error"
	case status == sdk.JOB_STATUS_FAILED_QUOTA:
		return "Quota"
	case status == sdk.JOB_STATUS_FAILED_SCHEDULEMISSED:
		return "Missed"
	case status == sdk.JOB_STATUS_FAILED_SKIPALREADYRUNNING:
		return "Skipped" // already running
	case status == sdk.JOB_STATUS_FAILED_TIMEOUT:
		return "Timeout"
	case status == sdk.JOB_STATUS_FAILED_WARNING:
		return "Warning"
	case (status >= sdk.JOB_STATUS_FAILED__MIN) && (status <= sdk.JOB_STATUS_FAILED__MAX):
		return "Failed" // 7xxx code, the job has stopped for an unsuccessful reason
	case status == sdk.JOB_STATUS_RUNNING_ACTIVE:
		return "Active"
	case status == sdk.JOB_STATUS_RUNNING_INDETERMINATE:
		return "huh?" // unused according to SDK
	case status == sdk.JOB_STATUS_RUNNING_REVIVED:
		return "Revived" // A backup job that was marked as stopped or abandoned, but has somehow continued to run
	case (status >= sdk.JOB_STATUS_RUNNING__MIN) && (status <= sdk.JOB_STATUS_RUNNING__MAX):
		return "Running" // 6xxx code, the job is still running.
	case status == sdk.JOB_STATUS_STOP_SUCCESS:
		return "Success"
	case (status >= sdk.JOB_STATUS_STOP_SUCCESS__MIN) && (status <= sdk.JOB_STATUS_STOP_SUCCESS__MAX):
		return "Success" // 5xxx code, the job has stopped for a successful reason
	}
	return "Unknown"
}

func jobClassificationText(status sdk.JobClassification) string {
	switch {
	case status == sdk.JOB_CLASSIFICATION_BACKUP:
		return "Backup"
	case status == sdk.JOB_CLASSIFICATION_DEEPVERIFY:
		return "DeepVerify"
	case status == sdk.JOB_CLASSIFICATION_DELETE_CUSTOM:
		return "DeleteCustom" // A specific snapshot has been deleted via the Restore wizard.
	case status == sdk.JOB_CLASSIFICATION_IMPORT:
		return "ImportSettings"
	case status == sdk.JOB_CLASSIFICATION_REINDEX:
		return "Re-index"
	case status == sdk.JOB_CLASSIFICATION_REMEASURE:
		return "Re-measure" // Explicitly re-measuring the size of a Vault (right-click > Advanced menu).
	case status == sdk.JOB_CLASSIFICATION_RESTORE:
		return "Restore"
	case status == sdk.JOB_CLASSIFICATION_RETENTION:
		return "Retention"
	case status == sdk.JOB_CLASSIFICATION_UNINSTALL:
		return "Uninstall"
	case status == sdk.JOB_CLASSIFICATION_UNKNOWN:
		return "Unknown" // hm?
	case status == sdk.JOB_CLASSIFICATION_UNLOCK:
		return "Clean locks" // Another process needed exclusive Vault access (e.g. for retention) but the process died. This task cleans up exclusive lockfiles.
	case status == sdk.JOB_CLASSIFICATION_UPDATE:
		return "Update"
	}
	return "Unknown"
}

func parseDate(date string) (int64, error) {
	var dateUnix int64

	timeNow := time.Now()

	if date == "" {
		dateUnix = timeNow.Unix()
	} else {
		dateParse, err := anytime.Parse(date, timeNow)
		if err != nil {
			return dateUnix, err
		}
		dateUnix = dateParse.Unix()
	}
	return dateUnix, nil
}
