package qiniu

import (
	"github.com/lvfeiyang/proxy/common"
	"github.com/lvfeiyang/proxy/common/config"
	"github.com/lvfeiyang/proxy/common/flog"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

func DelFile(bucket, file string) error {
	bucketManager := getBucketManager()
	return bucketManager.Delete(bucket, file)
}
func DelRepeatFile(bucket string) ([]string, error) {
	var delFiles []string
	var existHash []string
	var marker string
	bucketManager := getBucketManager()
	for {
		entries, _, nextMarker, hasNext, err := bucketManager.ListFiles(bucket, "", "", marker, 1000)
		if err != nil {
			return delFiles, err
		}
		//处理相同hash的文件
		for _, entry := range entries {
			if common.Exist(existHash, entry.Hash) {
				if err := bucketManager.Delete(bucket, entry.Key); err != nil {
					flog.LogFile.Println("del qiniu", err)
				}
				delFiles = append(delFiles, entry.Key)
			} else {
				existHash = append(existHash, entry.Hash)
			}
		}

		if hasNext {
			marker = nextMarker
		} else {
			break
		}
	}
	return delFiles, nil
}
func getBucketManager() *storage.BucketManager {
	mac := qbox.NewMac(config.ConfigVal.QiniuAK, config.ConfigVal.QiniuSK)
	cfg := storage.Config{
		UseHTTPS: false,
	}
	return storage.NewBucketManager(mac, &cfg)
}
func GetAllFile(bucket string) ([]string, error) {
	var files []string
	var marker string
	bucketManager := getBucketManager()
	for {
		entries, _, nextMarker, hasNext, err := bucketManager.ListFiles(bucket, "", "", marker, 1000)
		if err != nil {
			return files, err
		}
		for _, entry := range entries {
			files = append(files, entry.Key)
		}
		if hasNext {
			marker = nextMarker
		} else {
			break
		}
	}
	return files, nil
}
