package main

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type data struct {
	name  string  `csv:"name"`
	date  string  `csv:"date"`
	sc    float64 `csv:"sc"`
	gift  float64 `csv:"gift"`
	guard float64 `csv:"guard"`
	all   float64 `csv:"all"`
}

type datas []*data

func CountData(datestr string) {
	var ds = datas{}

	if len(datestr) != 8 && len(datestr) != 6 {
		slog.Error("日期格式错误")
		return
	}

	stamps, stampe := timearea(datestr)
	pathes := findFile(datestr)

	for _, v := range pathes {
		s := strings.Split(strings.TrimSuffix(path.Base(v), path.Ext(v)), "_")
		if len(s) < 3 || s[1] != "paid" {
			continue
		}
		d := &data{
			name:  s[2],
			date:  datestr,
			sc:    0,
			gift:  0,
			guard: 0,
			all:   0,
		}
		ds = append(ds, d)
		d.count(v, stamps, stampe)
	}
	for _, v := range ds {
		fmt.Printf("|%.2s|%9.1f|%9.1f|%9.1f|%9.1f|\n", v.name, v.sc, v.gift, v.guard, v.all)
	}
	makeDir("./count")
	csvfile, err := os.OpenFile("./count/"+datestr+"_count.csv", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer csvfile.Close()
	writer := csv.NewWriter(csvfile)
	defer writer.Flush()
	header := []string{"name", "date", "sc", "gift", "guard", "all"}
	writer.Write(header)
	for _, v := range ds {
		writer.Write([]string{
			v.name,
			v.date,
			fmt.Sprintf("%.2f", v.sc),
			fmt.Sprintf("%.2f", v.gift),
			fmt.Sprintf("%.2f", v.guard),
			fmt.Sprintf("%.2f", v.all),
		})
	}
}
func (d *data) count(path string, start int64, end int64) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		if len(record) < 4 {
			continue
		}
		stamp, _ := strconv.ParseInt(record[0], 10, 64)
		if stamp < start || stamp >= end {
			continue
		}
		if len(record) == 4 {
			guard, _ := strconv.ParseFloat(record[3], 32)
			d.guard += guard
		}
		if len(record) == 5 {
			sc, _ := strconv.ParseFloat(record[4], 32)
			d.sc += sc
		}
		if len(record) == 6 {
			num, _ := strconv.ParseFloat(record[4], 32)
			price, _ := strconv.ParseFloat(record[5], 32)
			d.gift += num * price
		}
	}
	d.all = d.sc + d.gift + d.guard
}

func timearea(date string) (int64, int64) {
	temp := func() string {
		if len(date) == 8 {
			return "20060102"
		}
		return "200601"
	}
	stamp, _ := time.ParseInLocation(temp(), date, time.Local)

	if len(date) == 8 {
		return stamp.UnixMilli(), stamp.AddDate(0, 0, 1).UnixMilli()
	}
	return stamp.UnixMilli(), stamp.AddDate(0, 1, 0).UnixMilli()
}

func findFile(date string) []string {
	var pathes []string

	for _, v := range []string{config.Dir, config.Backdir} {
		filepath.Walk(v, func(p string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if len(filepath.Base(p)) < len(date) {
				return nil
			}
			if filepath.Base(p)[:6] == date[:6] && filepath.Ext(p) == ".csv" {
				pathes = append(pathes, p)
			}
			return nil
		})

	}
	return pathes
}
