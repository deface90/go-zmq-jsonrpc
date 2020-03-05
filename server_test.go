package zmqjsonrpc

import (
    zmq "github.com/pebbe/zmq4"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "io/ioutil"
    "log"
    "net/rpc"
    "strings"
    "testing"
)

type Obj int
func (o *Obj) Test(args int, reply *int) error {
    *reply = 1000
    return nil
}

func TestServer_Create(t *testing.T) {
    server := &Server{Port: "1234"}
    err := server.Create()
    require.NoError(t, err)

    serverErr := &Server{Port: "1234"}
    err = serverErr.Create()
    assert.Error(t, err)

    server.Close()
}

func TestServer_ObtainRequest(t *testing.T) {
    var obj Obj
    err := rpc.Register(&obj)
    assert.NoError(t, err)

    server := &Server{}
    err = server.Create()
    assert.NoError(t, err)

    s := strings.NewReader("{\"id\":10, \"method\":\"Obj.Test\", \"params\":[]}")
    answer := server.obtainRequest(s)

    response, err := ioutil.ReadAll(answer)
    assert.NoError(t, err)
    assert.Equal(t, "{\"id\":10,\"result\":1000,\"error\":null}\n", string(response))

    s1 := strings.NewReader("{\"id\":20, \"method\":\"Obj.MissingTest\", \"params\":[]}")
    answer1 := server.obtainRequest(s1)

    response1, err := ioutil.ReadAll(answer1)
    assert.NoError(t, err)
    assert.Equal(t, "{\"id\":20,\"result\":null,\"error\":\"rpc: can't find method Obj.MissingTest\"}\n", string(response1))

    s2 := strings.NewReader("{\"id\":30, \"unmethod\":\"Obj.BadRequest\", \"params\":[]}")
    answer2 := server.obtainRequest(s2)

    response2, err := ioutil.ReadAll(answer2)
    assert.NoError(t, err)
    assert.Equal(t, "{\"id\":30,\"result\":null,\"error\":\"rpc: service/method request ill-formed: \"}\n", string(response2))
}

func TestServer_Serve(t *testing.T) {
    obj := new(Obj)
    err := rpc.Register(obj)
    assert.NoError(t, err)

    server := &Server{}
    err = server.Create()
    require.NoError(t, err)
    defer server.Close()

    go server.Serve()

    socket, err := getZMQSocket("5555", zmq.REQ)
    assert.NoError(t, err)

    data := "{\"id\":10, \"method\":\"Obj.Test\", \"params\":[]}"
    _, err = socket.Send(data, 0)
    assert.NoError(t, err)

    resp, err := socket.Recv(0)
    assert.NoError(t, err)
    assert.Equal(t, "{\"id\":10,\"result\":1000,\"error\":null}\n", resp)

    dataErr := "{\"id\":20, \"method\":\"Obj.BadJSON\", \"params\":[]"
    _, err = socket.Send(dataErr, 0)
    assert.NoError(t, err)

    resp1, err := socket.Recv(0)
    assert.NoError(t, err)
    assert.Equal(t, "", resp1)
}

func getZMQSocket(port string, socketType zmq.Type) (socket *zmq.Socket, err error) {
    socket, err = zmq.NewSocket(socketType)
    if err != nil {
        return nil, err
    }

    if socketType == zmq.REP {
        err = socket.Bind("tcp://*:" + port)
    } else {
        err = socket.Connect("tcp://127.0.0.1:" + port)
    }
    if err != nil {
        log.Printf("[ERROR] Failed to connect to zmq socket")
        return nil, err
    }

    return socket, nil
}
