package xraymanager

import (
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

type XrayManager struct {
	config   string
	xrayPath string

	mu *sync.Mutex

	cmd     *exec.Cmd
	running bool
	doneCh  chan error
}

func NewXrayManager(config string, xrayPath string) *XrayManager {
	return &XrayManager{
		config:   config,
		xrayPath: xrayPath,
		mu:       new(sync.Mutex),
	}
}

var ErrAlreadyRunning = errors.New("xray already running")

func (m *XrayManager) Start() error {
	log.Println("starting xray core")
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running == true {
		return ErrAlreadyRunning
	}

	cmd := exec.Command(m.xrayPath, "run")
	cmd.Stdin = strings.NewReader(m.config)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return err
	}

	m.running = true
	m.doneCh = make(chan error, 1)
	m.cmd = cmd

	go m.watch(cmd)

	return nil
}

func (m *XrayManager) watch(cmd *exec.Cmd) {
	err := cmd.Wait()
	m.doneCh <- err

	m.mu.Lock()
	defer m.mu.Unlock()

	m.running = false
	m.cmd = nil
}

func (m *XrayManager) Stop(ctx context.Context) error {
	log.Println("stopping xray core")
	m.mu.Lock()
	if !m.running || m.cmd == nil {
		m.mu.Unlock()
		return nil
	}

	done := m.doneCh
	cmd := m.cmd
	m.mu.Unlock()

	err := cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		return err
	}

	select {
	case err := <-done:
		if terminatedBySIGTERM(err) {
			return nil
		}
		return err
	case <-ctx.Done():
		log.Println("kill him!")

		cmd.Process.Kill()
		<-done

		return ctx.Err()
	}
}

func (m *XrayManager) UpdateConfig(newConfig string) error {
	log.Println("updating config xray core")
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = newConfig
	return nil
}

var XrayNotRunning = errors.New("xray not running")

func (m *XrayManager) Restart(ctx context.Context) error {
	log.Println("restarting xray core")
	if !m.running {
		return XrayNotRunning
	}

	err := m.Stop(ctx)
	if err != nil {
		return err
	}

	return m.Start()
}

func (m *XrayManager) GetCurrentConfig() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.config
}

func terminatedBySIGTERM(err error) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}
	ws, ok := exitErr.Sys().(syscall.WaitStatus)
	if !ok {
		return false
	}
	return ws.Signaled() && ws.Signal() == syscall.SIGTERM
}

func (m *XrayManager) IsRunning() bool {
	log.Println("checking IsRunning xray core")
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}
