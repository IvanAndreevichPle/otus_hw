package main

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})
}

// --- Тест: ошибка при подключении к несуществующему адресу
func TestTelnetClient_ConnectUnavailable(t *testing.T) {
	client := NewTelnetClient("127.0.0.1:0", time.Second, io.NopCloser(&bytes.Buffer{}), &bytes.Buffer{})
	err := client.Connect()
	require.Error(t, err)
}

// --- Тест: корректное закрытие без соединения
func TestTelnetClient_CloseWithoutConnect(t *testing.T) {
	client := NewTelnetClient("127.0.0.1:0", time.Second, io.NopCloser(&bytes.Buffer{}), &bytes.Buffer{})
	err := client.Close()
	require.NoError(t, err)
}

// --- Тест: ошибка Send без соединения
func TestTelnetClient_SendWithoutConnect(t *testing.T) {
	client := NewTelnetClient("127.0.0.1:0", time.Second, io.NopCloser(&bytes.Buffer{}), &bytes.Buffer{})
	err := client.Send()
	require.Error(t, err)
}

// --- Тест: ошибка Receive без соединения
func TestTelnetClient_ReceiveWithoutConnect(t *testing.T) {
	client := NewTelnetClient("127.0.0.1:0", time.Second, io.NopCloser(&bytes.Buffer{}), &bytes.Buffer{})
	err := client.Receive()
	require.Error(t, err)
}

// --- Тест: ошибка при чтении из in (Send)
type errReader struct{}

func (e *errReader) Read(p []byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (e *errReader) Close() error               { return nil }

func TestTelnetClient_SendWithError(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	defer func() {
		err := l.Close()
		require.NoError(t, err)
	}()

	done := make(chan struct{})
	go func() {
		conn, err := l.Accept()
		require.NoError(t, err)
		err = conn.Close()
		require.NoError(t, err)
		close(done)
	}()

	client := NewTelnetClient(l.Addr().String(), time.Second, &errReader{}, &bytes.Buffer{})
	require.NoError(t, client.Connect())
	defer func() {
		err := client.Close()
		require.NoError(t, err)
	}()

	err = client.Send()
	require.Error(t, err)
	<-done
}

// --- Тест: ошибка при записи в out (Receive)
type errWriter struct{}

func (e *errWriter) Write(p []byte) (int, error) { return 0, io.ErrClosedPipe }
func TestTelnetClient_ReceiveWithError(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	defer func() {
		err := l.Close()
		require.NoError(t, err)
	}()

	done := make(chan struct{})
	go func() {
		conn, err := l.Accept()
		require.NoError(t, err)
		n, err := conn.Write([]byte("x")) // отправляем 1 байт
		require.NoError(t, err)
		require.Equal(t, 1, n)
		err = conn.Close()
		require.NoError(t, err)
		close(done)
	}()

	client := NewTelnetClient(l.Addr().String(), time.Second, io.NopCloser(&bytes.Buffer{}), &errWriter{})
	require.NoError(t, client.Connect())
	defer func() {
		err := client.Close()
		require.NoError(t, err)
	}()
	err = client.Receive()
	require.Error(t, err)
	<-done
}

// --- Тест: таймаут соединения
func TestTelnetClient_ConnectTimeout(t *testing.T) {
	// Используем несуществующий IP, чтобы гарантировать таймаут
	client := NewTelnetClient("10.255.255.1:65000", time.Millisecond*100, io.NopCloser(&bytes.Buffer{}), &bytes.Buffer{})
	start := time.Now()
	err := client.Connect()
	elapsed := time.Since(start)
	require.Error(t, err)
	require.GreaterOrEqual(t, elapsed, time.Millisecond*100)
}
