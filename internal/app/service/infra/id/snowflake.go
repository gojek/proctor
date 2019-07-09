package id

import "github.com/sony/sonyflake"

var snowflake *sonyflake.Sonyflake

func init() {
	var snowflakeSetting sonyflake.Settings
	snowflake = sonyflake.NewSonyflake(snowflakeSetting)
}

func NextId() (uint64, error) {
	return snowflake.NextID()
}

func Extract(id uint64) map[string]uint64 {
	return sonyflake.Decompose(id)
}
