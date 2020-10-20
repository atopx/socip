package main

import (
	"flag"
	"math/rand"
	"path"
	"strconv"

	"github.com/oschwald/maxminddb-golang"
	"github.com/yanmengfei/socip/core"
	"github.com/yanmengfei/socip/global"
)

func RandInt(min, max int) int {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	return rand.Intn(max-min) + min
}

type basicResult struct {
	Country   string `json:"country"`
	Province  string `json:"province"`
	City      string `json:"city"`
	District  string `json:"district"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	Owner     string `json:"owner"`
	Source    string `json:"source"`
}

type sceneResult struct {
	Scene string `maxminddb:"scene"`
}

func validateScene(db *maxminddb.Reader, fields []string) bool {
	var result sceneResult
	min, _ := strconv.Atoi(fields[0])
	max, _ := strconv.Atoi(fields[1])
	validateIpv4 := core.DecToIpv4(RandInt(min, max))
	err := db.Lookup(validateIpv4, &result)
	if err != nil {
		return false
	}
	return result.Scene == fields[2]
}

func validateBasic(db *maxminddb.Reader, fields []string) bool {
	var result basicResult
	min, _ := strconv.Atoi(fields[1])
	max, _ := strconv.Atoi(fields[2])
	validateIpv4 := core.DecToIpv4(RandInt(min, max))
	err := db.Lookup(validateIpv4, &result)
	if err != nil {
		return false
	}
	validateString := fields[6] + fields[7] + fields[8] + fields[9] + fields[12] + fields[13] + fields[17]
	matchString := result.Country + result.Province + result.City + result.District + result.Longitude + result.Latitude + result.Owner
	return matchString == validateString
}

func validateSocDB(totalCount int, dbType string) bool {

	var db *maxminddb.Reader
	var err error
	var filepath string
	var callback func(db *maxminddb.Reader, fields []string) bool

	switch dbType {
	case "basic":
		filepath = path.Join(global.Config.Folder, global.Config.Basic.Input)
		db, err = maxminddb.Open(global.Config.Basic.Output)
		callback = validateBasic
	case "scene":
		filepath = path.Join(global.Config.Folder, global.Config.Scene.Input)
		db, err = maxminddb.Open(global.Config.Scene.Output)
		callback = validateScene
	}
	if err != nil {
		global.Logger.Panic(err)
	}

	lines := make(chan string)
	reader := make(chan error)
	go core.LoadSource(filepath, lines, reader)

	var successCount, failedCount int
	for line := range lines {
		fields := core.ParseLine(line)
		if callback(db, fields) {
			global.Logger.Println("Validate scene success!")
			successCount += 1
		} else {
			global.Logger.Println("Validate scene failed!")
			failedCount += 1
		}
		if successCount+failedCount == totalCount {
			break
		}
	}
	global.Logger.Printf("Success: %d, Failed: %d, SuccessRate: %d \n", successCount, failedCount, successCount/totalCount)
	return failedCount == 0
}

func main() {
	var t string
	var n int
	flag.StringVar(&t, "t", "all", "验证的数据库类型,支持[basic, scene, all]")
	flag.IntVar(&n, "n", 50000, "验证的个数, 范围[10000 - 200000]")
	flag.Parse()

	if n < 10000 || n > 200000 {
		global.Logger.Println("无效的参数，请指定 -n + 验证的个数, 范围[10000 - 200000]")
	} else {
		if t == "basic" {
			validateSocDB(n, "basic")
		} else if t == "scene" {
			validateSocDB(n, "scene")
		} else if t == "all" {
			validateSocDB(n, t)
			validateSocDB(n, t)
		} else {
			global.Logger.Println("无效的参数，请指定 -t + 验证的数据库类型,支持[basic, scene, all]")
		}
	}
}
