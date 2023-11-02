package httpserver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"github.com/wencan/gox/async"
)

// Server http.Server的包装。
// 支持收到SIGTERM、SIGINT信号时结束监听。
// 支持等待全部处理过程结束。
type Server struct {
	srv *http.Server

	listener net.Listener

	graceful *async.Graceful

	serveError error

	stopFlag chan interface{}

	stopped uint32

	shutdownNotify chan interface{}
}

// NewServer 创建一个新的服务。
func NewServer(ctx context.Context, srv *http.Server) *Server {
	s := &Server{
		srv:            srv,
		graceful:       async.DefaultGraceful.NewBranch("http_server_" + srv.Addr),
		stopFlag:       make(chan interface{}),
		shutdownNotify: make(chan interface{}),
	}

	next := s.srv.Handler
	s.srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 如果要给每个路由，加上各自的graceful，应该在路由层实现
		s.graceful.Run(func() {
			next.ServeHTTP(w, r)
		})
	})

	return s
}

// Start 开始服务。
// 服务直至收到SIGTERM、SIGINT信号，或Stop被调用。
// 开始后返回。
func (s *Server) Start(ctx context.Context) (listenAddr string, err error) {
	s.listener, err = net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return "", err
	}

	go s.graceful.NewBranch("http_server_start_waitexit").Run(func() {
		sigExit := make(chan os.Signal, 1)
		signal.Notify(sigExit, os.Interrupt, syscall.SIGTERM)

		select {
		case <-sigExit:
		case <-s.stopFlag:
		}

		close(s.shutdownNotify)
		s.srv.Shutdown(context.Background())
	})

	go s.graceful.NewBranch("http_server_start_serve").Run(func() {
		defer s.Stop(context.Background())

		err := s.srv.Serve(s.listener)
		if err != nil && err != http.ErrServerClosed {
			s.serveError = err
		}
	})

	return s.listener.Addr().String(), nil
}

// StartTLS 开始TLS服务。开始后返回。
func (s *Server) StartTLS(ctx context.Context, certFile, keyFile string) (listenAddr string, err error) {
	s.listener, err = net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return "", err
	}

	go s.graceful.NewBranch("http_server_starttls_waitexit").Run(func() {
		sigExit := make(chan os.Signal, 1)
		signal.Notify(sigExit, os.Interrupt, syscall.SIGTERM)

		select {
		case <-sigExit:
		case <-s.stopFlag:
		}

		s.srv.Shutdown(context.Background())
	})

	go s.graceful.NewBranch("http_server_starttls_servetls").Run(func() {
		defer s.Stop(context.Background())

		err := s.srv.ServeTLS(s.listener, certFile, keyFile)
		if err != nil && err != http.ErrServerClosed {
			s.serveError = err
		}
	})

	return s.listener.Addr().String(), nil
}

// Stop 停止服务。
func (s *Server) Stop(ctx context.Context) {
	if atomic.AddUint32(&s.stopped, 1) == 1 {
		close(s.stopFlag)
	}
}

// Wait 等待服务内全部后台过程结束。如果等待超时，返回错误包含还未退出的过程信息。
func (s *Server) Wait(ctx context.Context) error {
	// 先等待Stop被调用
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.stopFlag:
	}

	err := s.graceful.Wait(ctx)
	if err != nil {
		return fmt.Errorf("%w. busy branches: %v", err, s.graceful.BusyBranches())
	}
	return nil
}

// ShutdownNotify 服务关闭通知。
func (s *Server) ShutdownNotify() <-chan interface{} {
	return s.shutdownNotify
}
