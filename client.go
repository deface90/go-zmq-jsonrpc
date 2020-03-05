package zmqjsonrpc

import (
    "encoding/json"
    "fmt"
    zmq "github.com/pebbe/zmq4"
    "time"
)

type Client struct {
    socket *zmq.Socket

    Host string
    Port string

    ConnTimeout time.Duration
    RecvTimeout time.Duration
}

type Request struct {
    ID     string        `json:"id"`
    Method string        `json:"method"`
    Params []interface{} `json:"params"`
}

type Response struct {
    ID     string      `json:"ID"`
    Result interface{} `json:"result"`
    Error  interface{} `json:"error"`
}

func (c *Client) Create() (err error) {
    c.socket, err = zmq.NewSocket(zmq.REQ)
    if err != nil {
        return err
    }

    if c.ConnTimeout != 0 {
        _ = c.socket.SetConnectTimeout(c.ConnTimeout)
    }
    if c.RecvTimeout != 0 {
        _ = c.socket.SetRcvtimeo(c.RecvTimeout)
    }

    if c.Host == "" {
        c.Host = "127.0.0.1"
    }
    if c.Port == "" {
        c.Port = "5555"
    }
    err = c.socket.Connect("tcp://" + c.Host + ":" + c.Port)
    if err != nil {
        return err
    }

    return nil
}

func (c *Client) Call(req Request) (resp *Response, err error) {
    resp = &Response{
        Error: nil,
    }
    jsonData, err := json.Marshal(req)
    if err != nil {
        resp.Error = fmt.Sprintf("zmqjsonrpc client: failed to marshal request: %v", err)
        return resp, err
    }

    _, err = c.socket.Send(string(jsonData), 0)
    if err != nil {
        resp.Error = fmt.Sprintf("zmqjsonrpc client: failed to send request to rpc server over zmq: %v", err)
        return nil, err
    }

    respData, err := c.socket.Recv(0)
    if err != nil {
        resp.Error = fmt.Sprintf("zmqjsonrpc client: failed to receive data from rpc server over zmq: %v", err)
        return nil, err
    }

    err = json.Unmarshal([]byte(respData), &resp)
    if err != nil {
        resp.Error = fmt.Sprintf("zmqjsonrpc client: failed to unmarshal rpc server responce: %v", err)
        return nil, err
    }

    return resp, err
}

func (c *Client) Close() error {
    return c.socket.Close()
}
