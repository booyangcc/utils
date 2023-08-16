package cmdutil

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

// CmdMeta cmd meta
type CmdMeta struct {
	JobID string
	Name  string
	Args  []string
}

// Cmd cmd.
type Cmd struct {
	log          *zap.Logger
	cmd          *exec.Cmd
	Cancel       context.CancelFunc
	mu           sync.Mutex
	Stdout       *bytes.Buffer
	Stderr       *bytes.Buffer
	IsSuccess    bool
	Canceled     bool
	hasCancelFun bool
}

// Option option.
type Option func(c *Cmd)

// WithDir with dir
func WithDir(dir string) Option {
	return func(c *Cmd) {
		c.cmd.Dir = dir
	}
}

// WithLog with log
func WithLog(log *zap.Logger) Option {
	return func(c *Cmd) {
		c.log = log
	}
}

// Runner runner
func Runner(m *CmdMeta, options ...Option) *Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	command := exec.CommandContext(ctx, m.Name, m.Args...)
	return getCommand(command, cancel, options...)
}

// RunnerWithCommand runnerWithCommand
func RunnerWithCommand(name string, args []string, options ...Option) *Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	command := exec.CommandContext(ctx, name, args...)
	return getCommand(command, cancel, options...)
}

// RunnerWithCommandStr runnerWithCommandStr
func RunnerWithCommandStr(cmdStr string, options ...Option) *Cmd {
	command := exec.Command("sh", "-c", cmdStr)
	return getCommand(command, nil, options...)
}

func getCommand(command *exec.Cmd, cancel context.CancelFunc, options ...Option) *Cmd {
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Stdout = &bytes.Buffer{}
	command.Stderr = &bytes.Buffer{}

	c := &Cmd{
		cmd:    command,
		mu:     sync.Mutex{},
		Cancel: cancel,
	}

	if cancel == nil {
		c.hasCancelFun = false
	}

	for _, option := range options {
		option(c)
	}

	return c
}

// Start start
func (c *Cmd) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.log != nil {
		c.log.Info(c.cmd.String())
	}

	err := c.cmd.Start()
	if err != nil {
		c.Stderr, _ = c.cmd.Stderr.(*bytes.Buffer)
		return err
	}

	return nil
}

// Wait wait
func (c *Cmd) Wait() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.cmd.Wait()
	if err != nil {
		c.Stderr, _ = c.cmd.Stderr.(*bytes.Buffer)
		return err
	}
	if c.Stderr == nil {
		c.IsSuccess = true
	}

	c.Stdout, _ = c.cmd.Stdout.(*bytes.Buffer)
	return nil
}

// StartAndWait startAndWait
func (c *Cmd) StartAndWait() error {
	err := c.Start()
	if err != nil {
		return err
	}

	err = c.Wait()
	if err != nil {
		return err
	}
	return nil
}

// Stop stop
func (c *Cmd) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.hasCancelFun {
		return errors.New("this command not have cancel func")
	}

	c.Canceled = true
	c.Cancel()
	// kill child as well
	if c.cmd.Process != nil {
		pgid, err := syscall.Getpgid(c.cmd.Process.Pid)
		if err == nil {
			_ = syscall.Kill(-pgid, syscall.SIGTERM)
		}
	}
	return nil
}

// ExecCommand execCommand
func ExecCommand(cmdStr string) (string, error) {
	c := exec.Command("sh", "-c", cmdStr)
	output, err := c.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), err
}
