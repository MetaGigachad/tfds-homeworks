package main

import (
	"net/http"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/meta"
    opt "chaddb/internal/options"
)

func factory_HttpApi() gen.ProcessBehavior {
	return &HttpApi{}
}

type HttpApi struct {
	act.Pool
}

// Init invoked on a start this process.
func (w *HttpApi) Init(args ...any) (act.PoolOptions, error) {
	var webOptions meta.WebServerOptions
	var poolOptions act.PoolOptions

	mux := http.NewServeMux()

	// create and spawn root handler meta-process.
	root := meta.CreateWebHandler(meta.WebHandlerOptions{})
	rootid, err := w.SpawnMeta(root, gen.MetaOptions{})
	if err != nil {
		w.Log().Error("unable to spawn WebHandler meta-process: %s", err)
		return poolOptions, err
	}

	mux.Handle("/{id}", root)
	w.Log().Info("started WebHandler to serve '/' (meta-process: %s)", rootid)

	webOptions.Port = uint16(opt.ApiPort)
	webOptions.Host = "localhost"

	webOptions.Handler = mux

	webserver, err := meta.CreateWebServer(webOptions)
	if err != nil {
		w.Log().Error("unable to create Web server meta-process: %s", err)
		return poolOptions, err
	}
	webserverid, err := w.SpawnMeta(webserver, gen.MetaOptions{})
	if err != nil {
		// invoke Terminate to close listening socket
		webserver.Terminate(err)
		return poolOptions, err
	}

	https := "http"
	if webOptions.CertManager != nil {
		https = "https"
	}
	w.Log().Info("started Web server %s: use %s://%s:%d/", webserverid, https, webOptions.Host, webOptions.Port)
	w.Log().Info("you may check it with command below:")
	w.Log().Info("   $ curl -k %s://%s:%d", https, webOptions.Host, webOptions.Port)

	poolOptions.WorkerFactory = factory_HttpApiWebWorker
	return poolOptions, nil
}
