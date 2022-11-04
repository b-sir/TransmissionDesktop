package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hekmon/cunits/v2"
	"github.com/hekmon/transmissionrpc/v2"
)

type myDownloadWidgetRenderer struct {
	mainLabel  *canvas.Text
	tipsLabel  *canvas.Text
	barBg, bar *canvas.Rectangle

	objects        []fyne.CanvasObject
	downloadWidget *myDownloadWidget
}

func (r *myDownloadWidgetRenderer) Layout(size fyne.Size) {
	mainAreaSize := fyne.Size{Width: size.Width, Height: size.Height - 1}
	r.barBg.Resize(mainAreaSize)
	r.bar.Move(fyne.Position{X: 0, Y: mainAreaSize.Height - 2})
	r.bar.Resize(fyne.Size{Width: 0, Height: 2})

	inset := fyne.Position{X: size.Width / 2, Y: 0}

	r.mainLabel.Text = "Title"
	r.tipsLabel.Text = "Tips"

	tsize := fyne.MeasureText(r.mainLabel.Text, r.mainLabel.TextSize, r.mainLabel.TextStyle)

	r.mainLabel.Move(inset)
	inset = inset.Add(fyne.Position{X: 0, Y: tsize.Height})
	r.tipsLabel.Move(inset)

	r.updateBar()
}

func (r *myDownloadWidgetRenderer) MinSize() fyne.Size {
	fontsize := r.mainLabel.TextSize
	tsize := fyne.MeasureText("100%", fontsize, r.mainLabel.TextStyle)

	return fyne.NewSize(tsize.Width+theme.Padding()*2, tsize.Height*1.7+2)
}

func (r *myDownloadWidgetRenderer) Refresh() {
	r.updateBar()

	r.bar.Refresh()
	r.barBg.Refresh()
	r.mainLabel.Refresh()
	r.tipsLabel.Refresh()

	//canvas.Refresh(r.downloadWidget)
}

func (r *myDownloadWidgetRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *myDownloadWidgetRenderer) Destroy() {
}

func (r *myDownloadWidgetRenderer) updateBar() {
	idx := r.downloadWidget.torrentShowingIdx

	g_mainTsClient.rwLock.RLock()
	defer g_mainTsClient.rwLock.RUnlock()

	if idx >= len(g_mainTsClient.showTorrents) {
		return
	}

	tID := g_mainTsClient.showTorrents[idx]

	ptrTrt := g_mainTsClient.torrentsMapByID[tID]
	ptrNewTrt := g_mainTsClient.newTorrentsMapByID[tID]

	widgetSize := r.downloadWidget.Size()
	mainAreaSize := fyne.Size{Width: widgetSize.Width, Height: widgetSize.Height - 1}
	if ptrTrt == nil || ptrTrt.DownloadDir == nil || ptrTrt.TotalSize == nil {
		//return "Creating ..."
		r.bar.Resize(fyne.Size{Width: 0, Height: 2})
		r.mainLabel.Text = "Creating ..."
		r.tipsLabel.Text = ""
		return
	}

	//DownloadDir TotalSize 通常不更新，从初种中取信息
	prefix := (*ptrTrt.DownloadDir)[:17]
	var fname string
	if prefix == "/mnt/utdir/video/" {
		fname = (*ptrTrt.DownloadDir)[17:]
	} else {
		fname = (*ptrTrt.DownloadDir)[14:]
	}
	totalSize := ptrTrt.TotalSize

	var torrent *transmissionrpc.Torrent = nil
	if ptrNewTrt != nil {
		torrent = ptrNewTrt
	} else {
		torrent = ptrTrt
	}

	if *torrent.Status == transmissionrpc.TorrentStatusDownload {
		var downloadSize = (cunits.Bits)(*torrent.DownloadedEver * 8)
		prec := (float32)(downloadSize) / (float32)(*totalSize)
		speed := torrent.ConvertDownloadSpeed()

		desc := fmt.Sprintf("(%s/s)[%s/%s]  %.2f%%", speed, downloadSize, *totalSize, prec*100)
		r.mainLabel.Text = fname
		r.tipsLabel.Text = desc

		r.bar.Resize(fyne.NewSize(mainAreaSize.Width*prec, 2))
	} else {
		r.mainLabel.Text = fname
		r.tipsLabel.Text = "seeding ..."
		r.bar.Resize(fyne.NewSize(mainAreaSize.Width, 2))
	}
}

type myDownloadWidget struct {
	widget.BaseWidget
	torrentShowingIdx int
}

func (w *myDownloadWidget) SetTorrentIdx(idx int) {
	w.torrentShowingIdx = idx
}

func (w *myDownloadWidget) CreateRenderer() fyne.WidgetRenderer {
	w.ExtendBaseWidget(w)

	barBg := canvas.NewRectangle(color.NRGBA{R: 0x88, G: 0xf6, B: 0xe9, A: 0x1c}) //88F6E9
	bar := canvas.NewRectangle(color.NRGBA{R: 0x88, G: 0xf6, B: 0xe9, A: 0x7f})   //88F6E9
	mainLabel := canvas.NewText("Name", theme.ForegroundColor())
	mainLabel.Alignment = fyne.TextAlignCenter
	tipsLabel := canvas.NewText("Tips", theme.ForegroundColor())
	tipsLabel.TextSize = mainLabel.TextSize * 0.7
	tipsLabel.Alignment = fyne.TextAlignCenter

	renderer := &myDownloadWidgetRenderer{
		mainLabel:      mainLabel,
		tipsLabel:      tipsLabel,
		barBg:          barBg,
		bar:            bar,
		downloadWidget: w,
		objects:        []fyne.CanvasObject{barBg, bar, mainLabel, tipsLabel},
	}

	return renderer
}

func (w *myDownloadWidget) MinSize() fyne.Size {
	w.ExtendBaseWidget(w)

	return w.BaseWidget.MinSize()
}

func NewDownloadBar() *myDownloadWidget {
	p := &myDownloadWidget{torrentShowingIdx: 0}

	//cache.Renderer(p).Layout(p.MinSize())
	return p
}
