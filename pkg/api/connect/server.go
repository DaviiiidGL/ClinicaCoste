package connectapi

import (
	"net/http"

	connect "connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	"ClinicaCoste/gen/go/proto/hospital/v1/hospitalv1connect"
)

// NewServer arma el dispatcher de Connect, registra la implementación del
// servicio en él y lo monta sobre un ServeMux estándar de net/http.
func NewServer(hospitalServer *HospitalServer) *http.ServeMux {
	server := connect.NewServer()
	hospitalv1connect.RegisterHospitalServiceHandler(server, hospitalServer)

	mux := http.NewServeMux()
	connecthttp.Mount(mux, server)
	return mux
}
