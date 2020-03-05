package zmqjsonrpc

import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "net/rpc"
    "testing"
)

type Args struct {
    A, B float64
}

type ClientObj int
func (o *ClientObj) Test(args *Args, reply *float64) error {
    *reply = args.A + args.B
    return nil
}

func TestClient_Call(t *testing.T) {
    obj := new(ClientObj)
    err := rpc.Register(obj)
    assert.NoError(t, err)

    server := &Server{
        Port: "1234",
    }
    err = server.Create()
    require.NoError(t, err)
    defer server.Close()

    go server.Serve()

    client := &Client{
        Port:"1234",
    }
    err = client.Create()
    require.NoError(t, err)

    params := &Args{10, 20}
    req := Request{
        ID:     "10",
        Method: "ClientObj.Test",
        Params: []interface{}{params},
    }

    resp, err := client.Call(req)
    assert.NoError(t, err)
    assert.Equal(t, nil, resp.Error)
    assert.Equal(t, 30.0, resp.Result)

    err = client.Close()
    assert.NoError(t, err)
}
