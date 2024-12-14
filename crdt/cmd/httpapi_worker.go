package main

import (
	"encoding/json"
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"net/http"
	"chadcrdt/apps/crdtnode"
	. "chadcrdt/internal/utils"
)

func factory_HttpApiWebWorker() gen.ProcessBehavior {
	return &HttpApiWebWorker{}
}

type HttpApiWebWorker struct {
	act.WebWorker
}

func (w *HttpApiWebWorker) Init(args ...any) error {
	w.Log().Info("started web worker process with args %v", args)
	return nil
}

func (w *HttpApiWebWorker) HandleGet(from gen.PID, writer http.ResponseWriter, request *http.Request) error {
    key := request.PathValue("id")
	w.Log().Info("got HTTP GET for key %s", key)
    val := Must1(w.Call(gen.Atom("crdtactor"), crdtnode.GetValueRequest{Key: key}))
    if _, ok := val.(crdtnode.KeyNotFound); ok {
        writer.Header().Set("Content-Type", "text/plain")
        writer.WriteHeader(404)
        writer.Write([]byte("Key not found"))
        return nil
    }
	writer.Header().Set("Content-Type", "application/json")
    json.NewEncoder(writer).Encode(val)
	return nil
}

func (w *HttpApiWebWorker) HandlePut(from gen.PID, writer http.ResponseWriter, request *http.Request) error {
    key := request.PathValue("id");
    var val string
    json.NewDecoder(request.Body).Decode(&val)
	w.Log().Info("got HTTP Put for key %s with value %s", key, val)
    Must(w.Send(gen.Atom("crdtactor"), crdtnode.NewClientRowMessage{Key: key, Value: val}))

	writer.WriteHeader(200)
	return nil
}
