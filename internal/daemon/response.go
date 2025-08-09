package daemon

import (
	"fmt"
	"net"
)

func writeError(conn net.Conn, format string, args ...any) {
	conn.Write(fmt.Appendf(nil, "ERROR: "+format+"\n", args...))
}

func writeResponse(conn net.Conn, format string, args ...any) {
	conn.Write(fmt.Appendf(nil, format+"\n", args...))
}
