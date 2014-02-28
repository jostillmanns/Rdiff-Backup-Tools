// Copyright (c) 2013 ICRL

// See the file license.txt for copying permission.

package main

import (
	"flag"
	"fmt"
	"github.com/jostillmanns/rdiff-backup-tools/testdata"
	"github.com/jostillmanns/rdiff-backup-tools/utils"
	"github.com/jostillmanns/rdiff-backup-tools/web"
	"path/filepath"
)

var operation string
var permission bool

func main() {
	flag.StringVar(&operation, "o", "", "available operations: update, web, generatetest")
	flag.BoolVar(&permission, "p", false, "set permissions after restoring")
	flag.Parse()

	if operation == "update" {
		err := utils.UpdateAllDirectories()
		if err != nil {
			fmt.Println(err)
		}

	} else if operation == "web" {
		web.StartServer(permission)
	} else if operation == "generatetest" {
		repo := utils.Repository{
			filepath.Join("repositories", "TestRepository", "basepath"),
			filepath.Join("repositories", "TestRepository", "directorystructure"),
			filepath.Join("repositories", "TestRepository", "origin"),
			"TestRepository",
		}
		out, err := testdata.GenerateTestRepo(repo)
		if err != nil {
			fmt.Println(string(out), err)
			testdata.Clean(repo)
		}
	}
}
