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

type StringObj string
func (s *StringObj) Test(str string, reply *string) error {
    *reply = "Hello, " + str
    return nil
}

type MultiObj int64
func (m *MultiObj) Test(numbers []int64, reply *int64) error {
    *reply = 1
    for _, v := range numbers {
        *reply *= v
    }

    return nil
}

func TestClient_Call(t *testing.T) {
    obj := new(ClientObj)
    err := rpc.Register(obj)
    assert.NoError(t, err)

    server := &Server{
        Port: "1237",
    }
    err = server.Create()
    require.NoError(t, err)
    defer server.Close()

    go server.Serve()

    client := &Client{
        Port:"1237",
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

func TestClient_String(t *testing.T) {
    obj := new(StringObj)
    err := rpc.Register(obj)
    assert.NoError(t, err)

    server := &Server{
        Port: "1238",
    }
    err = server.Create()
    require.NoError(t, err)
    defer server.Close()

    go server.Serve()

    client := &Client{
        Port:"1238",
    }
    err = client.Create()
    require.NoError(t, err)

    req := Request{
        ID:     "10",
        Method: "StringObj.Test",
        Params: []interface{}{"World"},
    }

    resp, err := client.Call(req)
    assert.NoError(t, err)
    assert.Equal(t, nil, resp.Error)
    assert.Equal(t, "Hello, World", resp.Result)

    err = client.Close()
    assert.NoError(t, err)
}

func TestClient_Multi(t *testing.T) {
    obj := new(MultiObj)
    err := rpc.Register(obj)
    assert.NoError(t, err)

    server := &Server{
        Port: "1239",
    }
    err = server.Create()
    require.NoError(t, err)
    defer server.Close()

    go server.Serve()

    client := &Client{
        Port:"1239",
    }
    err = client.Create()
    require.NoError(t, err)

    params := []int64{1, 2, 3}
    req := Request{
        ID:     "10",
        Method: "MultiObj.Test",
        Params: []interface{}{params},
    }

    resp, err := client.Call(req)
    assert.NoError(t, err)
    assert.Equal(t, nil, resp.Error)
    assert.Equal(t, 6.0, resp.Result)

    err = client.Close()
    assert.NoError(t, err)
}
