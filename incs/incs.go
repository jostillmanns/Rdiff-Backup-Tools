package incs

import (
	"fmt"
	"github.com/jostillmanns/rdiff-backup-tools/types"
	"github.com/jostillmanns/rdiff-backup-tools/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func Files(me *utils.Repository, path string, t time.Time) ([]string, error) {
	workingdir := filepath.Join(me.BasePath, types.DATA, types.INCREMENTS, path)
	times, err := me.TimeStamps()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve timestamps", types.DATA, err)
	}

	_, err = os.Stat(filepath.Join(me.BasePath, path))
	files := []os.FileInfo{}
	if err == nil {
		files, err = ioutil.ReadDir(filepath.Join(me.BasePath, path))
		if err != nil {
			return nil, fmt.Errorf("unable to parse directory", filepath.Join(me.BasePath, path), err)
		}
	}

	increments, err := ioutil.ReadDir(filepath.Join(me.BasePath, types.DATA, types.INCREMENTS, path))
	if err != nil {
		return nil, fmt.Errorf("unable to parse directory", filepath.Join(me.BasePath, types.DATA, types.INCREMENTS, path), err)
	}

	resultsetphy, err := FilterFiles(files, increments, times, t, workingdir)
	if err != nil {
		return nil, err
	}

	// deleted files are not to be found in the physical representation of files
	var filesinc []os.FileInfo
	for i, _ := range increments {
		if increments[i].IsDir() {
			continue
		}
		if filepath.Ext(utils.TrimExt(increments[i].Name())) != ".snapshot" {
			continue
		}

		needle := utils.TrimExt(utils.TrimExt(utils.TrimExt(increments[i].Name())))
		d, err := os.Stat(filepath.Join(me.BasePath, types.DATA, types.INCREMENTS, path, needle))
		if err == nil && d.IsDir() {
			// points to a directory, ignore this file
			continue
		} else if err != nil && !os.IsNotExist(err) {
			// real error case
			return nil, fmt.Errorf("unable to stat file", needle, err)
		}
		// check if resultphy already contains this file
		flag := false
		for _, e := range resultsetphy {
			if e == needle {
				flag = true
				break
			}
		}
		if !flag {
			filesinc = append(filesinc, &types.RdiffFileInfo{Name_: needle, IsDir_: false})
		}
	}
	resultsetinc, err := FilterFiles(filesinc, increments, times, t, workingdir)
	if err != nil {
		return nil, err
	}

	resultset := make([]string, len(resultsetphy)+len(resultsetinc))
	copy(resultset, resultsetphy)
	copy(resultset[len(resultsetphy):], resultsetinc)

	return resultset, nil
}

func FilterFiles(files, increments []os.FileInfo, times []string, t time.Time, workingdir string) ([]string, error) {
	resultset := make([]string, 0)
	metafiles := make([]string, 0, len(files))
	for _, e := range increments {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(utils.TrimExt(e.Name())) != ".snapshot" && filepath.Ext(e.Name()) != ".missing" {
			continue
		}
		unquote, err := utils.Unquote(e.Name())
		if err != nil {
			return nil, err
		}
		metafiles = append(metafiles, unquote)
	}

	for i, _ := range files {
		if files[i].IsDir() {
			continue
		}

		break_ := false
		for _, e := range times {
			unquoted, err := utils.Unquote(e)
			if err != nil {
				return nil, err
			}
			parsedtime, err := time.Parse(types.TIMESTAMP_FMT, unquoted)
			if err != nil {
				return nil, fmt.Errorf("unable to parse time", e, err)
			}

			// t >= parsedtime
			if t.After(parsedtime) {

				if !utils.StringInSlice(files[i].Name()+"."+unquoted+".snapshot.gz", metafiles) {
					continue
				}

				// file was deleted in the past, dont add it to the resultset
				break_ = true
				break
			}

			if t.Before(parsedtime) || t.Equal(parsedtime) {
				if !utils.StringInSlice(files[i].Name()+"."+unquoted+".missing", metafiles) {
					continue
				}

				// file was created in the future, dont add it to the resultset
				break_ = true
				break
			}
		}
		if break_ {
			continue
		}
		resultset = append(resultset, files[i].Name())
	}

	return resultset, nil
}

