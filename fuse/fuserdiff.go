package fuse

import (
	"bitbucket.org/jostillmanns/rdiff-backup-tools/incs"
	"fmt"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"os"
	"path/filepath"
	"time"
)

var (
	increments          = filepath.Join("./rdiff", "target")
	directory_structure = filepath.Join("./rdiff", "directory_structure")
	repo                = incs.Repository{increments, directory_structure}
)

type RdiffFs struct {
	pathfs.FileSystem
}

func basedir(path string) string {
	for filepath.Dir(path) != "." {
		path = filepath.Dir(path)
	}

	return path
}

func (me *RdiffFs) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {

	// fmt.Println("GetAttr", name)

	if basedir(name) == name {
		return &fuse.Attr{Mode: fuse.S_IFDIR | 0755}, fuse.OK
	}

	for _, e := range []string{filepath.Join(repo.BasePath, incs.DATA, incs.INCREMENTS, name[len(basedir(name)):]), filepath.Join(repo.BasePath, name[len(basedir(name)):])} {
		info, err := os.Stat(e)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			panic(err)
		}

		// found ya
		if info.IsDir() {
			return &fuse.Attr{Mode: fuse.S_IFDIR | 0755}, fuse.OK
		} else {
			return &fuse.Attr{Mode: fuse.S_IFREG | 0755}, fuse.OK
		}
	}
	return &fuse.Attr{Mode: fuse.S_IFREG | 0755}, fuse.OK
}

func (me *RdiffFs) OpenDir(name string, context *fuse.Context) (c []fuse.DirEntry, code fuse.Status) {

	if name == "" {
		timestamps, err := repo.TimeStamps()
		if err != nil {
			fmt.Println(err)
			return nil, fuse.ENOENT
		}

		c = make([]fuse.DirEntry, len(timestamps))
		for i, e := range timestamps {
			c[i] = fuse.DirEntry{Name: e, Mode: fuse.S_IFREG}

		}
		return c, fuse.OK
	}

	timestamp, err := incs.Unquote(basedir(name))
	if err != nil {
		fmt.Println(err)
		return nil, fuse.ENOENT
	}
	timepoint, err := time.Parse(incs.TIMESTAMP_FMT, timestamp)
	if err != nil {
		fmt.Println(err)
		return nil, fuse.ENOENT
	}

	path := name[len(basedir(name)):]
	directories, err := repo.Directories(path, timepoint)
	if err != nil {
		panic(err)
	}
	files, err := repo.Files(path, timepoint)
	if err != nil {
		panic(err)
	}
	c = make([]fuse.DirEntry, len(directories)+len(files))
	for i, e := range directories {
		c[i] = fuse.DirEntry{Name: e, Mode: fuse.S_IFDIR}
	}
	for i, e := range files {
		c[i+len(directories)] = fuse.DirEntry{Name: e, Mode: fuse.S_IFREG}
	}

	return c, fuse.OK
}

func (me *RdiffFs) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {

	// fmt.Println("Open", name)
	// if name != "file.txt" {
	// 	return nil, fuse.ENOENT
	// }
	// if flags&fuse.O_ANYWRITE != 0 {
	// 	return nil, fuse.EPERM
	// }
	return nodefs.NewDataFile([]byte(name)), fuse.OK
}

// func main() {
// 	flag.Parse()

// 	nfs := pathfs.NewPathNodeFs(&RdiffFs{FileSystem: pathfs.NewDefaultFileSystem()}, nil)
// 	server, _, err := nodefs.MountFileSystem(flag.Arg(0), nfs, nil)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	server.Serve()
// }
