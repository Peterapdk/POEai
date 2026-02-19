package ssh

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	host string
	user string
	key  string
}

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func New(host, user, keyPath string) *Client {
	return &Client{host: host, user: user, key: keyPath}
}

func (c *Client) Exec(ctx context.Context, cmd string) (Result, error) {
	key, err := os.ReadFile(c.key)
	if err != nil {
		return Result{}, fmt.Errorf("read key: %w", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return Result{}, fmt.Errorf("parse key: %w", err)
	}
	cfg := &ssh.ClientConfig{
		User:            c.user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: support known_hosts
	}

	// Dial with context
	conn, err := ssh.Dial("tcp", c.host+":22", cfg)
	if err != nil {
		return Result{}, fmt.Errorf("dial %s: %w", c.host, err)
	}
	defer conn.Close()

	sess, err := conn.NewSession()
	if err != nil {
		return Result{}, fmt.Errorf("session: %w", err)
	}
	defer sess.Close()

	var stdout, stderr bytes.Buffer
	sess.Stdout = &stdout
	sess.Stderr = &stderr

	// Handle context cancellation
	done := make(chan error, 1)
	go func() {
		done <- sess.Run(cmd)
	}()

	select {
	case <-ctx.Done():
		sess.Signal(ssh.SIGKILL)
		return Result{}, ctx.Err()
	case err := <-done:
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*ssh.ExitError); ok {
				exitCode = exitErr.ExitStatus()
			} else {
				return Result{}, fmt.Errorf("run: %w", err)
			}
		}
		return Result{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: exitCode}, nil
	}
}
