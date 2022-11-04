package main

// 使用 https://github.com/hekmon/transmissionrpc
// SDK路径 C:\Users\zhaobi\AppData\Local\Android\Sdk
// go install fyne.io/fyne/v2/cmd/fyne@latest
// fyne package -os android -appID org.zhaobi.transmissionclient
// fyne package -os windows -icon myapp.png

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type DownloadType int

const (
	DOWNLOADTYPE_TV      DownloadType = 1
	DOWNLOADTYPE_MOVIE   DownloadType = 2
	DOWNLOADTYPE_ACG     DownloadType = 3
	DOWNLOADTYPE_ARTSHOW DownloadType = 4
	DOWNLOADTYPE_BT      DownloadType = 5
	DOWNLOADTYPE_UNKNOWN DownloadType = 99
)

func (tp DownloadType) GetDownloadDir() string {
	switch tp {
	case DOWNLOADTYPE_TV:
		return "/mnt/utdir/video/剧集/"
	case DOWNLOADTYPE_MOVIE:
		return "/mnt/utdir/video/电影/"
	case DOWNLOADTYPE_ACG:
		return "/mnt/utdir/video/动漫/"
	case DOWNLOADTYPE_ARTSHOW:
		return "/mnt/utdir/video/综艺/"
	case DOWNLOADTYPE_BT:
		return "/mnt/utdir/BT/"
	default:
	}
	return "/mnt/utdir/unknown/"
}

func (tp DownloadType) String() string {
	switch tp {
	case DOWNLOADTYPE_TV:
		return "剧集"
	case DOWNLOADTYPE_MOVIE:
		return "电影"
	case DOWNLOADTYPE_ACG:
		return "动漫"
	case DOWNLOADTYPE_ARTSHOW:
		return "综艺"
	case DOWNLOADTYPE_BT:
		return "BT"
	default:
	}
	return "unknown"
}

var g_mainTsClient transmissionClient

var g_win fyne.Window

var gui_dlListVBox *widget.List
var gui_titleLabel *widget.Label

func main() {
	g_app := app.NewWithID("org.zhaobi.transmissiondesktop")

	g_app.Settings().SetTheme(&myTheme{})
	//g_app.SetIcon(resourceIconPng)
	g_win = g_app.NewWindow("BSir's Transmission")

	g_win.SetContent(container.NewCenter(widget.NewLabel("Connecting Server ...\nPowered By ZhaoBi @2022"))) //临时首次显示
	g_win.Resize(fyne.NewSize(500, 400))
	g_win.Show()

	//fmt.Println(g_app.Storage().RootURI())
	//读取上次配置
	iIp := (g_app.Preferences().StringWithFallback("server", "129.28.167.223"))
	iPort := (g_app.Preferences().IntWithFallback("port", 39091))
	iUsername := (g_app.Preferences().StringWithFallback("username", "zhaobi"))
	iPassward := (g_app.Preferences().StringWithFallback("passward", "ZhaoBi@"))

	if iPort > 0 {
		go initTransmissionClient(iIp, iPort, iUsername, iPassward)
	} else {
		go initWelcomeContent()
	}

	g_win.ShowAndRun()

}

func initWelcomeContent() {
	serverip := widget.NewEntry()
	serverip.SetPlaceHolder("Server or IP address")

	port := widget.NewEntry()
	port.SetPlaceHolder("Port")

	username := widget.NewEntry()
	username.SetPlaceHolder("Port")

	passward := widget.NewEntry()
	passward.SetPlaceHolder("Passward")

	g_app := fyne.CurrentApp()
	//读取上次配置
	serverip.SetText(g_app.Preferences().StringWithFallback("server", "129.28.167.22"))
	port.SetText(strconv.Itoa(g_app.Preferences().IntWithFallback("port", 9091)))
	username.SetText(g_app.Preferences().StringWithFallback("username", "zhaobi"))
	passward.SetText(g_app.Preferences().StringWithFallback("passward", "ZhaoBi@"))

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Server/IP", Widget: serverip, HintText: "Server or IP address"},
			{Text: "Port", Widget: port, HintText: "port"},
			{Text: "Username", Widget: username},
			{Text: "Passward", Widget: passward},
		},
		OnSubmit: func() {
			iIp := serverip.Text
			iPort, err := strconv.Atoi(port.Text)
			iUsername := username.Text
			iPassward := passward.Text
			if err != nil {
				dialog.ShowConfirm("错误", err.Error(), func(ret bool) {}, g_win)
				return
			}

			g_app.Preferences().SetString("server", iIp)
			g_app.Preferences().SetInt("port", iPort)
			g_app.Preferences().SetString("username", iUsername)
			g_app.Preferences().SetString("passward", iPassward)

			g_win.SetContent(container.NewCenter(widget.NewLabel("Connecting Server ...\nPowered By ZhaoBi @2022")))
			go initTransmissionClient(iIp, iPort, iUsername, iPassward)
		},
	}

	g_win.SetContent(form)
}

