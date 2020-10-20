package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/yanmengfei/socip/core"
	"github.com/yanmengfei/socip/core/socdb"
	"github.com/yanmengfei/socip/global"
)

func init() {
	exists, err := core.PathExists(global.Config.Folder)
	if err != nil {
		_ = core.CreateDir(global.Config.Folder)
	}
	if exists == false {
		_ = core.CreateDir(global.Config.Folder)
	}
}

func parseSceneLine(line string) (net.IP, net.IP, socdb.Map) {
	fields := core.ParseLine(line)
	start, _ := strconv.Atoi(fields[0])
	end, _ := strconv.Atoi(fields[1])
	return core.DecToIpv4(start), core.DecToIpv4(end), socdb.Map{
		"scene":  socdb.String(fields[2]),
		"source": socdb.String("soc-scene"), // 来源
	}
}

func parseBasicLine(line string) (net.IP, net.IP, socdb.Map) {
	fields := core.ParseLine(line)
	start, _ := strconv.Atoi(fields[1])
	end, _ := strconv.Atoi(fields[2])
	return core.DecToIpv4(start), core.DecToIpv4(end), socdb.Map{
		"continent": socdb.String(fields[3]),   // 大陆
		"country":   socdb.String(fields[6]),   // 国家
		"province":  socdb.String(fields[7]),   // 省
		"city":      socdb.String(fields[8]),   // 市
		"district":  socdb.String(fields[9]),   // 区
		"longitude": socdb.String(fields[12]),  // 经度
		"latitude":  socdb.String(fields[13]),  // 纬度
		"radius":    socdb.String(fields[14]),  // 范围
		"scene":     socdb.String(fields[15]),  // 场景
		"accuracy":  socdb.String(fields[16]),  // 精准度
		"owner":     socdb.String(fields[17]),  // 运营商
		"zipcode":   socdb.String(""),          // 邮编
		"source":    socdb.String("soc-geoip"), // 来源
	}
}

func buildSocDB(filepath string, dbType string) {
	options := socdb.Options{
		IPVersion:               global.Config.IpVersion,
		RecordSize:              global.Config.RecordSize,
		Languages:               []string{"en", "cn"},
		IncludeReservedNetworks: true,
	}
	var outputFile string
	var callback func(line string) (net.IP, net.IP, socdb.Map)
	switch dbType {
	case "basic":
		callback = parseBasicLine
		outputFile = global.Config.Basic.Output
		options.DatabaseType = global.Config.Basic.DatabaseType
		options.Description = global.Config.Basic.Description
	case "scene":
		callback = parseSceneLine
		outputFile = global.Config.Scene.Output
		options.DatabaseType = global.Config.Scene.DatabaseType
		options.Description = global.Config.Scene.Description
	default:
		outputFile = global.Config.Basic.Output
		options.DatabaseType = global.Config.Basic.DatabaseType
		options.Description = global.Config.Basic.Description
		callback = parseBasicLine
	}

	writer, err := socdb.New(options)
	if err != nil {
		panic(err)
	}

	lines := make(chan string)
	reader := make(chan error)

	go core.LoadSource(filepath, lines, reader)

	for line := range lines {
		start, end, data := callback(line)

		for _, cidr := range core.IPRange2CIDRs(start, end) {
			err := writer.Insert(cidr, data)
			if err != nil {
				panic(err)
			}
			global.Logger.Println(cidr, " --> ", data)
		}
	}
	if err := <-reader; err != nil {
		panic(err)
	}

	output, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	_, err = writer.WriteTo(output)
	if err != nil {
		panic(err)
	}
	global.Logger.Println("Build", dbType, "DB success!")
}

func buildBasicDB(isoWeekString string) {
	var basic = core.BasicData{
		FileName: fmt.Sprintf("IP_basic_%s_single_WGS84_mysql", isoWeekString),
		Output:   fmt.Sprintf("IP_basic_%s_single_WGS84.txt", isoWeekString),
	}
	filepath := basic.Get()
	buildSocDB(filepath, "basic")
}

func buildSceneDB(isoWeekString string) {
	var scene = core.SceneData{
		FileName: fmt.Sprintf("IP_scene_all_cn_%s_mysql", isoWeekString),
		Output:   fmt.Sprintf("IP_scene_all_cn_%s.txt", isoWeekString),
	}
	filepath := scene.Get()
	buildSocDB(filepath, "scene")
}

func main() {
	var t string
	flag.StringVar(&t, "t", "all", "构建的数据库类型,支持[basic, scene, all]")
	flag.Parse()

	var year, week int = time.Now().ISOWeek()
	var isoWeekString = fmt.Sprintf("%dW%d", year, week)
	if t == "basic" {
		buildBasicDB(isoWeekString)
	} else if t == "scene" {
		buildSceneDB(isoWeekString)
	} else if t == "all" {
		go buildSceneDB(isoWeekString)
		buildBasicDB(isoWeekString)
	} else {
		global.Logger.Println("无效的参数，请指定 -t + 构建的数据库类型,支持[basic, scene, all]")
	}
}
