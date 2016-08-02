// main
package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	SHOOTERURL = "https://shooter.cn/api/subapi.php"
)

type logmgr struct {
	debug    bool
	debugLog *log.Logger
	infoLog  *log.Logger
}

func (p *logmgr) Debugln(args ...interface{}) {
	if p.debug {
		p.debugLog.Println(args...)
	}
}
func (p *logmgr) Debugf(format string, args ...interface{}) {
	if p.debug {
		p.debugLog.Printf(format, args...)
	}
}

func (p *logmgr) Infoln(args ...interface{}) {
	p.infoLog.Println(args...)
}
func (p *logmgr) Infof(format string, args ...interface{}) {
	p.infoLog.Printf(format, args...)
}
func (p *logmgr) SetDebug(debug bool) {
	p.debug = debug
}

var gLog = &logmgr{false, log.New(os.Stderr, "[DEBUG] ", 0), log.New(os.Stderr, "[INFO] ", 0)}

var scanFinish = errors.New("finish by interrupt")

func initExtMap() map[string]int {
	extmap := map[string]int{}
	extmap[".avi"] = 1
	extmap[".mp4"] = 1
	extmap[".mkv"] = 1
	extmap[".rm"] = 1
	extmap[".rmvb"] = 1
	return extmap
}

func getfilehash(fullpath string) string {
	fp, err := os.Open(fullpath)
	if err != nil {
		gLog.Infof("open file %v err: %v\n", fullpath, err)
		return ""
	}
	defer fp.Close()
	stats, statsErr := fp.Stat()
	if statsErr != nil {
		gLog.Infof("stat file %v err: %v\n", fullpath, err)
		return ""
	}
	filelen := stats.Size()
	offsetary := [...]int64{4096, (filelen / 3) * 2, filelen / 3, filelen - 8192}
	buf := make([]byte, 4096)
	hashary := make([]string, 0, len(offsetary))
	for _, offset := range offsetary {
		fp.Seek(offset, 0)
		n, _ := io.ReadFull(fp, buf)
		hashary = append(hashary, fmt.Sprintf("%x", md5.Sum(buf[:n])))
	}
	return strings.Join(hashary, ";")
}

func fileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func downloadsub(fullpath string, language string) {
	gLog.Infoln("handle name:", fullpath)
	hash := getfilehash(fullpath)
	u, _ := url.Parse(SHOOTERURL)
	param := u.Query()
	param.Add("filehash", hash)
	_, filename := filepath.Split(fullpath)
	param.Add("pathinfo", filename)
	param.Add("format", "json")
	param.Add("lang", language)
	u.RawQuery = param.Encode()

	fullurl := u.String()
	gLog.Debugln("org url:", fullurl)

	response, err := http.Get(fullurl)
	if err != nil {
		gLog.Infof("http get[%v] err:%v\n", fullurl, err)
		return
	}
	defer response.Body.Close()
	result, err := ioutil.ReadAll(response.Body)
	gLog.Debugf("url return: %s\n", result)

	subInfo, err := getSubInfo(result)
	if err != nil {
		gLog.Infof("parse json error:%v\n", err)
	}
	basename := fullpath[:strings.LastIndex(fullpath, ".")]
	for i, _ := range subInfo {
		for _, info := range subInfo[i].Files {
			var subfile string
			if i == 0 {
				subfile = fmt.Sprintf("%s.%s.%s", basename, "chn", info.Ext)
			} else {
				subfile = fmt.Sprintf("%s.%s%d.%s", basename, "chn", i, info.Ext)
			}
			if fileExist(subfile) {
				gLog.Infof("subfile %v exist,skip\n", subfile)
			} else {
				fetchsubdata(subfile, info.Link, subInfo[i].Delay)
			}
		}
	}
}

type FileInfo struct {
	Ext  string `json:"Ext"`
	Link string `json:"Link"`
}

type SubInfo struct {
	Desc  string     `json:"Desc"`
	Delay int        `json:Delay`
	Files []FileInfo `json:Files`
}

func getSubInfo(data []byte) ([]SubInfo, error) {
	var result []SubInfo
	err := json.Unmarshal(data, &result)
	return result, err
}

func fetchsubdata(path string, url string, delay int) error {
	begin := time.Now()
	tmppath := path + ".tmp"
	response, err := http.Get(url)
	if err != nil {
		gLog.Infof("http get[%v] err:%v\n", url, err)
		return err
	}
	defer response.Body.Close()

	if delay != 0 {
		delayPath := path + ".delay"
		tmpfile, err := os.OpenFile(delayPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			gLog.Infof("open err:%v\n", err)
			return err
		} else {
			fmt.Fprintf(tmpfile, "%d\n", delay)
			tmpfile.Close()
			gLog.Infof("write file %v ok,delay:%v\n", delayPath, delay)
		}
	}
	tmpfile, err := os.OpenFile(tmppath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		gLog.Infof("open err:%v\n", err)
		return err
	}

	_, err = io.Copy(tmpfile, response.Body)
	if err != nil {
		gLog.Infof("copy data to file %v err:%v\n", tmppath, err)
		tmpfile.Close()
		os.Remove(tmppath)
		return err
	}
	tmpfile.Close()
	err = os.Rename(tmppath, path)
	spend := time.Now().Sub(begin)
	if err != nil {
		gLog.Infof("rename file error:%v,spend:%v\n", err, spend)
		return err
	} else {
		gLog.Infof("download file %v ok,spend:%v\n", path, spend)
	}
	return nil
}

func getwalkfunc(root string, skipSubDir bool, language string, extfiltermap map[string]int, checkcnt *int) filepath.WalkFunc {
	*checkcnt = 0
	return func(path string, info os.FileInfo, err error) error {
		if false && *checkcnt >= 1 {
			return scanFinish
		}
		if err != nil {
			gLog.Infof("stat file %v err: %v\n", path, err)
			return nil
		}
		if skipSubDir && info.IsDir() && path != root {
			gLog.Infof("skip dir %v\n", path)
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		extName := filepath.Ext(path)
		_, ok := extfiltermap[strings.ToLower(extName)]
		if !ok {
			return nil
		}
		downloadsub(path, language)
		(*checkcnt)++
		<-time.After(time.Second)
		return nil
	}
}

func main() {
	var extList string
	var debug bool
	var help bool
	flag.BoolVar(&help, "h", false, "print help info")
	flag.BoolVar(&debug, "d", false, "need output debug info")
	flag.StringVar(&extList, "ext", "", "check ext file list,default include:avi/mp4/mkv/rm/rmvb")
	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [option] path\n", os.Args[0])
		flag.PrintDefaults()
	}

	if help {
		flag.Usage()
		os.Exit(0)
	}
	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	root := args[0]

	gLog.SetDebug(debug)
	extfiltermap := initExtMap()
	if extList != "" {
		for _, v := range strings.Split(extList, ",") {
			extfiltermap["."+v] = 1
		}
	}
	checkcnt := 0
	err := filepath.Walk(root, getwalkfunc(root, false, "chn", extfiltermap, &checkcnt))
	if err != nil && err != scanFinish {
		log.Fatalf("scan dir %v err: %v\n", root, err)
	}
	if checkcnt == 0 {
		gLog.Infoln("check file cnt:", checkcnt)
	}
}
