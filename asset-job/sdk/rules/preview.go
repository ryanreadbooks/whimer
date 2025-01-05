package rules

import (
	"crypto/md5"
	"encoding/hex"
	"hash/crc32"

	"github.com/ryanreadbooks/whimer/misc/utils"
)

const (
	numShard = 64
)

var (
	shards [numShard]string
)

func init() {
	// 固定的分片规则
	const magic = "WHIMER_OSS_MAGIC_7C00"
	sumer := md5.New()

	for i := range numShard {
		sumer.Reset()
		var temp = magic

		for range i + 1 {
			sumer.Write([]byte(temp))
			res := sumer.Sum(nil)
			temp = hex.EncodeToString(res)
			sumer.Reset()
		}

		shards[i] = temp
	}

}

const PreviewSuffix = "@prv_webp_50"

// 生成预览图的key
//
// parentKey: 原图key，不带bucket
func PreviewKey(parentKey string) string {
	idx := crc32.ChecksumIEEE(utils.StringToBytes(parentKey)) % numShard
	return shards[idx] + "/" + parentKey + PreviewSuffix
}
