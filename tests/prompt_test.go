package tests

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2/terminal"
	expect "github.com/Netflix/go-expect"
	pseudotty "github.com/creack/pty"
	"github.com/hinshun/vt10x"
	"log"
	"os"
	"os/exec"
	"testing"
)

type expectConsole interface {
	ExpectString(string)
	ExpectEOF()
	SendLine(string)
	Send(string)
	Console() *expect.Console
}

type consoleWithErrorHandling struct {
	console *expect.Console
	t       *testing.T
}

func (c *consoleWithErrorHandling) ExpectString(s string) {
	if _, err := c.console.ExpectString(s); err != nil {
		c.t.Helper()
		c.t.Fatalf("ExpectString(%q) = %v", s, err)
	}
}

func (c *consoleWithErrorHandling) SendLine(s string) {
	if _, err := c.console.SendLine(s); err != nil {
		c.t.Helper()
		c.t.Fatalf("SendLine(%q) = %v", s, err)
	}
}

func (c *consoleWithErrorHandling) Send(s string) {
	if _, err := c.console.Send(s); err != nil {
		c.t.Helper()
		c.t.Fatalf("Send(%q) = %v", s, err)
	}
}

func (c *consoleWithErrorHandling) ExpectEOF() {
	if _, err := c.console.ExpectEOF(); err != nil {
		c.t.Helper()
		c.t.Fatalf("ExpectEOF() = %v", err)
	}
}

func (c *consoleWithErrorHandling) Console() *expect.Console {
	return c.console
}

func RunTest(t *testing.T, procedure func(expectConsole), test func(terminal.Stdio) error) {
	t.Helper()
	t.Parallel()

	pty, tty, err := pseudotty.Open()
	if err != nil {
		t.Fatalf("failed to open pseudotty: %v", err)
	}

	term := vt10x.New(vt10x.WithWriter(tty))
	c, err := expect.NewConsole(expect.WithStdin(pty), expect.WithStdout(term), expect.WithCloser(pty, tty))
	if err != nil {
		t.Fatalf("failed to create console: %v", err)
	}
	defer c.Close()

	donec := make(chan struct{})
	go func() {
		defer close(donec)
		procedure(&consoleWithErrorHandling{console: c, t: t})
	}()

	stdio := terminal.Stdio{In: c.Tty(), Out: c.Tty(), Err: c.Tty()}
	if err := test(stdio); err != nil {
		t.Error(err)
	}

	if err := c.Tty().Close(); err != nil {
		t.Errorf("error closing Tty: %v", err)
	}
	<-donec
}

func TestPrompt(t *testing.T) {
	c, err := expect.NewConsole(expect.WithStdout(os.Stdout))
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	cmd := exec.Command("./togomak_coverage", "-C", "../examples/prompt", "--ci=false")

	RunTest(t, func(c expectConsole) {
		c.SendLine("Shinji Ikari\n\n")
		fmt.Println(c.Console().ExpectEOF())
	}, func(stdio terminal.Stdio) error {
		cmd.Stdin = stdio.In
		cmd.Stdout = stdio.Out
		cmd.Stderr = stdio.Out
		cmd.Env = append(os.Environ(), "QUIT_IF_NOT_SHINJI=true", fmt.Sprintf("GOCOVERDIR=%s", os.Getenv("PROMPT_GOCOVERDIR")))
		return cmd.Start()

	})

	err = cmd.Wait()
	if err != nil {
		d, err := os.ReadFile("/tmp/quit")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(d))
		t.Error(err)
	}
	fmt.Println(cmd.ProcessState.ExitCode())

}

func TestInterrupt(t *testing.T) {

}
