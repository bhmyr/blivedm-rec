package main

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

// 红包
type POPULARITY_RED_POCKET_NEW struct {
	Cmd  string `json:"cmd"`
	Data struct {
		LotID       int    `json:"lot_id"`
		StartTime   int    `json:"start_time"`
		CurrentTime int    `json:"current_time"`
		WaitNum     int    `json:"wait_num"`
		Uname       string `json:"uname"`
		Uid         int64  `json:"uid"`
		Action      string `json:"action"`
		Num         int    `json:"num"`
		GiftName    string `json:"gift_name"`
		GiftID      int    `json:"gift_id"`
		Price       int    `json:"price"`
		NameColor   string `json:"name_color"`
		MedalInfo   struct {
			TargetID         int    `json:"target_id"`
			Special          string `json:"special"`
			IconID           int    `json:"icon_id"`
			AnchorUname      string `json:"anchor_uname"`
			AnchorRoomid     int    `json:"anchor_roomid"`
			MedalLevel       int    `json:"medal_level"`
			MedalName        string `json:"medal_name"`
			MedalColor       int    `json:"medal_color"`
			MedalColorStart  int    `json:"medal_color_start"`
			MedalColorEnd    int    `json:"medal_color_end"`
			MedalColorBorder int    `json:"medal_color_border"`
			IsLighted        int    `json:"is_lighted"`
			GuardLevel       int    `json:"guard_level"`
		} `json:"medal_info"`
	} `json:"data"`
}

// 高能
type ONLINE_RANK_COUNT struct {
	Cmd  string `json:"cmd"`
	Data struct {
		Count           int    `json:"count"`
		CountText       string `json:"count_text"`
		OnlineCount     int    `json:"online_count"`
		OnlineCountText string `json:"online_count_text"`
	} `json:"data"`
}
