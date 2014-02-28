package utils

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"github.com/jostillmanns/rdiff-backup-tools/types"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// func LastBackup(path string) string, error {

// }

type Repository struct {
	BasePath, DirectoryStructure, Origin string
	Name                                 string
}

func InitRepositories() ([]Repository, error) {
	infos, err := ioutil.ReadDir(filepath.Clean("repositories"))
	if err != nil {
		fmt.Errorf("unable to read repositories directory", err)
	}

	repos := make([]Repository, 0)
	for _, e := range infos {
		if !e.IsDir() {
			continue
		}

		repos = append(repos,
			Repository{
				filepath.Join("repositories", e.Name(), "basepath"),
				filepath.Join("repositories", e.Name(), "directorystructure"),
				filepath.Join("repositories", e.Name(), "origin"),
				e.Name(),
			})
	}

	return repos, nil
}

func (me *Repository) Snapshotfiles() ([]string, error) {
	files := make([]string, 0)
	dataDir := filepath.Join(me.BasePath, types.DATA)

	fileInfos, err := ioutil.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	for _, e := range fileInfos {
		if len(e.Name()) < len(types.SNAPSHOT) {
			continue
		}

		if e.Name()[len(e.Name())-len(types.SNAPSHOT):] != types.SNAPSHOT {
			continue
		}

		if e.Name()[:len(types.EXT_ATTRIB)] == types.EXT_ATTRIB {
			continue
		}

		files = append(files, e.Name())
	}

	return files, nil
}

func (me *Repository) Unzips(snapshotfiles []string) error {
	for _, e := range snapshotfiles {
		snapbytes, err := os.Open(filepath.Join(me.BasePath, types.DATA, e))
		targetName := strings.Replace(e, types.GZ, "", 1)
		target, err := os.Create(filepath.Join(me.DirectoryStructure, targetName))
		if err != nil {
			return fmt.Errorf("unable to create target", err)
		}
		reader, err := gzip.NewReader(snapbytes)
		if err != nil {
			return fmt.Errorf("unable to read file", targetName, err)
		}
		_, err = io.Copy(target, reader)
		if err != nil {
			return fmt.Errorf("unable to copy target data", err)
		}
		target.Close()
		reader.Close()
		target.Close()
	}

	return nil
}

func ReadDirectoriesFromSnapshot(snapshot *os.File) []string {
	directories := make([]string, 0)

	scanner := bufio.NewScanner(snapshot)
	var file string
	window := 0
	for scanner.Scan() {
		s := scanner.Text()
		if len(s) >= 6 && s[:6] != "File ." && s[:5] == "File " {
			file = s[5:]
			window = 0
		}
		if strings.Contains(s, "Type dir") && window == 1 {
			directories = append(directories, file)
		}
		window += 1
	}
	return directories
}

