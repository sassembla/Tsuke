package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"gopkg.in/fsnotify.v1"
)

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

	// os.Stat()
}

func main() {
	// 特定のフォルダのファイルを取得する
	targetFolderPath := "target"

	// 手元にrecordフォルダを生成する

	// settings.txtを作ればいいか。あれば、そこから設定を読み込む。
	settingsInfo, err := os.Stat("settings.txt")
	if err == os.ErrNotExist {
		// file, _ := os.Create("settings.txt")

		// file.
	} else if err != nil {
		return
	}

	fmt.Println("settings:", settingsInfo.)

	// settings file is generated or exists.

	logPath := "log.txt"

	// log
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	CheckError(err)

	// stdoutとファイル両方に書き出す
	multiWrite := io.MultiWriter(file, os.Stdout)
	logger := log.New(multiWrite, "tsuke:", log.Ldate|log.Ltime|log.Lshortfile)

	logger.Println("eventing0.")

	// // ファイルパスに対して中身を取得する。うーん、変更検知のターゲットとして
	// filepath.Walk(
	// 	targetFolderPath,
	// 	func(path string, info os.FileInfo, err error) error {
	// 		if err != nil {
	// 			logger.Println("failed to traverse.", err)
	// 			return nil
	// 		}

	// 		// ディレクトリは無視する？それともずっと追うか？ 副産物とかもあるから追いそう。
	// 		if info.IsDir() {
	// 			return nil
	// 		}

	// 		// // 指定されている場合拡張子を見る？まあいいや。
	// 		// ext := strings.ToLower(filepath.Ext(info.Name()))
	// 		logger.Println("file:", info.Name())
	// 		return nil
	// 	})

	watcher, err := fsnotify.NewWatcher()
	CheckError(err)

	defer watcher.Close()
	logger.Println("eventing1.")

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
	err2 := watcher.Add(targetFolderPath)

	CheckError(err2)
	<-done

	logger.Println("eventing2.")
}
