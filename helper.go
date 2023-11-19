package epy

import (
	"sync"
	"os"
	"time"
)

type pyCtx struct {
	slx *XStarlark
	mt   time.Time
}

var (
	pyCtxCache map[string]*pyCtx
	lock *sync.Mutex
)

func InitPyCache() {
	if lock != nil {
		return
	}
	lock = &sync.Mutex{}
	pyCtxCache = make(map[string]*pyCtx)
}

func LoadFileFromCache(path string, vars map[string]interface{}) (ctx *XStarlark, existing bool, err error) {
	lock.Lock()
	defer lock.Unlock()

	pyC, ok := pyCtxCache[path]

	if !ok {
		ctx = New()
		if err = ctx.LoadFile(path, vars); err != nil {
			return
		}
		fi, _ := os.Stat(path)
		pyC = &pyCtx{
			slx: ctx,
			mt: fi.ModTime(),
		}
		pyCtxCache[path] = pyC
		return
	}

	fi, e := os.Stat(path)
	if e != nil {
		err = e
		return
	}
	mt := fi.ModTime()
	if !pyC.mt.Equal(mt) {
		ctx = New()
		if err = ctx.LoadFile(path, vars); err != nil {
			return
		}
		pyC.slx = ctx
		pyC.mt = mt
	} else {
		existing = true
		ctx = pyC.slx
	}
	return
}
