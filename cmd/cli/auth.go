package main

import (
	"fmt"
	"github.com/fatih/color"
	"log"
	"path/filepath"
	"time"
)

func doAuth() error {
	rootPath := filepath.ToSlash(fin.RootPath)
	//migrations
	dbType := fin.DB.DataType
	fileName := fmt.Sprintf("%d_create_auth_tables", time.Now().UnixMicro())
	upFile := rootPath + "/migrations/" + fileName + ".up.sql"
	downFile := rootPath + "/migrations/" + fileName + ".down.sql"

	log.Println(dbType, upFile, downFile)

	err := copyFileFromTemplate("templates/migrations/auth_tables."+dbType+".sql", upFile)
	if err != nil {
		exitGracefully(err)
	}

	err = copyDataToFile([]byte("drop table if exists users cascade; drop table if exists tokens cascade; drop table if exists remember_tokens cascade;"), downFile)
	if err != nil {
		exitGracefully(err)
	}

	// run migrations
	err = doMigrate("up", "")
	if err != nil {
		exitGracefully(err)
	}

	// copy files
	err = copyFileFromTemplate("templates/data/user.go.txt", rootPath+"/data/user.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/data/token.go.txt", rootPath+"/data/token.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/data/remember_token.go.txt", rootPath+"/data/remember_token.go")
	if err != nil {
		exitGracefully(err)
	}

	// copy middleware
	err = copyFileFromTemplate("templates/middleware/auth.go.txt", rootPath+"/middleware/auth.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/middleware/auth-token.go.txt", rootPath+"/middleware/auth-token.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/middleware/remember.go.txt", rootPath+"/middleware/remember.go")
	if err != nil {
		exitGracefully(err)
	}

	// copy handlers
	err = copyFileFromTemplate("templates/handlers/auth-handlers.go.txt", rootPath+"/handlers/auth-handlers.go")
	if err != nil {
		exitGracefully(err)
	}

	// copy mail templates
	err = copyFileFromTemplate("templates/mailer/password-reset.html.tmpl", rootPath+"/mailer/password-reset.html.tmpl")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/mailer/password-reset.plain.tmpl", rootPath+"/mailer/password-reset.plain.tmpl")
	if err != nil {
		exitGracefully(err)
	}

	// copy views
	err = copyFileFromTemplate("templates/views/login.jet", rootPath+"/views/login.jet")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/views/forgot.jet", rootPath+"/views/forgot.jet")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/views/reset-password.jet", rootPath+"/views/reset-password.jet")
	if err != nil {
		exitGracefully(err)
	}

	color.Yellow("  - users, tokens and remember_tokens migrations created and executed")
	color.Yellow("  - user and token models created")
	color.Yellow("  - auth middleware created")
	color.Yellow("")
	color.Yellow("Reminder: add user and token models in data/models.go")
	color.Yellow("Reminder: add auth.go and auth-token.go middleware to your routes")

	return nil
}
