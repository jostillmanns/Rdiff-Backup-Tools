// Copyright (c) 2013 ICRL

// See the file license.txt for copying permission.

package types

import (
	"os"
	"time"
)

const (
	DATA              = "rdiff-backup-data"
	INCREMENTS        = "increments"
	SNAPSHOT          = "snapshot.gz"
	MISSING           = ".missing"
	GZ                = ".gz"
	DIREXT            = ".structure"
	EXT_ATTRIB        = "extended_attributes"
	DIR               = ".dir"
	DATAEXT           = ".data"
	SESSIONSTATISTICS = "session_statistics."
)

const TIMESTAMP_FMT = "2006-01-02T15:04:05-07:00"

type RdiffFileInfo struct {
	Name_    string
	IsDir_   bool
	ModTime_ time.Time
}

func (me *RdiffFileInfo) Name() string {
	return me.Name_
}

func (me *RdiffFileInfo) IsDir() bool {
	return me.IsDir_
}

func (me *RdiffFileInfo) Size() int64 {
	return 0
}

func (me *RdiffFileInfo) Mode() os.FileMode {
	return 0
}

func (me *RdiffFileInfo) ModTime() time.Time {
	return me.ModTime_
}

func (me *RdiffFileInfo) Sys() interface{} {
	return nil
}
