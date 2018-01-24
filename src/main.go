package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/fsnotify.v1"
)

// YYYY_MM_DD_HH
const layout = "2006_01_02_15"

type Config struct {
	TargetFolders []string
}

/*
   指定されたフォルダ内のファイルを監視する
   監視しているファイルが新しくなっていたら、手元へとコピーする。
*/

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func CopyAllInto(locatePath string) {
	fmt.Println("locatePath:", locatePath)

	// タイムスタンプからフォルダパスを生成
	formatted := time.Now().Format(layout)
	locateTargetPath := strings.Replace(locatePath, "/", "_", -1)

	// 新規フォルダを生成
	newRecordPath := filepath.Join("records", locateTargetPath, formatted)
	os.Mkdir(newRecordPath, 0700)

	// copyする。
	CopyRecursive(locatePath, newRecordPath)
}

func CopyRecursive(from, to string) {
	fmt.Println("from:", from, "to:", to)
	_, err := os.Stat(to)
	if err != nil {
		os.MkdirAll(to, 0700)
	}

	paths, _ := ioutil.ReadDir(from)

	for _, path := range paths {
		if path.IsDir() {
			folderName := filepath.Base(path.Name())
			CopyRecursive(filepath.Join(from, path.Name()), filepath.Join(to, folderName))
			continue
		}

		// path is file path. copy from to.
		fileName := filepath.Base(path.Name())
		fullPath := filepath.Join(from, path.Name())
		CopyFile(fullPath, filepath.Join(to, fileName))
	}
}

func CopyFile(from, to string) {
	in, err := os.Open(from)
	if err != nil {
		fmt.Println("open", err)
		return
	}
	defer in.Close()

	out, err := os.Create(to)
	if err != nil {
		fmt.Println("out", err)
		return
	}

	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	_, err = io.Copy(out, in)

	if err != nil {
		fmt.Println("copy", err)
		return
	}

	err = out.Sync()
}

func main() {
	// 手元にrecordフォルダを生成する

	// settings.txtを作ればいいか。あれば、そこから設定を読み込む。
	_, err := os.Stat("settings.toml")
	if err != nil {
		// create file anyway.
		_, err := os.Create("settings.toml")
		CheckError(err)
	}

	settings, err := ioutil.ReadFile("settings.toml")

	// settings file is generated or exists.
	var conf Config
	if _, err := toml.Decode(string(settings), &conf); err != nil {
		// handle error
		CheckError(err)
	}

	var targetFullPaths []string

	// 全てのpathに対して、フルパスを取得する。
	for i, path := range conf.TargetFolders {
		targetCandidatePath, _ := filepath.Abs(path)
		fmt.Println("targetCandidatePath:", targetCandidatePath, i)

		_, err := os.Stat(targetCandidatePath)
		if err != nil {
			// ignore if error exists.
			continue
		}

		// 有無をチェック、あれば追加。
		targetFullPaths = append(targetFullPaths, targetCandidatePath)
	}

	// logging.
	logPath := "log.txt"

	// log
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	CheckError(err)

	// ログをstdoutとファイル両方に書き出す(最終的にdaemonとかにしたいね。)
	multiWrite := io.MultiWriter(file, os.Stdout)
	logger := log.New(multiWrite, "tsuke:", log.Ldate|log.Ltime|log.Lshortfile)

	watcher, err := fsnotify.NewWatcher()
	CheckError(err)

	defer watcher.Close()

	/*
		ファイルイベント発生時に動かす関数。
		dirが入ってきた場合、監視対象に追加する必要がある。
	*/
	checkUpdate := func(newOrUpdatedFilePath string) {
		// .始まりのファイルなら拒否
		if strings.HasPrefix(filepath.Base(newOrUpdatedFilePath), ".") {
			return
		}

		absPath, _ := filepath.Abs(newOrUpdatedFilePath)
		// targetFullPathsに含まれているかどうかをチェックして、含まれていれば監視対象なので、フォルダを確認、保存する。
		for _, targetPath := range targetFullPaths {

			if !strings.HasPrefix(absPath, targetPath) {
				// not target.
				continue
			}

			// path is under target path.

			locatePath := strings.Replace(targetPath, "/", "_", -1)

			recordPath := filepath.Join("records", locatePath)

			_, err := os.Stat(recordPath)

			// 対象のtargetフォルダが見つからなかった=新規保存
			if err != nil {
				// create.
				os.MkdirAll(filepath.Join("records", locatePath), 0700)

				// 既存のファイルが存在するフォルダの中身を丸っとコピー
				CopyAllInto(targetPath)
				continue
			}

			// records/アイテムのフォルダは生成されているので、あとは日付フォルダの一覧を取得して、時間変換したら規定時間が過ぎている場合、コピー
			recordInfos, _ := ioutil.ReadDir(recordPath)
			var compareTargets []string
			for _, recordInfo := range recordInfos {
				if !recordInfo.IsDir() {
					continue
				}
				compareTargets = append(compareTargets, recordInfo.Name())
			}

			if len(compareTargets) <= 0 {
				continue
			}

			// targetパスの内部の要素なのが確定

			// このpathがディレクトリの場合、このディレクトリ自体も監視対象にいれないといけない。
			data, _ := os.Stat(absPath)
			if data.IsDir() {
				err := watcher.Add(absPath)
				if err != nil {
					fmt.Println("add err:", err)
				}
			} else { // 対象はファイルなのだけれど、

			}

			compareTargetPath := compareTargets[len(compareTargets)-1]
			compareTargetTimeSource := filepath.SplitList(compareTargetPath)
			compareTargetTime := compareTargetTimeSource[len(compareTargetTimeSource)-1]

			recordTime, _ := time.Parse(layout, compareTargetTime)

			current := time.Now().Format(layout)
			currentTime, _ := time.Parse(layout, current)

			// check distance of record.
			duration := currentTime.Sub(recordTime)

			// フォルダがコピーされた場合、この門を突破できない可能性がある。ふーむ、空のフォルダ作られても困るからな、、
			fmt.Println("incoming", absPath)
			if duration.Hours() == 0 {
				continue
			}

			CopyAllInto(targetPath)
		}
	}

	checkDelete := func(newOrUpdatedFilePath string) {
		absPath, _ := filepath.Abs(newOrUpdatedFilePath)

		data, _ := os.Stat(absPath)
		if data.IsDir() {
			watcher.Remove(absPath)
			if err != nil {
				fmt.Println("remove err:", err)
			}
		}
	}

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
					checkUpdate(event.Name)
					continue
				}

				if event.Op&fsnotify.Create == fsnotify.Remove {
					checkDelete(event.Name)
					continue
				}

				// logger.Println("other event:", event.Name, event.Op)

			case err := <-watcher.Errors:
				logger.Println("watcher err:", err)
			}
		}
	}()

	// 監視対象のフォルダをrecursiveに指定する。
	for _, targetFolderPath := range conf.TargetFolders {
		err2 := watcher.Add(targetFolderPath)
		CheckError(err2)
	}

	<-done
}

// ReadDir(dirname string) ([]os.FileInfo, error) フォルダの中のファイル一覧を取得する
// files, _ := ioutil.ReadDir("/tmp")
// for _, file := range files {
//     fmt.Println(file.Name())
// }
