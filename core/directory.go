package core

import (
	"os"

	"github.com/yanmengfei/socip/global"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreateDir(dirs ...string) (err error) {
	for _, v := range dirs {
		exist, err := PathExists(v)
		if err != nil {
			return err
		}
		if !exist {
			global.Logger.Println("create directory", v)
			err = os.MkdirAll(v, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}
	}
	return err
}
