package web

import (
	"encoding/json"
	"fmt"
	"github.com/jostillmanns/rdiff-backup-tools/incs"
	"github.com/jostillmanns/rdiff-backup-tools/types"
	"github.com/jostillmanns/rdiff-backup-tools/utils"
	"github.com/jostillmanns/rdiff-backup-tools/wrapper"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type element struct {
	Name  string `json:"name"`
	Ext   string `json:"ext"`
	Size  string `json:"size"`
	Date  string `json:"date"`
	IsDir bool   `json:"isDir"`
	Path  string `json:"path"`
}

var repos []utils.Repository
var permissions_ bool

func phyElements(w http.ResponseWriter, r *http.Request) {
	dest := ""
	if r.FormValue("dest") == "" && len(utils.SplitPath(html.UnescapeString(r.FormValue("path")))) > 0 {
		// go up one directory
		dest = filepath.Dir(html.UnescapeString(r.FormValue("path")))
	} else {
		dest = filepath.Join(html.UnescapeString(r.FormValue("path")), html.UnescapeString(r.FormValue("dest")))
	}

	level := len(utils.SplitPath(dest))
	var data []byte
	var elements []element

	if level == 0 {
		elements = make([]element, len(repos))
		for i, e := range repos {
			elements[i] = element{e.Name, "DIR", "0", "0", true, dest}
		}
	} else if level > 0 {
		fp := filepath.Join("repositories", utils.SplitPath(dest)[0], "origin")
		if level > 1 {
			for _, e := range utils.SplitPath(dest)[1:] {
				fp = filepath.Join(fp, e)
			}
		}
		infos, err := ioutil.ReadDir(fp)
		if err != nil && !os.IsNotExist(err) {
			log.Panic(err)
		}
		if err == nil {
			elements = make([]element, len(infos))
			for i, e := range infos {
				ext := filepath.Ext(e.Name())
				if e.IsDir() {
					ext = "DIR"
				}
				elements[i] = element{e.Name(), ext, strconv.FormatInt(e.Size(), 10), e.ModTime().Format(types.TIMESTAMP_FMT), e.IsDir(), dest}
			}
		}
	}

	if len(elements) == 0 {
		elements = []element{element{"", "DUMMY0", "0", "0", false, dest}}
	}

	data, err := json.Marshal(elements)
	if err != nil {
		log.Panic(err)
	}

	w.Write(data)
}

func rdiffElements(w http.ResponseWriter, r *http.Request) {
	dest := ""
	if r.FormValue("dest") == "" && len(utils.SplitPath(html.UnescapeString(r.FormValue("path")))) > 0 {
		// we want to go up one directory
		dest = filepath.Dir(html.UnescapeString(r.FormValue("path")))
	} else {
		dest = filepath.Join(html.UnescapeString(r.FormValue("path")), html.UnescapeString(r.FormValue("dest")))
	}

	level := len(utils.SplitPath(dest))
	var data []byte
	var elements []element

	if level == 0 {
		elements = make([]element, len(repos))
		for i, e := range repos {
			elements[i] = element{e.Name, "DIR", "0", "0", true, dest}
		}
	}
	if level == 1 {
		name := utils.SplitPath(dest)[0]
		path := filepath.Join("repositories", name)
		repo := utils.Repository{
			filepath.Join(path, "basepath"),
			filepath.Join(path, "directorystructure"),
			filepath.Join(path, "origin"),
			name,
		}

		times, err := repo.TimeStamps()
		if err != nil {
			log.Panic(err)
		}

		elements = make([]element, len(times))
		for i, e := range times {
			u, err := utils.Unquote(e)
			if err != nil {
				log.Panic(err)
			}
			elements[i] = element{u, "DIR", "0", "0", true, dest}
		}
	}
	if level > 1 {
		name := utils.SplitPath(dest)[0]
		path := filepath.Join("repositories", name)
		repo := utils.Repository{
			filepath.Join(path, "basepath"),
			filepath.Join(path, "directorystructure"),
			filepath.Join(path, "origin"),
			name,
		}

		var p string
		if len(utils.SplitPath(dest)) == 2 {
			p = ""
		} else {
			for _, e := range utils.SplitPath(dest)[2:] {
				p = filepath.Join(p, e)
			}
		}
		time_, err := utils.Unquote(utils.SplitPath(dest)[1])
		if err != nil {
			log.Panic(err)
		}
		timepoint, err := time.Parse(types.TIMESTAMP_FMT, time_)
		if err != nil {
			log.Panic(err)
		}
		dirs, err := incs.Directories(&repo, p, timepoint)
		if err != nil {
			log.Panic(err)
		}
		files, err := incs.Files(&repo, p, timepoint)
		if err != nil {
			log.Panic(err)
		}
		elements = make([]element, len(dirs)+len(files))
		for i, e := range dirs {
			elements[i] = element{e, "DIR", "0", "0", true, dest}
		}
		for i, e := range files {
			elements[i+len(dirs)] = element{e, filepath.Ext(e), "0", "0", false, dest}
		}

		if len(elements) == 0 {
			elements = []element{element{"", "DUMMY0", "0", "0", false, dest}}
		}
	}

	data, err := json.Marshal(elements)
	if err != nil {
		log.Panic("unable to marshal table rows", err)
	}
	w.Write(data)
}

func restore(w http.ResponseWriter, r *http.Request) {
	source := utils.SplitPath(html.UnescapeString(r.FormValue("source")))
	element := html.UnescapeString(r.FormValue("element"))
	target := utils.SplitPath(html.UnescapeString(r.FormValue("target")))

	time := source[1]
	sourcePath, err := filepath.EvalSymlinks(filepath.Join("repositories", source[0], "basepath"))
	if err != nil {
		log.Panic("Sourcepath: unable to find EvalSymlinks Path", err)
	}
	for _, e := range source[2:] {
		sourcePath = filepath.Join(sourcePath, e)
	}
	sourcePath = filepath.Join(sourcePath, element)

	targetPath, err := filepath.EvalSymlinks(filepath.Join("repositories", target[0], "origin"))
	if err != nil {
		log.Panic("Targetpath: unable to find EvalSymlinks Path", err)
	}
	for _, e := range target[1:] {
		targetPath = filepath.Join(targetPath, e)
	}
	targetPath = filepath.Join(targetPath, element)

	fmt.Println(sourcePath, targetPath, time)
	output, error := wrapper.Restore(sourcePath, targetPath, time)
	if error != nil {
		log.Panic("unable to restore", sourcePath, target, time, error)
	}
	log.Println(string(output))

	if permissions_ {
		output, err = wrapper.Permissions(targetPath)
		if err != nil {
			log.Panic("while restoring:", err)
		}
		log.Println(string(output))
	}
}

func initLog() *os.File {
	f, err := os.OpenFile(filepath.Join("web", "view", "web.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("unable to open log file")
	}

	log.SetOutput(f)
	return f
}

func StartServer(permissions bool) {
	permissions_ = permissions
	f := initLog()

	var err error
	repos, err = utils.InitRepositories()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/rdiff-elements", rdiffElements)
	http.HandleFunc("/phy-elements", phyElements)
	http.HandleFunc("/restore", restore)
	http.Handle("/", http.FileServer(http.Dir(filepath.Clean("web/view"))))
	err = http.ListenAndServe(":4000", nil)
	if err != nil {
		log.Fatal(err)
	}

	(*f).Close()
}
