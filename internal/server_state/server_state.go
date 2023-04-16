package server_state

import (
	"fmt"
	"github.com/glide-im/glide/internal/im_server"
	"github.com/glide-im/glide/pkg/logger"
	"io"
	"net/http"
)

type httpHandler struct {
	g *im_server.GatewayServer
}

func (h *httpHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			bytes := []byte(fmt.Sprintf("%v", err))
			logger.E("%v", err)
			_, _ = writer.Write(bytes)
		}
	}()
	if request.URL.Path == "/state" {
		writer.WriteHeader(200)
		state := h.g.GetState()
		metricsLog := printBeautifulState(&state)
		//bytes, err2 := json.Marshal(&state)
		//if err2 != nil {
		//	panic(err2)
		//}
		_, err := writer.Write([]byte(metricsLog))
		if err != nil {
			panic(err)
		}
	} else {
		writer.WriteHeader(http.StatusNotFound)
	}
}

func StartSrv(port int, server *im_server.GatewayServer) {
	a := fmt.Sprintf("0.0.0.0:%d", port)
	err := http.ListenAndServe(a, &httpHandler{g: server})
	if err != nil {
		return
	}
}

func ShowServerState(addr string) {
	url := fmt.Sprintf("http://%s/state", addr)
	fmt.Printf("get state from %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Printf("get state from %s failed, status code %d\n", url, resp.StatusCode)
		return
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	if len(bytes) <= 0 {
		fmt.Println("get state failed, empty body")
		return
	}
	fmt.Println(string(bytes))
}

func printBeautifulState(state *im_server.GatewayMetrics) string {
	return fmt.Sprintf("\nServerId:\t%s", state.ServerId) +
		fmt.Sprintf("\nAddr:\t\t%s", state.Addr) +
		fmt.Sprintf("\nPort:\t\t%d", state.Port) +
		fmt.Sprintf("\nStartAt:\t%s", state.StartAt.Format("2006-01-02 15:04:05")) +
		fmt.Sprintf("\nRunningHours:\t\t%.2f", state.RunningHours) +
		fmt.Sprintf("\n== metric ==") +
		fmt.Sprintf("\nOnlineClients:\t\t%d", state.Conn.ConnectionCounter.Count()) +
		fmt.Sprintf("\nOnlineTempClients:\t%d", state.Conn.OnlineTempCounter.Count()) +
		fmt.Sprintf("\nTempConnAliveMaxSec:\t%d", state.Conn.AliveTempH.Max()) +
		fmt.Sprintf("\nTempConnAliveMeanSec:\t%f", state.Conn.AliveTempH.Mean()) +
		fmt.Sprintf("\nOnlineTempClients:\t%d", state.Conn.OnlineTempCounter.Count()) +
		fmt.Sprintf("\nLoggedConnAliveMaxSec:\t%d", state.Conn.AliveLoggedH.Max()) +
		fmt.Sprintf("\nLoggedConnAliveMeanSec:\t%f", state.Conn.AliveLoggedH.Mean()) +
		fmt.Sprintf("\nOutMessages:\t\t%d", state.Message.OutCounter.Count()) +
		fmt.Sprintf("\nOutMessageFails:\t%d", state.Message.FailsCounter.Count()) +
		fmt.Sprintf("\nInMessages:\t\t%d", state.Message.InCounter.Count()) +
		fmt.Sprintf("\n")
}
