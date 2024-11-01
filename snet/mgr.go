package snet

import (
	"server/com/log"
	"sync"
	"sync/atomic"
)

type SvcType struct {
	name     string
	ses      map[uint32]*Session
	sesSlice []*Session
	idx      int
}

func (svc *SvcType) rebuild() {
	svc.sesSlice = make([]*Session, 0, len(svc.ses))
	for _, v := range svc.ses {
		svc.sesSlice = append(svc.sesSlice, v)
	}
	svc.idx = 0
}

var (
	sesMgr = make(map[string]*SvcType)
	mtx    sync.Mutex
)

func AddSession(name string, id uint32, session *Session) {
	mtx.Lock()
	defer mtx.Unlock()

	svc, ok := sesMgr[name]
	if !ok {
		svc = &SvcType{
			name: name,
			ses:  make(map[uint32]*Session),
		}
		sesMgr[name] = svc
	} //else{
	//	old := svc.ses[id]
	//	if old != nil && old != session{
	//		old.Close()
	//	}
	//}
	svc.ses[id] = session
	svc.rebuild()
	log.Infof("add session %s[%s]", session.desc, session.sn)
}

func DelSession(name string, id uint32, sn string) {
	mtx.Lock()
	defer mtx.Unlock()

	svc, ok := sesMgr[name]
	if !ok {
		return
	}
	if s, ok := svc.ses[id]; ok && s.sn == sn {
		delete(svc.ses, id)
		svc.rebuild()
		log.Infof("delete session %s_%d[%s]", name, id, sn)
	}
}

func GetSession(name string, id uint32) *Session {
	mtx.Lock()
	defer mtx.Unlock()

	svc, ok := sesMgr[name]
	if !ok {
		return nil
	}
	return svc.ses[id]
}

func RandSession(name string) *Session {
	mtx.Lock()
	defer mtx.Unlock()

	svc, ok := sesMgr[name]
	if !ok || len(svc.sesSlice) == 0 {
		return nil
	}

	if svc.idx >= len(svc.sesSlice) {
		svc.idx = 0
	}
	s := svc.sesSlice[svc.idx]
	svc.idx++
	return s
}

func RandSessionHash(name string, selfID uint32) *Session {
	mtx.Lock()
	defer mtx.Unlock()

	svc, ok := sesMgr[name]
	if !ok || len(svc.sesSlice) == 0 {
		return nil
	}

	s := svc.sesSlice[int(selfID)%len(svc.sesSlice)]
	return s
}

func SvcSessionIDs(name string) (ids []uint32) {
	mtx.Lock()
	defer mtx.Unlock()

	svc, ok := sesMgr[name]
	if !ok {
		return nil
	}

	for k := range svc.ses {
		ids = append(ids, k)
	}
	return
}

func SvcSessions(name string) (sess []*Session) {
	mtx.Lock()
	defer mtx.Unlock()

	svc, ok := sesMgr[name]
	if !ok {
		return nil
	}

	for _, v := range svc.ses {
		sess = append(sess, v)
	}
	return
}

func AllSessions() (sess []*Session) {
	mtx.Lock()
	defer mtx.Unlock()

	for _, svc := range sesMgr {
		for _, v := range svc.ses {
			sess = append(sess, v)
		}
	}
	return
}

// **-----------------------------*/
var traceProto int32

func IsTraceProto() bool {
	return atomic.LoadInt32(&traceProto) == 1
}

func SetTraceProto(v bool) {
	if v {
		atomic.StoreInt32(&traceProto, 1)
	} else {
		atomic.StoreInt32(&traceProto, 0)
	}
}
