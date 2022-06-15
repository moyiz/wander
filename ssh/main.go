package main

// An example Bubble Tea server. This will put an ssh session into alt screen
// and continually print up to date terminal information.

import (
	"context"
	"fmt"
	"github.com/robinovitch61/wander/components/app"
	"github.com/robinovitch61/wander/constants"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/gliderlabs/ssh"
)

func main() {
	host := os.Getenv(constants.WanderSSHHost)
	if host == "" {
		fmt.Printf("Set environment variable %s\n", constants.WanderSSHHost)
		os.Exit(1)
	}

	portStr := os.Getenv(constants.WanderSSHPort)
	if portStr == "" {
		fmt.Printf("Set environment variable %s\n", constants.WanderSSHPort)
		os.Exit(1)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}

	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithMiddleware(
			bm.Middleware(teaHandler),
			lm.Middleware(),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Starting SSH server on %s:%d", host, port)
	go func() {
		if err = s.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	<-done
	log.Println("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	return app.InitialModel(), []tea.ProgramOption{tea.WithAltScreen()}
}
