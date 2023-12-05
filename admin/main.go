package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	_ "github.com/lengzhao/font/autoload"
)

func WriteString(configPath, str string) error {
	file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("无法打开文件:", err)
		return err
	}
	defer file.Close()

	_, err = file.WriteString(str)
	if err != nil {
		fmt.Println("写入文件失败:", err)
		return err
	}

	return nil
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("管理面板")
	myWindow.Resize(fyne.NewSize(500, 500))
	myWindow.CenterOnScreen()

	dir := ""

	pathShow := widget.NewLabel("你还没有选择文件夹")

	openBtn := widget.NewButton("选择文件夹", func() {
		selectFolderPath := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if uri != nil {
				dir = filepath.ToSlash(uri.Path())
				pathShow.SetText("你选择的文件夹是：" + dir)
			}

		}, myWindow)
		selectFolderPath.SetConfirmText("确定")
		selectFolderPath.SetDismissText("取消")
		selectFolderPath.Show()

	})

	saveBtn := widget.NewButton("生成配置文件", func() {

		if dir == "" {
			dialog.ShowError(fmt.Errorf("请选择文件夹"), myWindow)
			return
		}
		clientDir := dir
		configPath := dir + "/patcher/config.ini"

		_, err := os.Stat(clientDir + "/patcher")
		if os.IsNotExist(err) {
			err := os.MkdirAll(clientDir+"/patcher", 0755)
			if err != nil {
				fmt.Println("无法创建文件夹:", err)
				return
			}
		}
		os.Remove(configPath)

		filepath.Walk(clientDir, func(path string, info os.FileInfo, err error) error {

			if err != nil {
				fmt.Println("访问文件时出错:", err)
				return nil
			}

			if info.IsDir() {
				return nil
			}

			fileName, err := filepath.Rel(dir, filepath.ToSlash(path))
			if err != nil {
				panic(err)
			}

			fileName = filepath.ToSlash(fileName)
			if fileName == "patcher/config.ini" {
				return nil
			}

			content, _ := os.ReadFile(path)
			hash := md5.Sum([]byte(content))
			hashString := hex.EncodeToString(hash[:])
			line := fmt.Sprintf("%s\t%s\t%d\n", fileName, hashString, info.Size())

			WriteString(configPath, line)
			return nil
		})
		dialog.ShowInformation("提示", "生成配置文件成功", myWindow)
	})

	content := container.NewVBox(pathShow, openBtn, saveBtn)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
