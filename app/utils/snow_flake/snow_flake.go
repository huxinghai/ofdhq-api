package snow_flake

import (
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/global/variable"
	"ofdhq-api/app/utils/snow_flake/snowflake_interf"
	"sync"
)

// 创建一个雪花算法生成器(生成工厂)
func CreateSnowflakeFactory() snowflake_interf.InterfaceSnowFlake {
	return &snowflake{
		timestamp: 0,
		machineId: variable.ConfigYml.GetInt64("SnowFlake.SnowFlakeMachineId"),
		sequence:  0,
	}
}

type snowflake struct {
	sync.Mutex
	timestamp int64
	machineId int64
	sequence  int64
}

// 生成分布式ID
func (s *snowflake) GetId() int64 {
	s.Lock()
	defer func() {
		s.Unlock()
	}()
	now := variable.NowTimeSH().UnixNano() / 1e6
	if s.timestamp == now {
		s.sequence = (s.sequence + 1) & consts.SequenceMask
		if s.sequence == 0 {
			for now <= s.timestamp {
				now = variable.NowTimeSH().UnixNano() / 1e6
			}
		}
	} else {
		s.sequence = 0
	}
	s.timestamp = now
	r := (now-consts.StartTimeStamp)<<consts.TimestampShift | (s.machineId << consts.MachineIdShift) | (s.sequence)
	return r
}
