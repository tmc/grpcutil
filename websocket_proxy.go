package grpcutil

import "net/http"

type thisCodeHasMoved error

// This function is deprecated and moved to https://github.com/tmc/grpc-websocket-proxy
func WebsocketProxy(h http.Handler) thisCodeHasMoved {
	return nil
}