func Directories(me *utils.Repository, path string, t time.Time) ([]string, error) {
	workingdir := filepath.Join(me.BasePath, types.DATA, types.INCREMENTS, path)
	resultset := make([]string, 0)
	times, err := me.TimeStamps()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve timestamps", types.DATA, err)
	}

	directories, err := ioutil.ReadDir(workingdir)
	if err != nil {
		return nil, fmt.Errorf("unable to parse directory", workingdir, err)
	}

	metafiles := make([]string, 0, len(directories))
	for _, e := range directories {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) != ".dir" && filepath.Ext(e.Name()) != ".missing" {
			continue
		}
		unquote, err := utils.Unquote(e.Name())
		if err != nil {
			return nil, err
		}
		metafiles = append(metafiles, unquote)
	}

	for i, _ := range directories {
		if !directories[i].IsDir() {
			continue
		}

		lastchange := ""
		addflag := false
		for _, e := range times {
			addflag = false

			unquoted, err := utils.Unquote(e)
			if err != nil {
				return nil, err
			}
			parsedtime, err := time.Parse(types.TIMESTAMP_FMT, unquoted)
			if err != nil {
				return nil, fmt.Errorf("unable to parse time", e, err)
			}

			// t >= parsedtime
			if t.After(parsedtime) {
				exists := utils.StringInSlice(directories[i].Name()+"."+unquoted+types.DIR, metafiles)
				if !exists {
					continue
				}
				lastchange = unquoted
				continue
			}

			direxists := utils.StringInSlice(directories[i].Name()+"."+unquoted+types.DIR, metafiles)
			missingexists := utils.StringInSlice(directories[i].Name()+"."+unquoted+types.MISSING, metafiles)

			if t.Equal(parsedtime) {
				if !direxists && !missingexists {
					// if there is no such files, there is nothing to do
					continue
				} else if !missingexists && direxists {
					// there exists a .dir file in present time
					lastchange = unquoted
					continue
				} else if missingexists && !direxists {
					// there exist a .missing file in present time
					// file was created right on requested point of time, we take it
					resultset = append(resultset, directories[i].Name())
					lastchange = ""
					addflag = true
					break
				}
			}

			if t.Before(parsedtime) {
				if !direxists && !missingexists {
					// if there is non of such files, there is nothing to do
					continue
				} else if !direxists && missingexists {
					// .missing exists, that means directory was created after current point of time, we dont take it
					// toogled addflag, with an empty lastchange means the directory will not be added
					addflag = true
					break
				} else if direxists && !missingexists {
					// there exists a change on that directory after the current point of time, so it has to be available
					resultset = append(resultset, directories[i].Name())
					lastchange = ""
					addflag = true
					break
				}
			}
		}
		// if addflag is not set we have to check if the last change was a delete
		if !addflag && lastchange != "" {
			nextSnap, err := me.NextSnapshot(lastchange)
			if err != nil {
				return nil, fmt.Errorf("unable to retrieve next snapshot", err)
			}
			snapshot, err := os.Open(filepath.Join(me.DirectoryStructure, nextSnap+types.DIREXT))
			if err != nil {
				return nil, fmt.Errorf("unable to open directorie structure file")
			}
			if !me.Contains(filepath.Join(path, directories[i].Name()), snapshot) {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("unable to stat in directory", filepath.Join(me.DirectoryStructure, nextSnap+".structure", path))
			}
		}
		if addflag && lastchange == "" {
			continue
		}

		// passed all checks, add it
		resultset = append(resultset, directories[i].Name())
	}
	return resultset, nil
}