func (me *Repository) InitDirectoryStructure(base string, files []string) error {
	directoryDir := filepath.Clean(me.DirectoryStructure)
	file, err := os.Create(filepath.Join(directoryDir, base))
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range files {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func (me *Repository) Contains(file string, snapshot *os.File) bool {
	scanner := bufio.NewScanner(snapshot)

	for scanner.Scan() {
		if scanner.Text() == file {
			return true
		}
	}
	return false
}

func Unquote(timestamp string) (string, error) {
	for -1 != strings.LastIndex(timestamp, ";") {
		index := strings.LastIndex(timestamp, ";")
		intval, err := strconv.ParseInt(timestamp[index+1:index+4], 10, 8)
		if err != nil {
			return "", fmt.Errorf("unable to parse int", err)
		}

		timestamp = timestamp[0:index] + string(intval) + timestamp[index+4:len(timestamp)]
	}

	return timestamp, nil
}

func MetafileToTimepoint(filename, ext string) (time.Time, error) {
	i := len(filename) - len(ext)
	timestamp := filepath.Ext(filename[:i])[1:]
	timestamp, err := Unquote(timestamp)
	if err != nil {
		return time.Time{}, err
	}
	timepoint, err := time.Parse(types.TIMESTAMP_FMT, timestamp)
	if err != nil {
		return time.Time{}, err
	}
	return timepoint, nil
}

func TrimExt(s string) string {
	ext := filepath.Ext(s)
	return s[0 : len(s)-len(ext)]
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func UpdateAllDirectories() error {
	repos, err := InitRepositories()
	if err != nil {
		return err
	}

	for _, e := range repos {

		snaps, err := e.Snapshotfiles()
		if err != nil {
			return err
		}

		for _, s := range snaps {
			needle := filepath.Join(e.DirectoryStructure, s[:len(s)-len(types.GZ)])
			_, err := os.Stat(needle)
			if os.IsNotExist(err) {
				e.Unzips([]string{s})
			} else if err != nil {
				return fmt.Errorf("unable to stat meta file", needle)
			}
		}

		err = e.UpdateDirectories()
		if err != nil {
			return err
		}
	}

	return nil
}

func (me *Repository) UpdateDirectories() error {
	snapshots, err := ioutil.ReadDir(me.DirectoryStructure)
	if err != nil {
		fmt.Errorf("unable to read directory container", err)
	}

	for _, e := range snapshots {
		if e.IsDir() {
			continue
		}

		if filepath.Ext(e.Name()) == types.DIREXT {
			continue
		}

		// check if snapshot is already set
		i, err := os.Stat(filepath.Join(me.DirectoryStructure, e.Name()+types.DIREXT))
		if i == nil && !os.IsNotExist(err) {
			return fmt.Errorf("error trying to read structure directory", err)
		} else if i != nil {
			continue
		}

		//this is where the magic happens
		file, err := os.Open(filepath.Join(me.DirectoryStructure, e.Name()))
		if err != nil {
			return fmt.Errorf("unable to read snapshot file", e.Name(), err)
		}

		files := ReadDirectoriesFromSnapshot(file)
		file.Close()

		err = me.InitDirectoryStructure(e.Name()+types.DIREXT, files)
		if err != nil {
			fmt.Errorf("unable to initialize directory structure", e.Name(), err)
		}
	}

	return nil
}

func (me *Repository) TimeStamps() ([]string, error) {
	infos, err := ioutil.ReadDir(filepath.Join(me.BasePath, types.DATA))
	if err != nil {
		return nil, fmt.Errorf("unable to read dir", types.DATA, err)
	}
	timestamps := make([]string, 0, len(infos))

	for _, e := range infos {
		if !strings.Contains(e.Name(), types.SESSIONSTATISTICS) && !(filepath.Ext(e.Name()) == types.DATAEXT) {
			continue
		}

		timestamps = append(timestamps, filepath.Ext(TrimExt(e.Name()))[1:])
	}
	return timestamps, nil
}

func (me *Repository) NextSnapshot(timepoint string) (string, error) {
	infos, err := ioutil.ReadDir(me.DirectoryStructure)
	if err != nil {
		return "", fmt.Errorf("unable parse snapshot directory", err)
	}
	for _, e := range infos {
		if filepath.Ext(e.Name()) == ".structure" {
			continue
		}
		input, err := time.Parse(types.TIMESTAMP_FMT, timepoint)
		if err != nil {
			return "", fmt.Errorf("unable to parse input time", timepoint, err)
		}
		t, err := MetafileToTimepoint(e.Name(), ".snapshot")
		if err != nil {
			return "", fmt.Errorf("unable to call metafileToTimepoint", e.Name(), ".snapshot", err)
		}

		if t.Before(input) {
			continue
		}
		return e.Name(), nil
	}
	return "", nil
}

func SplitPath(path string) []string {

	if len(path) == 0 {
		return []string{}
	}
	splitpath := make([]string, 0)

	b := (filepath.Base(path))
	for b != "." && b != "/" {
		splitpath = append(splitpath, b)
		path = filepath.Dir(path)
		b = (filepath.Base(path))
	}

	result := make([]string, len(splitpath))
	index := 0
	for i := len(splitpath); i > 0; i-- {
		result[index] = splitpath[i-1]
		index++
	}
	return result
}
