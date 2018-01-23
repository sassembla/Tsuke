package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"gopkg.in/fsnotify.v1"
)

type Config struct {
	TargetFolders []string
}

/*
   指定されたフォルダ内のファイルを監視する
   監視しているファイルが新しくなっていたら、手元へとコピーする。
*/

// CheckError checks for errors
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func checkUpdate(fileName string) {
	fmt.Println("テスト、ファイル名はフルパスが来る。", fileName)
	// このファイル名に対して、ファイルの直上のフォルダに関してのrecordpathがなければ作る。
	// うーん、監視フォルダ名を持っておいて、それを上部に含んでいたら、とかの方がいいか。
	targetPath := "target"
	fmt.Println("targetPath:", targetPath)

	// os.Stat()
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

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
					checkUpdate(event.Name)
					continue
				}

				// logger.Println("other event:", event.Name, event.Op)

			case err := <-watcher.Errors:
				logger.Println("watcher err:", err)
			}
		}
	}()

	// 監視対象のフォルダを指定する。
	for _, targetFolderPath := range conf.TargetFolders {
		err2 := watcher.Add(targetFolderPath)
		CheckError(err2)
	}

	<-done

	logger.Println("eventing2.")
}

// ReadDir(dirname string) ([]os.FileInfo, error) フォルダの中のファイル一覧を取得する
// files, _ := ioutil.ReadDir("/tmp")
// for _, file := range files {
//     fmt.Println(file.Name())
// }
