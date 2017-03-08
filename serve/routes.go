package serve

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/dimfeld/httptreemux"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

func (s *Service) setupRoutes() (http.Handler, error) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, ClientContextKey, s.Client)
	ctx = context.WithValue(ctx, ServiceContextKey, s)
	router := httptreemux.New()
	router.HeadCanUseGet = true
	router.DefaultContext = ctx
	ctxRoot := router.UsingContext()
	ctxRoot.Handler(mGET, "/", handler(root))
	ctxRoot.Handler(mGET, "/_all_dbs", handler(allDBs))
	ctxRoot.Handler(mGET, "/_log", handler(log))
	ctxRoot.Handler(mPUT, "/:db", handler(createDB))
	ctxRoot.Handler(mHEAD, "/:db", handler(dbExists))
	ctxRoot.Handler(mPOST, "/:db/_ensure_full_commit", handler(flush))
	ctxRoot.Handler(mGET, "/_config", handler(getConfig))
	ctxRoot.Handler(mGET, "/_config/:section", handler(getConfigSection))
	ctxRoot.Handler(mGET, "/_config/:section/:key", handler(getConfigItem))
	// ctxRoot.Handler(mDELETE, "/:db", handler(destroyDB) )
	// ctxRoot.Handler(http.MethodGet, "/:db", handler(getDB))

	handle := http.Handler(router)
	if s.Config().GetBool("httpd", "enable_compression") {
		level := s.Config().GetInt("httpd", "compression_level")
		if level == 0 {
			level = 8
		}
		gzipHandler, err := gziphandler.NewGzipLevelHandler(int(level))
		if err != nil {
			return nil, errors.Wrapf(err, "invalid httpd.compression_level '%s'", level)
		}
		s.Info("Enabling HTTPD cmpression, level %d", level)
		handle = gzipHandler(handle)
	}
	handle = requestLogger(s, handle)
	return handle, nil
}
