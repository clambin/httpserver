package httpserver_test

import (
	"fmt"
	"github.com/clambin/httpserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestPrometheus_Run(t *testing.T) {
	p := &httpserver.Prometheus{}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := p.Run()
		require.NoError(t, err)
		wg.Done()
	}()

	assert.Eventually(t, func() bool {
		resp, err2 := http.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics", p.GetPort()))
		if err2 == nil {
			_ = resp.Body.Close()
		}
		return err2 == nil && resp.StatusCode == http.StatusOK
	}, time.Second, 10*time.Millisecond)

	err := p.Shutdown(5 * time.Second)
	require.NoError(t, err)
	wg.Wait()
}
