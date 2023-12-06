package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/getlantern/elevate"
	_ "github.com/lengzhao/font/autoload"
)

const PATCH_URL = "http://patch.mt2.com/"
const PATCH_CONFIG_URL = PATCH_URL + "patcher/config.ini"
const WEB_URL = "http://localhost:8080/"
const REGISTER_WEB_URL = WEB_URL + "register"
const NEWS_WEB_URL = WEB_URL + "news/list"

type News struct {
	Title string
	Time  string
	Link  string
}

type FileInfo struct {
	Name string
	Md5  string
	Size int
}

func getData() ([]FileInfo, int, error) {
	// 发送HTTP GET请求
	response, err := http.Get(PATCH_CONFIG_URL)
	if err != nil {
		return nil, 0, err
	}
	defer response.Body.Close()

	// 检查响应状态码
	if response.StatusCode != http.StatusOK {
		fmt.Println("请求返回错误:", response.StatusCode)
		return nil, 0, fmt.Errorf("请求返回错误：%d", response.StatusCode)
	}

	// 读取响应体的内容
	var data []FileInfo
	totalSize := 0

	body := response.Body
	scanner := bufio.NewScanner(body)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, "\t")
		if len(fields) != 3 {
			return nil, 0, fmt.Errorf("无效的行：%s", line)
		}

		// 解析行内容
		name := fields[0]
		fileMd5 := fields[1]
		size, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, 0, fmt.Errorf("size转换失败：%s", line)
		}

		// 文件不存在则添加到列表
		if _, err := os.Stat(name); os.IsNotExist(err) {
			obj := FileInfo{
				Name: name,
				Md5:  fileMd5,
				Size: size,
			}

			data = append(data, obj)
			totalSize += size
		} else {
			// 读取文件内容并计算md5值
			content, err := os.ReadFile(name)
			if err != nil {
				return nil, 0, fmt.Errorf("读取文件内容时发生错误：%s", err)
			}
			hash := md5.Sum([]byte(content))
			hashString := hex.EncodeToString(hash[:])

			// 判断md5值是否一致，不一致则添加到列表
			if hashString != fileMd5 {
				obj := FileInfo{
					Name: name,
					Md5:  fileMd5,
					Size: size,
				}

				data = append(data, obj)
				totalSize += size
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, fmt.Errorf("读取文件内容时发生错误: %s", err)
	}

	return data, totalSize, nil
}

func downloadFile(name string) error {
	// 发送HTTP GET请求
	fileUrl := PATCH_URL + name
	response, err := http.Get(fileUrl)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// 检查响应状态码
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("请求返回错误：%d", response.StatusCode)
	}

	dir := filepath.Dir(name)

	if _, err = os.Stat(dir); os.IsNotExist(err) {
		// 创建目录
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)

	if err != nil {
		return err
	}

	return nil
}

func makeToolbarTab(_ fyne.Window, homeFunc func(), accountFunc func(), configFunc func()) fyne.CanvasObject {
	t := widget.NewToolbar(
		widget.NewToolbarAction(theme.HomeIcon(), homeFunc),
		widget.NewToolbarAction(theme.AccountIcon(), accountFunc),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.SettingsIcon(), configFunc),
	)

	return container.NewBorder(t, nil, nil, nil)
}

func getNewsList() ([]News, error) {

	response, err := http.Get(NEWS_WEB_URL)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var newsList []News
	err = json.Unmarshal(body, &newsList)

	if err != nil {
		return nil, err
	}

	return newsList, nil
}

func main() {

	// 管理员打开
	if len(os.Args) < 2 || os.Args[1] != "--escalate" {
		cmd := elevate.Command(os.Args[0], "--escalate")
		cmd.Run()
		os.Exit(0)
	}

	myApp := app.New()
	myWindow := myApp.NewWindow("更新程序")
	// myWindow.Resize(fyne.NewSize(550, 0))
	myWindow.CenterOnScreen()
	// myWindow.SetFixedSize(true)
	// myWindow.SetPadded(true)

	// 状态栏
	status := widget.NewLabel("")

	// 顶部菜单栏
	homeFunc := func() {
		parsedURL, err := url.Parse(WEB_URL)
		if err != nil {
			status.SetText(fmt.Sprintf("%s", err))
			return
		}
		app.New().OpenURL(parsedURL)
	}
	accountFunc := func() {
		parsedURL, err := url.Parse(REGISTER_WEB_URL)
		if err != nil {
			status.SetText(fmt.Sprintf("%s", err))
			return
		}
		app.New().OpenURL(parsedURL)
	}
	configFunc := func() {
		cmd := exec.Command("./config.exe")
		err := cmd.Start() // 启动程序，不等待它完成
		if err != nil {
			status.SetText(fmt.Sprintf("%s", err))
		}
	}
	toolbarContainer := makeToolbarTab(myWindow, homeFunc, accountFunc, configFunc)

	// 新闻
	newsList, err := getNewsList()
	if err != nil {
		fyne.LogError("GET NEWS LIST:", err)
		return
	}

	newsBar := container.NewVBox()
	for _, news := range newsList {
		link, err := url.Parse(WEB_URL + news.Link)
		if err != nil {
			fyne.LogError("Could not parse URL", err)
			return
		}
		hyperlink := widget.NewHyperlink(news.Title + "     " + news.Time, link)
		newsBar.Add(hyperlink)
	}

	newsCard := widget.NewCard("", "欢迎来到长安之角，新角色，新地图，新玩法。", newsBar)

	// 进度条
	progress := widget.NewProgressBar()
	progress.SetValue(0)

	// 启动按钮
	startBtn := &widget.Button{
		Text:       "启动",
		Importance: widget.SuccessImportance,
		OnTapped: func() {
			cmd := exec.Command("./Metin2Client.bin")
			err := cmd.Start()

			if err != nil {
				status.SetText(fmt.Sprintf("%s", err))
			} else {
				myApp.Quit()
			}
		},
	}

	startBtn.Disable()

	// 布局
	myWindow.SetContent(container.NewVBox(toolbarContainer, newsCard, layout.NewSpacer(), progress, status, startBtn))

	go func() {
		status.SetText("正在初始化...")
		data, totalSize, err := getData()

		if err != nil {
			status.SetText(fmt.Sprintf("%s", err))
			return
		}

		if totalSize == 0 {
			status.SetText("更新完成")
			progress.SetValue(1)
			startBtn.Enable()
			return
		}

		downloadSize := 0
		for _, value := range data {
			status.SetText("正在下载文件 " + value.Name)
			err = downloadFile(value.Name)
			if err != nil {
				status.SetText(fmt.Sprintf("%s", err))
				return
			}
			downloadSize += value.Size
			progress.SetValue(float64(downloadSize) / float64(totalSize))
		}

		if downloadSize == totalSize {
			status.SetText("更新完成")
			progress.SetValue(1)
			startBtn.Enable()
		}
	}()

	myWindow.ShowAndRun()
}
