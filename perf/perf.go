package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer cancel()

	sl := []server{
		{
			Port:   8082,
			Snakes: []string{"boomboom"},
		}, {
			Port:   8083,
			Ref:    1,
			Snakes: []string{"boomboom", "basic"},
		},
	}

	opts := options{Total: 50}

	res, err := run(ctx, opts, sl...)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(3)
	}
	fmt.Println(res)
}

type options struct {
	Total int
}

type server struct {
	Ref    int
	Port   int
	Snakes []string
}

type result struct {
	Wins map[string]int
}

func run(ctx context.Context, opt options, sl ...server) (result, error) {
	exec.Command("pkill", "bsnake").Run()

	for i, s := range sl {
		branch := "main"
		if s.Ref > 0 {
			branch = fmt.Sprintf("main^%d", s.Ref)
		}
		out, err := exec.Command("git", "checkout", branch).CombinedOutput()
		if err != nil {
			return result{}, fmt.Errorf("gco %v: %s", err, out)
		}

		cmd := exec.CommandContext(ctx, "go", "run", "github.com/corverroos/bsnake")
		cmd.Dir = "/Users/corver/core/repos/bsnake"
		cmd.Env = append([]string{fmt.Sprintf("BIND=localhost:%d", s.Port)}, os.Environ()...)
		go func(cmd *exec.Cmd) {
			fmt.Printf("Starting server %s\n", cmd.String())
			out, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("server error %d: %s", i, out)
			}
		}(cmd)
		time.Sleep(time.Second)

		_, err = exec.Command("git", "checkout", "main").CombinedOutput()
		if err != nil {
			return result{}, fmt.Errorf("gco %v: %s", err, out)
		}
	}

	results := make(map[string]int)

	for i := 0; i < opt.Total; i++ {
		args := []string{"play", "-W", "11", "-H", "11"}
		for _, s := range sl {
			for _, snake := range s.Snakes {
				args = append(args, "--name", nameSnake(s, snake))
				args = append(args, "--url", fmt.Sprintf("http://localhost:%d/%s/", s.Port, snake))
			}
		}
		cmd := exec.CommandContext(ctx, "battlesnake", args...)
		fmt.Printf("Running game %d %s\n", i, cmd.String())
		out, err := cmd.CombinedOutput()
		if err != nil {
			return result{}, fmt.Errorf("game error %v: %s", err, out)
		}

		for _, line := range strings.Split(string(out), "\n") {
			if !strings.Contains(line, "Game completed after") {
				continue
			}
			suffix := strings.Split(line, "turns. ")
			if len(suffix) != 2 {
				fmt.Printf("Game no winner %d: %s", i, line)
				break
			}
			winner := strings.TrimSpace(strings.TrimSuffix(suffix[1], " is the winner."))
			if strings.Contains(winner, " ") {
				fmt.Printf("Game no winner %d: %s", i, line)
				break
			}
			results[winner]++
			fmt.Printf("Winner %s\n", winner)
		}
	}

	return result{
		Wins: results,
	}, nil
}

func nameSnake(srv server, path string) string {
	return fmt.Sprintf("%d_%s", srv.Ref, path)
}