func initTransmissionClient(iIp string, iPort int, iUsername string, iPassward string) {
	vBoxTop := container.NewVBox()

	vBoxBtm := container.NewVBox()

	err := g_mainTsClient.InitClient(iIp, iPort, iUsername, iPassward)
	if err != nil {
		//panic(err)
		g_win.SetContent(container.NewCenter(widget.NewLabel(err.Error())))
		g_win.Resize(fyne.NewSize(520, 410))

		time.Sleep(time.Duration(3) * time.Second)
		go initWelcomeContent()

		return
	}

	vBoxBtm.Add(widget.NewSeparator())
	gui_titleLabel = widget.NewLabel(fmt.Sprintf("当前有%d个种子正在下载...", 0))
	vBoxBtm.Add(gui_titleLabel)
	dSHBox := container.NewHBox(
		widget.NewButton("剧集", func() { g_mainTsClient.setShowListByDownloadType(DOWNLOADTYPE_TV, updateListArea) }),
		widget.NewButton("电影", func() { g_mainTsClient.setShowListByDownloadType(DOWNLOADTYPE_MOVIE, updateListArea) }),
		widget.NewButton("动漫", func() { g_mainTsClient.setShowListByDownloadType(DOWNLOADTYPE_ACG, updateListArea) }),
		widget.NewButton("综艺", func() { g_mainTsClient.setShowListByDownloadType(DOWNLOADTYPE_ARTSHOW, updateListArea) }),
		widget.NewButton("BT", func() { g_mainTsClient.setShowListByDownloadType(DOWNLOADTYPE_BT, updateListArea) }),
		widget.NewButton("下载中", func() { g_mainTsClient.setShowListInDownloading(updateListArea) }),
	)
	vBoxBtm.Add(dSHBox)

	//列表获取
	err1 := g_mainTsClient.GetAllTorrents()
	if err1 != nil {
		g_win.SetContent(container.NewCenter(widget.NewLabel(err1.Error())))
		g_win.Resize(fyne.NewSize(520, 410))
		return
	}

	//var mapProgressBarToIdx map[*widget.ProgressBar]int = make(map[*widget.ProgressBar]int)

	gui_dlListVBox = widget.NewList(func() int {
		g_mainTsClient.rwLock.RLock()
		defer g_mainTsClient.rwLock.RUnlock()
		return len(g_mainTsClient.showTorrents)
	}, func() fyne.CanvasObject {
		bar := NewDownloadBar() //;widget.NewProgressBar().Refresh()
		return bar
	}, func(lii widget.ListItemID, co fyne.CanvasObject) {
		pBar := co.(*myDownloadWidget)
		pBar.torrentShowingIdx = lii
		pBar.Refresh()
		//co.(*widget.Label).SetText(desc)
	})

	gui_dlListVBox.OnSelected = func(id widget.ListItemID) {
		DeleteATorrentData(id, func() {
			gui_dlListVBox.UnselectAll()
		})
	}

	g_mainTsClient.rwLock.RLock()
	gui_titleLabel.SetText(fmt.Sprintf("当前有%d个种子正在下载...", len(g_mainTsClient.showTorrents)))
	g_mainTsClient.rwLock.RUnlock()

	vBoxTop.Add(widget.NewSeparator())
	var dlType DownloadType = DOWNLOADTYPE_TV
	//创建新任务相关
	curDlTypeLable := widget.NewLabel(dlType.String())

	freeSpace, err2 := g_mainTsClient.GetFreeSpace()
	if err2 != nil {
		g_win.SetContent(container.NewCenter(widget.NewLabel(err2.Error())))
		g_win.Resize(fyne.NewSize(520, 410))
		return
	}

	vBoxTop.Add(container.NewHBox(widget.NewLabel(fmt.Sprintf("新建下载任务[%s Free]:", freeSpace)), curDlTypeLable))
	dTypeHBox := container.NewHBox(
		widget.NewButton("剧集", func() { dlType = DOWNLOADTYPE_TV; curDlTypeLable.SetText(dlType.String()) }),
		widget.NewButton("电影", func() { dlType = DOWNLOADTYPE_MOVIE; curDlTypeLable.SetText(dlType.String()) }),
		widget.NewButton("动漫", func() { dlType = DOWNLOADTYPE_ACG; curDlTypeLable.SetText(dlType.String()) }),
		widget.NewButton("综艺", func() { dlType = DOWNLOADTYPE_ARTSHOW; curDlTypeLable.SetText(dlType.String()) }),
		widget.NewButton("BT", func() { dlType = DOWNLOADTYPE_BT; curDlTypeLable.SetText(dlType.String()) }),
		//curDlTypeLable,
	)
	vBoxTop.Add(dTypeHBox)

	inputTitle := widget.NewEntry()
	inputTitle.SetPlaceHolder("请输入标题...")
	vBoxTop.Add(inputTitle)

	inputUrl := widget.NewEntry()
	inputUrl.SetPlaceHolder("请输入种子URL")
	vBoxTop.Add(inputUrl)

	vBoxTop.Add(widget.NewButton("提交", func() {
		StartADownload(dlType, inputTitle.Text, inputUrl.Text, inputTitle, inputUrl)
	}))
	vBoxTop.Add(widget.NewSeparator())

	content := container.NewBorder(vBoxTop, vBoxBtm, nil, nil, gui_dlListVBox)

	g_win.SetContent(content)
	g_win.Resize(fyne.NewSize(520, 410))

	go checkDownloadPrecent()
}

