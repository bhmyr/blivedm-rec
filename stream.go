package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Akegarasu/blivedm-go/message"
)

var GuardLevel = map[int]string{
	1: "总督",
	2: "提督",
	3: "舰长",
}

// 主播粉丝实时更新
type ROOM_REAL_TIME_MESSAGE_UPDATE struct {
	Cmd  string `json:"cmd"`
	Data struct {
		Roomid    int `json:"roomid"`
		Fans      int `json:"fans"`
		RedNotice int `json:"red_notice"`
		FansClub  int `json:"fans_club"`
	} `json:"data"`
}

func (liver *Liver) fetch() {
	c := liver.Client
	// 开播
	c.OnLive(func(live *message.Live) {
		liver.Writer.Write(false, []string{
			fmt.Sprint(time.Now().UnixMilli()),
			"开播",
			fmt.Sprint(live.LiveTime),
		})
		logger.Info(liver.Name + " 开播")
	})
	// PREPARING 下播？
	c.RegisterCustomEventHandler("PREPARING", func(s string) {
		liver.Writer.Write(false, []string{
			fmt.Sprint(time.Now().UnixMilli()),
			"下播",
		})
		logger.Info(liver.Name + " 下播")
	})

	// 主播粉丝实时更新
	c.RegisterCustomEventHandler("ROOM_REAL_TIME_MESSAGE_UPDATE", func(s string) {
		var data = ROOM_REAL_TIME_MESSAGE_UPDATE{}
		json.Unmarshal([]byte(s), &data)
		liver.Writer.Write(false, []string{
			fmt.Sprint(time.Now().UnixMilli()),
			"粉丝更新",
			fmt.Sprint(data.Data.Fans),
			fmt.Sprint(data.Data.FansClub),
		})
	})
	// 弹幕事件
	if config.Nonpaid {
		c.OnDanmaku(func(danmaku *message.Danmaku) {
			liver.Writer.Write(false, []string{
				fmt.Sprint(time.Now().UnixMilli()),
				"弹幕",
				fmt.Sprint(danmaku.Sender.Uid),
				danmaku.Content,
			})
		})
	}
	// 醒目留言事件
	c.OnSuperChat(func(superChat *message.SuperChat) {
		liver.Writer.Write(true, []string{
			fmt.Sprint(time.Now().UnixMilli()),
			"SC",
			fmt.Sprint(superChat.Uid),
			superChat.Message,
			fmt.Sprint(superChat.Price),
		})
	})
	// 礼物事件
	c.OnGift(func(gift *message.Gift) {
		if gift.CoinType == "gold" {
			liver.Writer.Write(true, []string{
				fmt.Sprint(time.Now().UnixMilli()),
				"付费礼物",
				fmt.Sprint(gift.Uid),
				gift.GiftName,
				fmt.Sprint(gift.Num),
				fmt.Sprint(float32(gift.Price) / 1000.0),
			})
		} else if gift.CoinType == "silver" && config.Nonpaid {
			liver.Writer.Write(false, []string{
				fmt.Sprint(time.Now().UnixMilli()),
				"免费礼物",
				fmt.Sprint(gift.Uid),
				gift.GiftName,
				fmt.Sprint(gift.Num),
			})
		}
	})
	// 上舰事件
	c.OnUserToast(func(userToast *message.UserToast) {
		liver.Writer.Write(true, []string{
			fmt.Sprint(time.Now().UnixMilli()),
			GuardLevel[userToast.GuardLevel],
			fmt.Sprint(userToast.Uid),
			fmt.Sprint(userToast.Price / 1000),
		})
	})

}

func (liver *Liver) Stream(ctx context.Context) {
	c := liver.Client
	go liver.fetch()
	err := c.Start()
	if err != nil {
		logger.Error(err.Error())
	}
	logger.Info(liver.Name + " started")
	select {
	case <-ctx.Done():
		fmt.Println("stop ", liver.Name)
		c.Stop()
		liver.Writer.Stop()
	default:
	}
}
