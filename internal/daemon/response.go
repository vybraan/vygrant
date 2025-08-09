package daemon

import (
	"fmt"
	"net"
)

func writeError(conn net.Conn, format string, args ...interface{}) {
	conn.Write([]byte(fmt.Sprintf("ERROR: "+format+"\n", args...)))
}

func writeResponse(conn net.Conn, format string, args ...interface{}) {
	conn.Write([]byte(fmt.Sprintf(format+"\n", args...)))
}
