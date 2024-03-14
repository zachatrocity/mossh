package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	_ "github.com/joho/godotenv/autoload"
)

var (
	host = "localhost"
	port = "23234"
)

func main() {
	authHandler := wish.WithPublicKeyAuth(func(ssh.Context, ssh.PublicKey) bool {
		log.Warn("MOSSH_ALLOW_LIST not set, allowing all public keys")
		return true
	})

	if os.Getenv("MOSSH_ALLOW_LIST") != "" {
		authHandler = wish.WithAuthorizedKeys(os.Getenv("MOSSH_ALLOW_LIST"))
	}

	if os.Getenv("MOSSH_PORT") != "" {
		port = os.Getenv("MOSSH_PORT")
	}

	if os.Getenv("MOSSH_HOST") != "" {
		host = os.Getenv("MOSSH_HOST")
	}

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		authHandler,
		ssh.AllocatePty(),
		wish.WithMiddleware(
			bubbletea.Middleware(func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
				return model{
					sess: s,
				}, []tea.ProgramOption{tea.WithAltScreen()}
			}),
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

type model struct {
	sess ssh.Session
	done bool
}

func (m model) Init() tea.Cmd {
	// return nil
	c := wish.Command(m.sess, "mods", m.sess.Command()...)
	cmd := tea.Exec(c, func(err error) tea.Msg {
		if err != nil {
			log.Error("shell finished", "error", err)
		}
		m.done = true
		return err
	})
	return cmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// we're immediately executing mods with the commands passed
	// from ssh, we can do that in the Init() function. After that runs
	// we can immediately quit.
	// I feel like we shouldn't need bubbletea at all but right we use it just
	// for it's renderer
	return m, tea.Quit
}

func (m model) View() string {
	return ""
}
