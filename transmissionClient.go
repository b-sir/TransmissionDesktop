package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"fyne.io/fyne/v2/dialog"
	"github.com/hekmon/cunits/v2"
	"github.com/hekmon/transmissionrpc/v2"
)

type transmissionClient struct {
	tsClient           *transmissionrpc.Client
	showTorrents       []int64
	torrentsMapByID    map[int64]*transmissionrpc.Torrent
	newTorrentsMapByID map[int64]*transmissionrpc.Torrent

	rwLock sync.RWMutex
}

func (c *transmissionClient) InitClient(iIp string, iPort int, iUsername string, iPassward string) error {
	var err0 error
	c.tsClient, err0 = transmissionrpc.New(iIp, iUsername, iPassward, &transmissionrpc.AdvancedConfig{HTTPS: false, Port: uint16(iPort)})

	if err0 != nil {
		return err0
	}

	ok, serverVersion, serverMinimumVersion, err1 := c.tsClient.RPCVersion(context.TODO())
	if err1 != nil {
		return err1
	}

	if !ok {
		return fmt.Errorf("remote transmission RPC version (v%d) is incompatible with the transmission library (v%d): remote needs at least v%d",
			serverVersion, transmissionrpc.RPCVersion, serverMinimumVersion)
	}

	return nil
}

func (c *transmissionClient) GetAllTorrents() error {
	torrents, err := c.tsClient.TorrentGetAll(context.TODO())

	if err != nil {
		return err
	}

	c.rwLock.Lock()
	defer c.rwLock.Unlock()

	c.showTorrents = make([]int64, 0, len(torrents))
	c.torrentsMapByID = make(map[int64]*transmissionrpc.Torrent)
	c.newTorrentsMapByID = make(map[int64]*transmissionrpc.Torrent)
	for i, torrent := range torrents {
		//if *torrent.Status == transmissionrpc.TorrentStatusDownload || *torrent.Status == transmissionrpc.TorrentStatusDownloadWait {
		c.showTorrents = append(c.showTorrents, *torrents[i].ID)
		//}

		if oldT, ok := c.torrentsMapByID[*torrent.ID]; ok {
			fmt.Println("种子ID冲突", *oldT.DownloadDir, *torrent.DownloadDir)
		} else {
			c.torrentsMapByID[*torrent.ID] = &torrents[i]
		}
	}

	return nil
}

func (c *transmissionClient) submitDownload(dirTo *string, url *string, updateFunc func(skipLock bool)) {
	torrent, err := c.tsClient.TorrentAdd(context.TODO(), transmissionrpc.TorrentAddPayload{
		Filename:    url,
		DownloadDir: dirTo,
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		dialog.ShowConfirm("错误", err.Error(), func(ret bool) {}, g_win)
	} else {
		// Only 3 fields will be returned/set in the Torrent struct
		// fmt.Println("添加成功!")
		// fmt.Println(*torrent.ID)
		// fmt.Println(*torrent.Name)
		// fmt.Println(*torrent.HashString)

		//torrent.DownloadDir = dirTo
		c.rwLock.Lock()

		//newSlot := len(c.showTorrents)
		c.showTorrents = append(c.showTorrents, *torrent.ID)
		c.torrentsMapByID[*torrent.ID] = &torrent

		c.rwLock.Unlock()
		updateFunc(false)

		for {
			torrents2, err2 := c.tsClient.TorrentGet(context.TODO(), []string{"id", "status", "downloadedEver", "rateDownload", "totalSize", "downloadDir"}, []int64{*torrent.ID})

			if err2 != nil {
				fmt.Fprintln(os.Stderr, err)
			} else {
				//c.showTorrents[newSlot] = &torrents2[0]
				if *torrents2[0].TotalSize > 0 {
					c.rwLock.Lock()
					c.torrentsMapByID[*torrent.ID] = &torrents2[0]
					c.rwLock.Unlock()
					updateFunc(false)
					break
				} else { //比如磁力链接正在转为种子，未能获取很多信息
					c.rwLock.Lock()
					c.torrentsMapByID[*torrent.ID] = &torrents2[0]
					c.rwLock.Unlock()
					time.Sleep(time.Duration(2) * time.Second)
				}
			}
		}
	}
}

// 移除任务、会将磁盘数据一起清空
func (c *transmissionClient) deleteTorrent(id int64, updateFunc func(skipLock bool)) {
	err := c.tsClient.TorrentRemove(context.TODO(), transmissionrpc.TorrentRemovePayload{
		IDs:             []int64{id},
		DeleteLocalData: true,
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		//showTorrents       []*transmissionrpc.Torrent
		//torrentsMapByID    map[int64]*transmissionrpc.Torrent
		//newTorrentsMapByID map[int64]*transmissionrpc.Torrent
		c.rwLock.Lock()
		defer c.rwLock.Unlock()

		delete(c.torrentsMapByID, id)
		delete(c.newTorrentsMapByID, id)
		for idx, iId := range c.showTorrents {
			if iId == id {
				c.showTorrents = append(c.showTorrents[:idx], c.showTorrents[idx+1:]...)
			}
		}
		updateFunc(true)
	}
}

func (c *transmissionClient) GetFreeSpace() (cunits.Bits, error) {
	freeBits, err := c.tsClient.FreeSpace(context.TODO(), "/mnt/utdir/video")

	return freeBits, err
}

func (c *transmissionClient) setShowListByDownloadType(tp DownloadType, updateFunc func(skipLock bool)) {
	c.rwLock.Lock()
	defer c.rwLock.Unlock()

	c.showTorrents = make([]int64, 0, len(c.torrentsMapByID))
	if tp == DOWNLOADTYPE_UNKNOWN {
		fmt.Println("UNKNOWN TYPE")
	} else {
		targetDir := tp.GetDownloadDir()
		lenT := len(targetDir)
		for _, pTorrent := range c.torrentsMapByID {
			if pTorrent.DownloadDir == nil || (*pTorrent.DownloadDir)[:lenT] == targetDir {
				c.showTorrents = append(c.showTorrents, *pTorrent.ID)
			}
		}
	}
	updateFunc(true)
}

func (c *transmissionClient) setShowListInDownloading(updateFunc func(skipLock bool)) {
	c.rwLock.Lock()
	defer c.rwLock.Unlock()
	c.showTorrents = make([]int64, 0, len(c.torrentsMapByID))

	for _, pTorrent := range c.torrentsMapByID {
		var torrent *transmissionrpc.Torrent = pTorrent
		if p, ok := c.newTorrentsMapByID[*pTorrent.ID]; ok {
			torrent = p
		}

		if torrent.Status == nil || *torrent.Status == transmissionrpc.TorrentStatusDownload || *torrent.Status == transmissionrpc.TorrentStatusDownloadWait {
			c.showTorrents = append(c.showTorrents, *pTorrent.ID)
		}
	}
	updateFunc(true)
}
