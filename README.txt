50% 复制 + 50% AI生成

需要修改的文件
    config.json
        nonpaid：true/false, true为不记录免费礼物与弹幕, false为记录
        cookies：bilibili网页登陆cookies，不填则无法记录观众id
    livers.json
        需要记录的主播名单

存储的文件
    log.txt
        开播、下播信息
    data/YYYYMM_free_livername.csv
        时间戳,开播,时间
        时间戳,下播
        时间戳,粉丝更新,粉丝数,粉丝团
        时间戳,高能,高能,在线
        时间戳,弹幕,观众id,内容
        时间戳,免费礼物,观众id,礼物名,数量
    data/YYYYMM_paid_livername.csv
        时间戳,SC,观众id,内容,价格
        时间戳,付费礼物,观众id,礼物名,数量,价格
        时间戳,舰队,观众id,价格
    count/日期_count.csv
        统计结果
    data 中文件将每月移动到 backup

使用方法
    newbili
        开始记录
    newbili -count YYYYMM
        统计月份的数据
    newbili -count YYYYMMDD
        统计日期的数据
