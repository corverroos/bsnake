package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/corverroos/bsrules/cli/commands"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer cancel()

	sl := []server{
		{
			Port:   8082,
			Snakes: []string{"v2", "v1", "mx3", "mx4"},
		},
		{
			Ref:    "e9de2a3c67548abc89fb59cea871ec02069ba26b",
			Port:   8083,
			Snakes: []string{"boomboom"},
		},
	}

	opts := options{
		Total:   50,
		Players: 4,
	}

	res, err := run(ctx, opts, sl...)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(3)
	}
	fmt.Println(res)
}

type options struct {
	Total   int
	Players int
}

type server struct {
	Ref    string
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
		if len(sl) > 1 {
			if s.Ref != "" {
				branch = s.Ref
			}

			out, err := exec.Command("git", "checkout", branch).CombinedOutput()
			if err != nil {
				return result{}, fmt.Errorf("gco %v: %s", err, out)
			}
		}

		cmd := exec.CommandContext(ctx, "go", "run", "github.com/corverroos/bsnake")
		cmd.Dir = "/Users/corver/core/repos/bsnake"
		cmd.Env = append([]string{fmt.Sprintf("BIND=localhost:%d", s.Port)}, os.Environ()...)
		go func(cmd *exec.Cmd) {
			fmt.Printf("Starting server %s on branch %v\n", cmd.String(), branch)
			out, err := cmd.CombinedOutput()
			if ctx.Err() == nil && err != nil {
				fmt.Printf("server error %d: %s", i, out)
			}
		}(cmd)

		time.Sleep(time.Second)

		if len(sl) > 1 {
			out, err := exec.Command("git", "checkout", "main").CombinedOutput()
			if err != nil {
				return result{}, fmt.Errorf("gco %v: %s", err, out)
			}
		}
	}

	wins := make(map[string]map[string]int)
	losses := make(map[string]map[string]int)
	plays := make(map[string]int)
	draws := make(map[string]int)
	for i := 0; i < opt.Total; i++ {
		if ctx.Err() != nil {
			break
		}

		func(i int) {
			var logs []string
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Panic: %v", r)
					fmt.Print(strings.Join(logs, "\n"))
				}
			}()

			game := []string{"standard", "royale"}[i%2]

			o := &commands.Options{
				Width:    11,
				Height:   11,
				GameType: game,
				Seed:     int64(i),
				Log: func(s string, i ...interface{}) {
					logs = append(logs, fmt.Sprintf(s, i...))
				},
			}

			for _, s := range sl {
				for _, snake := range s.Snakes {
					o.Names = append(o.Names, nameSnake(s, len(sl), snake))
					o.URLs = append(o.URLs, fmt.Sprintf("http://localhost:%d/%s/", s.Port, snake))
				}
			}

			for opt.Players > 0 && len(o.Names) > opt.Players {
				drop := rand.Intn(len(o.Names))
				o.Names = append(o.Names[0:drop], o.Names[drop+1:]...)
				o.URLs = append(o.URLs[0:drop], o.URLs[drop+1:]...)
			}

			t0 := time.Now()
			res := commands.Run(o)

			for snake, info := range res.Infos {
				if _, ok := plays[snake]; !ok {
					wins[snake] = make(map[string]int)
					losses[snake] = make(map[string]int)
					fmt.Printf("Infos %s: %v\n", snake, info)
				}
				plays[snake]++

				if res.Winner == "" {
					draws[snake]++
				} else {
					for other := range res.Infos {
						if other == snake && snake == res.Winner {
							wins[snake]["total"]++
						} else if other == snake && snake != res.Winner {
							losses[snake]["total"]++
						} else if snake == res.Winner {
							wins[snake][other]++
						} else if other == res.Winner {
							losses[snake][other]++
						}
					}
				}
			}

			var msg []string
			for s, p := range plays {
				msg = append(msg, fmt.Sprintf("%s p=%d\tw=%v\tl=%v", s, p, wins[s], losses[s]))
			}
			sort.Strings(msg)

			fmt.Printf("Results\t[winner=%s, turns=%d, game=%s, duration=%.0fs]\n\t%s\n", res.Winner, res.Turn, game, time.Since(t0).Seconds(), strings.Join(msg, "\n\t"))
		}(i)
	}

	return result{
		//Wins: wins,
	}, nil
}

func nameSnake(srv server, sl int, path string) string {
	if sl == 1 || srv.Ref == "" {
		return path
	}
	return fmt.Sprintf("%s_%s", srv.Ref[:1], path)
}
