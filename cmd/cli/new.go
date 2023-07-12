package main

import "strings"

func doNew(appName string) {
	appName = strings.ToLower(appName)

	// sanitize application name (convert url to single word)
	if strings.Contains(appName, "/") {
		sploded := strings.SplitAfter(appName, "/")
		appName = sploded[(len(sploded) - 1)]
	}

	// git clone the skeleton app

	// remove .git directory

	// create a ready to go .env file

	// create a makefile

	// update the go.mod file

	// update existing .go files with correct name/imports

	// run go mod tidy in the project directory
}
