package main

import (
	"encoding/json"
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"net/http"
	"chaddb/apps/dbnode"
	. "chaddb/internal/utils"
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
    val := Must1(w.Call(gen.Atom("storageactor"), dbnode.StorageGet{Key: key}))
    if _, ok := val.(dbnode.KeyNotFound); ok {
        writer.Header().Set("Content-Type", "text/plain")
        writer.WriteHeader(404)
        writer.Write([]byte("Key not found"))
        return nil
    }
	writer.Header().Set("Content-Type", "application/json")
    json.NewEncoder(writer).Encode(val)
	return nil
}

func (w *HttpApiWebWorker) HandlePost(from gen.PID, writer http.ResponseWriter, request *http.Request) error {
    key := request.PathValue("id");
    var val string
    json.NewDecoder(request.Body).Decode(&val)
	w.Log().Info("got HTTP Post for key %s with value %s", key, val)
    // Must1(w.Call(gen.Atom("storageactor"), dbnode.StorageSet{Key: key, Value: val}))
    Must1(w.CallWithTimeout(gen.Atom("raftactor"), dbnode.AddEntry{Key: key, Value: val, Tombstone: false}, 10 * 1000))

	writer.WriteHeader(200)
	return nil
}

func (w *HttpApiWebWorker) HandleDelete(from gen.PID, writer http.ResponseWriter, request *http.Request) error {
    key := request.PathValue("id");
	w.Log().Info("got HTTP Delete for key %s", key)
    // Must1(w.Call(gen.Atom("storageactor"), dbnode.StorageDel{Key: key}));
    Must1(w.CallWithTimeout(gen.Atom("raftactor"), dbnode.AddEntry{Key: key, Tombstone: true}, 1000))
	writer.WriteHeader(200)
	return nil
}
