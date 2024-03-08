package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
)

const (
	host = "localhost"
	port = "23234"
)

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		ssh.AllocatePty(),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			func(next ssh.Handler) ssh.Handler {
				return func(sess ssh.Session) {

					// renderer := bubbletea.MakeRenderer(sess)

					if len(sess.Command()) > 0 {
						// commands or args provided pass it to mods
						mods := exec.Command("mods", sess.Command()...)
						mods.Stdout = sess
						mods.Stderr = sess.Stderr()

						mods.Run()

						_ = sess.Exit(1)
					}

					next(sess)
				}
			},
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	// Set up the model with the current session and styles.
	// We'll use the session to call wish.Command, which makes it compatible
	// with tea.Command.
	m := model{
		sess: s,
	}
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

type model struct {
	sess ssh.Session
	done bool
}

func (m model) Init() tea.Cmd {
	return nil
}

type cmdFinishedMsg struct{ err error }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.done {
		c := wish.Command(m.sess, "mods", m.sess.Command()...)
		cmd := tea.Exec(c, func(err error) tea.Msg {
			if err != nil {
				log.Error("shell finished", "error", err)
			}
			m.done = true
			return cmdFinishedMsg{err: err}
		})
		m.done = true
		return m, cmd
	}

	return m, tea.Quit
}

func (m model) View() string {
	return ""
}
