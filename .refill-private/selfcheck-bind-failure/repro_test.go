package selfcheckbindfailure

import (
	"context"
	"net"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestSelfcheckReportsBindFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/tactile-review", "-selfcheck", "-addr="+listener.Addr().String())
	cmd.Dir = filepath.Join("..", "..")
	output, runErr := cmd.CombinedOutput()
	if runErr == nil {
		t.Fatalf("selfcheck exited successfully while its listener address was already in use")
	}
	if ctx.Err() != nil {
		t.Fatalf("selfcheck did not terminate after bind failure: %v", ctx.Err())
	}
	if len(output) == 0 {
		t.Fatalf("selfcheck failed without reporting the listener bind error")
	}
}