func checkDownloadPrecent() {

	for {
		time.Sleep(time.Duration(3) * time.Second)

		g_mainTsClient.rwLock.RLock()
		lenTs := len(g_mainTsClient.showTorrents)

		if lenTs <= 0 {
			g_mainTsClient.rwLock.RUnlock()
			continue
		}

		keys := make([]int64, 0, lenTs)
		keys = append(keys, g_mainTsClient.showTorrents...)

		g_mainTsClient.rwLock.RUnlock()

		torrents, err := g_mainTsClient.tsClient.TorrentGet(context.TODO(), []string{"id", "status", "downloadedEver", "rateDownload"}, keys)

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			g_mainTsClient.rwLock.Lock()

			for i, s := range torrents {
				g_mainTsClient.newTorrentsMapByID[*s.ID] = &torrents[i]
			}
			g_mainTsClient.rwLock.Unlock()

			updateListArea(false)
		}
	}
}

func StartADownload(dType DownloadType, name string, url string, inputName *widget.Entry, inputUrl *widget.Entry) {
	var dirTo string = dType.GetDownloadDir()

	dirTo = dirTo + name

	if len(url) < 5 {
		dialog.ShowInformation("提示", "URL不正确", g_win)
		return
	}

	var title string = "提示"
	if (url)[:4] == "http" {
		title = "种子链接!"
	} else if (url)[:7] == "magnet:" {
		title = "磁力链接!"
	} else {
		magnet := "magnet:?xt=urn:btih:" + url
		url = magnet
		title = "转为磁力链接!"
	}

	tips := fmt.Sprintf("%s\n%s", dType, name)

	fmt.Printf("提交任务:%s\n%s\n", dirTo, url)

	dialog.ShowConfirm(title, tips, func(ret bool) {
		if ret {
			inputName.SetText("")
			inputUrl.SetText("")
			go g_mainTsClient.submitDownload(&dirTo, &url, updateListArea)
		} else {
			fmt.Println("取消任务!")
		}
	}, g_win)

}

func DeleteATorrentData(idxInShowlist int, endFunc func()) {
	g_mainTsClient.rwLock.RLock()
	id := g_mainTsClient.showTorrents[idxInShowlist]
	dir := g_mainTsClient.torrentsMapByID[id].DownloadDir
	g_mainTsClient.rwLock.RUnlock()
	dialog.ShowConfirm("删除所有数据吗？", *dir, func(ret bool) {
		if ret {
			g_mainTsClient.deleteTorrent(id, updateListArea)
		}
		endFunc()
	}, g_win)
}

func updateListArea(skipLock bool) {
	go gui_dlListVBox.Refresh()
	if !skipLock {
		g_mainTsClient.rwLock.RLock()
	}
	gui_titleLabel.SetText(fmt.Sprintf("当前有%d个种子正在下载...", len(g_mainTsClient.showTorrents)))
	if !skipLock {
		g_mainTsClient.rwLock.RUnlock()
	}
}
