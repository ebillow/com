package mod

import (
	"context"
	"google.golang.org/protobuf/proto"
	"server/com/util"
	"sync"
	"time"
)

type IMod interface {
	OnProto(message proto.Message, ctxWithValue context.Context)
	Save()
}

type param struct {
	msg          proto.Message
	ctxWithValue context.Context
}

type Mod struct {
	c    chan param
	opt  *Options
	iMod IMod
}

func (m *Mod) run(ctx context.Context) {
	tSave := time.NewTicker(m.opt.SaveSpan)
	for {
		select {
		case p := <-m.c:
			m.iMod.OnProto(p.msg, p.ctxWithValue)
		case <-tSave.C:
			m.iMod.Save()
		case <-ctx.Done():
			m.iMod.Save()
			wg.Done()
			return
		}
	}
}

func (m *Mod) runSafe(ctx context.Context) {
	tSave := time.NewTicker(m.opt.SaveSpan)
	for {
		select {
		case p := <-m.c:
			util.RunSafe(func() {
				m.iMod.OnProto(p.msg, p.ctxWithValue)
			})
		case <-tSave.C:
			util.RunSafe(func() {
				m.iMod.Save()
			})
		case <-ctx.Done():
			util.RunSafe(func() {
				m.iMod.Save()
			})

			wg.Done()
			return
		}
	}
}

var (
	mods        = make(map[int32]*Mod)
	wg          = &sync.WaitGroup{}
	ctx, cancel = context.WithCancel(context.Background())
)

func Register(modId int32, iMod IMod, opts ...Option) error {
	opt := newOptions()
	for _, o := range opts {
		if err := o(opt); err != nil {
			return err
		}
	}
	m := &Mod{
		c:    make(chan param, opt.MsgCache),
		opt:  opt,
		iMod: iMod,
	}
	mods[modId] = m
	return nil
}

func Post(modId int32, message proto.Message, ctxWithValue context.Context) {
	mod, ok := mods[modId]
	if !ok {
		return
	}
	mod.c <- param{msg: message, ctxWithValue: ctxWithValue}
}

func Start() {
	for _, m := range mods {
		for i := 0; i < m.opt.RunCnt; i++ {
			wg.Add(1)
			if m.opt.SafeMode {
				go m.runSafe(ctx)
			} else {
				go m.run(ctx)
			}
		}
	}
}

func Stop() {
	cancel()
	wg.Wait()
}
