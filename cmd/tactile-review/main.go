package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"tactile-review/internal/application"
	"tactile-review/internal/store"
	"tactile-review/internal/web"
	"time"
)

func address(v string) string {
	if v != "" {
		return v
	}
	if p := os.Getenv("PORT"); p != "" {
		return "127.0.0.1:" + p
	}
	return "127.0.0.1:19081"
}
func main() {
	addr := flag.String("addr", "", "监听地址")
	self := flag.Bool("selfcheck", false, "运行自检")
	flag.Parse()
	path := filepath.Join(os.TempDir(), "tactile-review.db")
	s, e := store.Open(path)
	if e != nil {
		panic(e)
	}
	defer s.Close()
	a := application.New(s)
	srv := &http.Server{Addr: address(*addr), Handler: web.New(a).Handler()}
	if *self {
		ln, e := net.Listen("tcp", srv.Addr)
		if e != nil {
			panic(e)
		}
		serveErr := make(chan error, 1)
		go func() { serveErr <- srv.Serve(ln) }()
		e = runSelfCheck(a)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		srv.Shutdown(ctx)
		cancel()
		if e != nil {
			panic(e)
		}
		if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
		return
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv.Shutdown(ctx)
}
func runSelfCheck(a *application.Service) error {
	c, e := a.Create("A区", "东门", "视障人士", "GB-2024", "designer", "measurer")
	if e != nil {
		return e
	}
	r := domainRevision()
	c, e = a.AddRevision(c.ID, c.Version, r, "selfcheck")
	if e != nil {
		return e
	}
	c, e = a.Check(c.ID, c.Version)
	if e != nil {
		return e
	}
	c, e = a.Review(c.ID, c.Version, "reviewer", "APPROVE", "")
	if e != nil {
		return e
	}
	c, e = a.Freeze(c.ID, c.Version)
	if e != nil {
		return e
	}
	cred, e := a.Issue(c.ID, c.Version, "publisher")
	if e != nil {
		return e
	}
	ok, status, e := a.Verify(c.ID, cred.VerificationCode)
	if e != nil || !ok || status != "VALID" {
		return fmt.Errorf("verify failed: %s", status)
	}
	fmt.Println("selfcheck ok")
	return nil
}
