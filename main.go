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
	"github.com/charmbracelet/lipgloss"
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

		// Allocate a pty.
		// This creates a pseudoconsole on windows, compatibility is limited in
		// that case, see the open issues for more details.
		ssh.AllocatePty(),
		wish.WithMiddleware(
			// run our Bubble Tea handler
			// bubbletea.Middleware(teaHandler),

			// ensure the user has requested a tty
			activeterm.Middleware(),
			wish.Middleware(
				func(next ssh.Handler) ssh.Handler {
					return func(sess ssh.Session) {
						cmd := sess.Command()

						if len(cmd) >= 1 {
							// at least one argument was passed
							// pipe it to mods command
							modsResponse := exec.Command("mods", cmd...)

							if err := modsResponse.Run(); err != nil {
								wish.Fatal(sess, err)
							}
							modsResponse.Wait()
							// out, _ := glamo.Render(aiResponse, "dark")
							println(modsResponse.Stdout)
							wish.Print(sess, modsResponse.Stdout)
							_ = sess.Exit(1)
						} else {
							// no args were passed launch bubbletea
						}

						next(sess)
					}
				},
			),
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
	// Create a lipgloss.Renderer for the session
	renderer := bubbletea.MakeRenderer(s)
	// Set up the model with the current session and styles.
	// We'll use the session to call wish.Command, which makes it compatible
	// with tea.Command.
	m := model{
		sess:     s,
		style:    renderer.NewStyle().Foreground(lipgloss.Color("8")),
		errStyle: renderer.NewStyle().Foreground(lipgloss.Color("3")),
	}
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

type model struct {
	err      error
	sess     ssh.Session
	style    lipgloss.Style
	errStyle lipgloss.Style
}

func (m model) Init() tea.Cmd {
	return nil
}

type cmdFinishedMsg struct{ err error }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c := wish.Command(m.sess, "mods", m.sess.Command()...)
		cmd := tea.Exec(c, func(err error) tea.Msg {
			if err != nil {
				log.Error("shell finished", "error", err)
			}
			return cmdFinishedMsg{err: err}
		})
		return m, cmd
	case cmdFinishedMsg:
		m.err = msg.err
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return m.errStyle.Render(m.err.Error() + "\n")
	}
	return m.style.Render("Press 'e' to edit, 's' to hop into a shell, or 'q' to quit...\n")
}
