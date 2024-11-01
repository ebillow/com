package snet

import (
	"server/com/log"
	"server/com/snet/service"
	"server/pb"
	"time"
)

func connMonitor(endPoint string, lname string, lid uint32) {
	monitorEndPoint = endPoint
	monitorOnce.Do(func() {
		localName = lname
		localId = lid

		service.NeedService(svcMgrType)
		wait := make(chan uint32)
		Connect(svcMgrType, func(svcId uint32) {
			s := GetMonitorSession()
			if s == nil {
				log.Warnf("can not get monitor session")
				return
			}
			s.Send(pb.MsgIDS2S_SvcService, &pb.MsgServiceInfo{
				Op:          pb.MsgServiceInfo_OpReqAll,
				IsToMonitor: true,
			})
		}, wait)
		go func() { //模拟收到 monitor service update
			t := time.NewTicker(time.Second * 3)
			defer t.Stop()
			service.Update(svcMgrType, 1, monitorEndPoint, pb.MsgServiceInfo_OpUpdate)
			for {
				select {
				case <-t.C:
					service.Update(svcMgrType, 1, monitorEndPoint, pb.MsgServiceInfo_OpUpdate)
				}
			}
		}()

		wait <- 1
		log.Infof("%s %d connect to monitor", localName, localId)
	})
}

func GetMonitorSession() *Session {
	return GetSession(svcMgrType, 1)
}
