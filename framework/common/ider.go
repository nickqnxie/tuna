/*
@Time : 2019/10/16 19:49
@Author : nickqnxie
@File : ider.go
*/

package common

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/sony/sonyflake"
	"hash/crc32"
	"time"
)

func NewId() (id uint64, err error) {
	st := sonyflake.Settings{}

	st.CheckMachineID = func(u uint16) bool {
		return true
	}

	st.MachineID = func() (uint16, error) {
		crc32q := crc32.MakeTable(0xD5828281)
		hash32 := crc32.Checksum([]byte(fmt.Sprintf("%s",
			fmt.Sprintf("%d", time.Now().UnixNano()))), crc32q)
		return uint16(hash32), nil
	}

	flake := sonyflake.NewSonyflake(st)
	id, err = flake.NextID()
	if err != nil {
		logrus.Fatalf("flake.NextID() failed with %s\n", err)
	}

	return
}
