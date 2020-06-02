package id

import (
	"github.com/sony/sonyflake"
	"math/rand"
	"time"
)

var snowflake *sonyflake.Sonyflake
var snowflakeSetting sonyflake.Settings

func init() {
	rand.Seed(time.Now().UnixNano())
	snowflakeSetting.MachineID = func() (machineId uint16, e error) {
		return uint16(rand.Uint64()), nil
	}
}

func NextID() (uint64, error) {
	snowflake = sonyflake.NewSonyflake(snowflakeSetting)
	return snowflake.NextID()
}

func Extract(id uint64) map[string]uint64 {
	return sonyflake.Decompose(id)
}
