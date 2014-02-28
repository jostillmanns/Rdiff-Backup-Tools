package testdata

import (
	"fmt"
	"github.com/jostillmanns/rdiff-backup-tools/utils"
	"github.com/jostillmanns/rdiff-backup-tools/wrapper"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func GenerateTestRepo(repo utils.Repository) ([]byte, error) {
	output, err := Init(repo)
	if err != nil {
		return output, fmt.Errorf("Unable to init repository", err)
	}

	for _, d := range []int{0, 1} {
		time.Sleep(time.Second)
		out, err := AddDirectory(repo, filepath.Join("dir"+strconv.Itoa(d)))
		if err != nil {
			return out, fmt.Errorf("unable to create dir"+strconv.Itoa(d), err)
		}

		for _, f := range []int{0, 1, 2, 3, 4} {
			time.Sleep(time.Second)
			out, err := AddFile(repo, filepath.Join("dir"+strconv.Itoa(d), strconv.Itoa(f)))
			if err != nil {
				return out, fmt.Errorf("unable to create file"+strconv.Itoa(f), "directory dir"+strconv.Itoa(f), err)
			}
		}
	}
	return nil, nil
}

func Init(repo utils.Repository) ([]byte, error) {
	i, _ := os.Stat(filepath.Clean("."))
	err := os.Mkdir(repo.Origin, i.Mode())
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	output, err := wrapper.Backup(filepath.Clean(repo.Origin), filepath.Clean(repo.BasePath))
	if err != nil {
		return nil, fmt.Errorf("unable to backup", err)
	}

	if repo.DirectoryStructure != "" {
		err = os.Mkdir(repo.DirectoryStructure, i.Mode())
		if err != nil {
			return nil, fmt.Errorf("unable to create directory for directory structure", nil)
		}
	}

	return output, nil
}

func Clean(repo utils.Repository) error {
	err := os.RemoveAll(filepath.Clean(repo.BasePath))
	if err != nil {
		return fmt.Errorf("unable to delete repo basepath", err)
	}
	err = os.RemoveAll(filepath.Clean(repo.Origin))
	if err != nil {
		return fmt.Errorf("unable to delete repo origin", err)
	}
	if repo.DirectoryStructure != "" {
		err = os.RemoveAll(repo.DirectoryStructure)
		if err != nil {
			return fmt.Errorf("unable to remove directory structure", nil)
		}
	}
	return nil
}

func AddFile(repo utils.Repository, file string) ([]byte, error) {
	_, err := os.Create(filepath.Join(repo.Origin, file))
	if err != nil {
		return nil, fmt.Errorf("unable to create file", err)
	}
	return wrapper.Backup(filepath.Clean(repo.Origin), filepath.Clean(repo.BasePath))
}

func AddDirectory(repo utils.Repository, dir string) ([]byte, error) {
	i, err := os.Stat(filepath.Clean(repo.BasePath))
	if err != nil {
		return nil, fmt.Errorf("unable to stat working directory", err)
	}
	err = os.Mkdir(filepath.Join(repo.Origin, dir), i.Mode())
	if err != nil {
		return nil, fmt.Errorf("unable to make requested directory", err)
	}
	return wrapper.Backup(filepath.Clean(repo.Origin), filepath.Clean(repo.BasePath))
}

func RemoveFile(repo utils.Repository, file string) ([]byte, error) {
	err := os.Remove(filepath.Join(repo.Origin, file))
	if err != nil {
		return nil, fmt.Errorf("unable to delete file", err)
	}
	return wrapper.Backup(filepath.Clean(repo.Origin), filepath.Clean(repo.BasePath))

}

func RemoveDirectory(repo utils.Repository, dir string) ([]byte, error) {
	err := os.RemoveAll(filepath.Join(repo.Origin, dir))
	if err != nil {
		return nil, fmt.Errorf("unable to remove requested directory", err)
	}
	return wrapper.Backup(filepath.Clean(repo.Origin), filepath.Clean(repo.BasePath))
}
