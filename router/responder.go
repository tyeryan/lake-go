package router

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"net/http"
)

const responseCodeContextKey = "responseCode"

func JSON(w http.ResponseWriter, r *http.Request, v interface{}) {
	enc := protojson.MarshalOptions{
		UseEnumNumbers:  false,
		EmitUnpopulated: false,
	}

	pb := v.(proto.Message)
	buf, err := enc.Marshal(pb)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if status, ok := r.Context().Value(responseCodeContextKey).(int); ok {
		w.WriteHeader(status)
	}
	w.Write(buf)
}
