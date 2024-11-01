package service

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"server/com/log"
	"server/pb"
	"sync"
	"time"
)

const (
	MaxMsgSize = 1024 * 1024 * 128
)

type Client struct {
	conn     *grpc.ClientConn //
	endPoint string
}

func (c *Client) isActive() bool {
	return c.conn.GetState() == connectivity.Ready
}

type callback struct {
	cb   func(id uint32)
	wait chan uint32
}

type Mgr struct {
	services      sync.Map
	dialCallbacks sync.Map
	needServices  sync.Map
}

// NewServiceMgr	创建新的服务管理器
func NewServiceMgr() *Mgr {
	m := &Mgr{}
	return m
}

func (m *Mgr) isNeed(svcType string) bool {
	_, ok := m.needServices.Load(svcType)
	return ok
}

func makeKey(svcType string, id uint32) string {
	return fmt.Sprintf("%s_%d", svcType, id)
}

// Add	添加服务
func (m *Mgr) Update(svcType string, id uint32, endPoint string) {
	if !m.isNeed(svcType) {
		return
	}
	//log.Infof("recv svc update %s %d %s", svcType, id, endPoint)
	go func() {
		success := false
		key := makeKey(svcType, id)
		iSrv, ok := m.services.Load(key)
		if !ok {
			conn, ok := createConn(endPoint, key)
			if ok {
				m.services.Store(key, &Client{
					conn:     conn,
					endPoint: endPoint,
				})
				success = true
			}
		} else {
			cli := iSrv.(*Client)
			if !cli.isActive() || cli.endPoint != endPoint {
				_ = cli.conn.Close() //只有覆盖新增，没有修改，不用锁
				conn, ok := createConn(endPoint, key)
				if ok {
					m.services.Store(key, &Client{
						conn:     conn,
						endPoint: endPoint,
					})
					success = true
				}
			}
		}

		if success {
			m.onDialSuccess(svcType, id)
		}
	}()
}

func (m *Mgr) onDialSuccess(svcType string, id uint32) {
	if v, ok := m.dialCallbacks.Load(svcType); ok {
		cb := v.(*callback)
		if cb.cb != nil {
			cb.cb(id)
		}
		if cb.wait != nil {
			<-cb.wait
			log.Infof("%s %d wait end", svcType, id)
		}
	}
}

func (m *Mgr) Del(svcType string, id uint32) {
	if !m.isNeed(svcType) {
		return
	}

	key := makeKey(svcType, id)
	iSrv, ok := m.services.Load(key)
	if !ok {
		return
	} else {
		cli := iSrv.(*Client)
		cli.conn.Close()
		log.Infof("delete svc %s %s", key, cli.endPoint)
		m.services.Delete(id)
	}
}

func createConn(endPoint string, key string) (c *grpc.ClientConn, ok bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	conn, err := grpc.DialContext(ctx, endPoint, grpc.WithBlock(), grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxMsgSize)))
	if err != nil {
		log.Errorf("create rpc conn to %s err:%v", key, err)
		return nil, false
	}

	log.Infof("add service=%s addr=[%s]", key, endPoint)
	return conn, true
}

// Get	获取服务
func (m *Mgr) Get(svcType string, id uint32) *grpc.ClientConn {
	key := makeKey(svcType, id)
	iSrv, ok := m.services.Load(key)
	if !ok {
		return nil
	} else {
		cli := iSrv.(*Client)
		return cli.conn
	}
}

func (m *Mgr) RegisterDialCallback(svcType string, cb func(id uint32), wait chan uint32) {
	m.dialCallbacks.Store(svcType, &callback{cb: cb, wait: wait})
}

var (
	service = NewServiceMgr()
)

// NeedService 设置需要的服务
func NeedService(services ...string) {
	for i, v := range services {
		service.needServices.Store(v, i)
	}
}

// Update	更新服务
func Update(svcType string, id uint32, endPoint string, op pb.MsgServiceInfo_Operator) {
	switch op {
	case pb.MsgServiceInfo_OpDel:
		service.Del(svcType, id)
	default:
		service.Update(svcType, id, endPoint)
	}
}

// Get	获取服务
func Get(svcType string, id uint32) *grpc.ClientConn {
	return service.Get(svcType, id)
}

func RegisterDialSvcCallback(svcType string, connCallback func(svcId uint32), wait chan uint32) {
	service.RegisterDialCallback(svcType, connCallback, wait)
}
