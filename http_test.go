package httpserver

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestHttpServer_Run(t *testing.T) {
	s := httpServer{}
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err2 := s.Run(0, h)
		require.NoError(t, err2)
		wg.Done()
	}()

	assert.Eventually(t, func() bool {
		resp, err2 := http.Get(fmt.Sprintf("http://127.0.0.1:%d", s.GetPort()))
		if err2 == nil {
			_ = resp.Body.Close()
		}
		return err2 == nil && resp.StatusCode == http.StatusOK
	}, time.Second, 10*time.Millisecond)

	assert.NotZero(t, s.GetPort())
	err := s.Shutdown(time.Minute)
	require.NoError(t, err)
	wg.Wait()
}

func TestHttpServer_Run_BadPort(t *testing.T) {
	s := httpServer{}
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})

	err := s.Run(-1, h)
	assert.Error(t, err)
}

func TestHttpServer_Duplicate_Port(t *testing.T) {
	s1 := httpServer{}
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})
	go func() {
		_ = s1.Run(8889, h)
	}()

	assert.Eventually(t, func() bool {
		resp, err := http.Get("http://127.0.0.1:8889/")
		if err != nil {
			return false
		}
		_ = resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, time.Second, 100*time.Millisecond)

	s2 := httpServer{}
	err := s2.Run(8889, h)
	assert.Error(t, err)

	_ = s1.Shutdown(time.Minute)
}
