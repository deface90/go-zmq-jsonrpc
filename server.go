package zmqjsonrpc

import (
    "bytes"
    "encoding/json"
    zmq "github.com/pebbe/zmq4"
    "io"
    "io/ioutil"
    "log"
    "net/rpc/jsonrpc"
    "strings"
)

// Server reporesents a JSON-RPC server over ZMQ
type Server struct {
    socket  *zmq.Socket
    running bool

    Port string
}

// rpcRequest represents a JSON-RPC request
// rpcRequest implements io.ReadWriteCloser
type rpcRequest struct {
    r    io.Reader
    rw   io.ReadWriter
    done chan bool
}

func (s *Server) Create() (err error) {
    s.socket, err = zmq.NewSocket(zmq.REP)
    if err != nil {
        return err
    }

    if s.Port == "" {
        s.Port = "5555"
    }
    err = s.socket.Bind("tcp://*:" + s.Port)
    if err != nil {
        log.Printf("[ERROR] Failed to init server zmq socket: %v", err)
        return err
    }

    return nil
}

func (s *Server) Close() {
    s.running = false
    err := s.socket.Close()
    if err != nil {
        log.Printf("[WARNING] Failed to close server zmq socket: %v", err)
    }
}

func (s *Server) Serve() {
	var tmpBuf interface{}
	s.running = true
    for s.running {
        data, err := s.socket.Recv(0)
        if err != nil {
            log.Printf("[WARNING] Error while receiving zmq img message: %v", err)
            continue
        }

        req := strings.NewReader(data)
        err = json.NewDecoder(req).Decode(&tmpBuf)
		if err != nil {
			log.Printf("[ERROR] Failed to unmarshal JSON-RPC request: %v. rpcRequest was: %v", err, data)
		}

		_, err = req.Seek(0, io.SeekStart)
		if err != nil {
			log.Printf("[ERROR] Failed to seek request reader to start: %v", err)
		}

		response := s.obtainRequest(req)
		outStr, err := ioutil.ReadAll(response)
		if err != nil {
			log.Printf("[ERROR] Failed to read obtained request data from response reader interface: %v", err)
		}

		s.sendString(string(outStr))
    }
}

func (s *Server) obtainRequest(r io.Reader) io.Reader {
    var buf bytes.Buffer
    done := make(chan bool)
    req := &rpcRequest{r, &buf, done}
    resp := req.Call()

    return resp
}

func (s *Server) sendString(str string) {
	_, err := s.socket.Send(str, 0)
	if err != nil {
		log.Printf("[WARNING] Failed to send string to ZMQ socket: %v", err)
	}
}

func (r *rpcRequest) Read(p []byte) (n int, err error) {
    return r.r.Read(p)
}

func (r *rpcRequest) Write(p []byte) (n int, err error) {
    return r.rw.Write(p)
}

func (r *rpcRequest) Close() error {
    r.done <- true
    return nil
}

func (r *rpcRequest) Call() io.Reader {
    go jsonrpc.ServeConn(r)
    <-r.done
    return r.rw
}
