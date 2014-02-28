package wrapper

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	RDIFF_BACKUP = "rdiff-backup"
	RESTORE_FLAG = "--restore-as-of"
	PARSABLE     = "--parsable-output"
	DATA         = "rdiff-backup-data"
	INCREMENTS   = "increments"
	SETFACL      = "setfacl"
	ACLGROUP     = "30006"
)

func clean(path *string) error {
	*path = filepath.Clean(*path)

	var err error
	if _, err = os.Stat(*path); os.IsNotExist(err) {
		return fmt.Errorf(*path, "not available")
	}

	if err != nil {
		return err
	}

	return nil
}

func Backup(input, target string) ([]byte, error) {
	err := clean(&input)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(RDIFF_BACKUP, input, target)
	buf, err := cmd.CombinedOutput()
	fmt.Println(RDIFF_BACKUP, cmd.Args)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func Restore(file, destination, t string) ([]byte, error) {
	cmd := exec.Command(RDIFF_BACKUP, RESTORE_FLAG, t, file, destination)
	fmt.Println(cmd.Args)
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func SnapshotTimes(target string) ([]time.Time, error) {
	result := make([]time.Time, 0)
	cmd := exec.Command(RDIFF_BACKUP, PARSABLE, "-l", filepath.Clean(target))
	b, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve rdiff-backup output", err)
	}

	buf := bytes.NewBuffer(b)
	break_ := true
	lines := make([]string, 0)
	for n := 1; break_; n++ {
		line, err := buf.ReadString('\n')
		if err != nil {
			break_ = false
		}
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}

	for index, line := range lines {
		if ((index+1)%9 != 0) && (index != len(lines)-1) {
			continue
		}

		line = strings.TrimRight(line, " directory\n")
		i, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to convert", line, err)
		}
		result = append(result, time.Unix(i, 0))
	}
	return result, nil
}

func Permissions(target string) ([]byte, error) {
	cmd := exec.Command(SETFACL, "-m", "g:"+ACLGROUP+":rwx", target)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("unable to set permissions", err)
	}

	return out, nil
}
