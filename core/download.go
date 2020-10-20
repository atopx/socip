package core

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	path "path/filepath"

	"github.com/yanmengfei/socip/global"
)

var baseURL = "https://mall.ipplus360.com/download/file?downloadId=%s&fileName=%s"

type source struct {
	FileName string
	Output   string
}

/**
 * 错误处理方法: 抛出异常
 * @param err: error
 */
func errorManager(err error) {
	if err != nil {
		panic(err)
	}
}

/**
 * 解压zip文件
 * @param filepath: 文件路径
 * @param output: 解压文件名称
 * @param newName: 重命名解压后的文件名
 * @return: 解压输出的文件路径
 */
func unzip(filepath string, output string, newName string) string {
	// 调用系统命令解压
	cmd := exec.Command("unzip", "-P", global.Config.Password, filepath, output, "-d", global.Config.Folder)
	err := cmd.Run()
	errorManager(err)
	// 输出文件重命名
	oldPath := path.Join(global.Config.Folder, output)
	newPath := path.Join(global.Config.Folder, newName)
	err = os.Rename(oldPath, newPath)
	errorManager(err)
	// 移除zip压缩文件
	err = os.Remove(filepath)
	errorManager(err)
	global.Logger.Println("Basic解压完成")
	return newPath
}

/**
 * 下载zip文件
 * @param filename: 需要下载的文件名称，作为URL参数
 * @param downloadId: 需要下载的文件ID, 作为URL参数
 * @return: 下载后的文件路径
 */
func download(filename string, id string) string {
	// 拼接路径
	var filepath = path.Join(global.Config.Folder, filename+".zip")
	var url = fmt.Sprintf(baseURL, id, filename)
	// 创建文件缓冲
	out, err := os.Create(filepath)
	errorManager(err)
	defer func() { _ = out.Close() }()
	// 创建http client
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	// 发送http请求
	resp, err := client.Get(url)
	errorManager(err)
	defer func() { _ = resp.Body.Close() }()
	// 状态码判断
	if resp.StatusCode != http.StatusOK {
		errorManager(fmt.Errorf("bad status: %s", resp.Status))
	}
	// 保存到文件
	_, err = io.Copy(out, resp.Body)
	errorManager(err)
	return filepath
}

type BasicData source

func (data *BasicData) Get() (filepath string) {
	global.Logger.Println("Download Basic Data -> Start")
	filepath = download(data.FileName, global.Config.Basic.Id)
	global.Logger.Println("Download Basic Data -> End")
	filepath = unzip(filepath, data.Output, global.Config.Basic.Input)
	global.Logger.Println("Success: ", filepath)
	return filepath
}

type SceneData source

func (data *SceneData) Get() (filepath string) {
	global.Logger.Println("Download Scene Data -> Start")
	filepath = download(data.FileName, global.Config.Scene.Id)
	global.Logger.Println("Download Scene Data -> End")
	filepath = unzip(filepath, data.Output, global.Config.Scene.Input)
	global.Logger.Println("Success: ", filepath)
	return filepath
}
