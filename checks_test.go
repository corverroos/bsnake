package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/luno/jettison/jtest"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

type input struct {
	Path    string
	Content []byte
	Req     GameRequest
}

func inputs(t *testing.T) []input {
	fl, err := filepath.Glob("testdata/input-*.json")
	jtest.RequireNil(t, err)

	var res []input
	for _, f := range fl {
		b, err := os.ReadFile(f)
		jtest.RequireNil(t, err)

		var req GameRequest
		jtest.RequireNil(t, json.Unmarshal(b, &req))

		res = append(res, input{
			Path:    f,
			Content: b,
			Req:     req,
		})
	}

	return res
}

func TestFixInput(t *testing.T) {

	for _, in := range inputs(t) {
		req := in.Req
		req.You = fixSnake(req.You)
		for i := 0; i < len(req.Board.Snakes); i++ {
			req.Board.Snakes[i] = fixSnake(req.Board.Snakes[i])
		}

		viz := boardToViz(req)
		err := os.WriteFile(strings.ReplaceAll(in.Path, ".json", ".board.txt"), []byte(viz), 0644)
		jtest.RequireNil(t, err)

		b, err := json.MarshalIndent(req, "", " ")
		jtest.RequireNil(t, err)
		if !bytes.Equal(b, in.Content) {
			err = os.WriteFile(in.Path, b, 0644)
			jtest.RequireNil(t, err)
		}
	}
}

func boardToViz(req GameRequest) string {
	var res [][]rune
	const (
		s  = '.'
		sh = 'S'
		sb = 's'
		yh = 'Y'
		yb = 'y'
		f  = '*'
	)
	for y := 0; y < req.Board.Height; y++ {
		var row []rune
		for x := 0; x < req.Board.Width; x++ {
			row = append(row, s)
		}
		res = append(res, row)
	}

	for _, snake := range req.Board.Snakes {
		h := sh
		b := sb
		if snake.ID == req.You.ID {
			h = yh
			b = yb
		}
		for i, c := range snake.Body {
			r := b
			if i == 0 {
				r = h
			}
			res[c.Y][c.X] = r
		}
	}
	for _, c := range req.Board.Food {
		res[c.Y][c.X] = f
	}
	var sl []string
	for _, row := range res {
		sl = append([]string{string(row)}, sl...)
	}
	return strings.Join(sl, "\n")
}

func TestArea(t *testing.T) {
	tests := []struct {
		Name string
		Exp  []string
	}{
		{
			Name: "001",
			Exp:  []string{"up:0", "down:32", "right:21", "left:39"},
		},
		{
			Name: "002",
			Exp:  []string{"up:0", "down:0", "right:2", "left:37"},
		},
		{
			Name: "003",
			Exp:  []string{"up:0", "down:0", "right:8", "left:8"},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			f, err := os.Open(path.Join("testdata", "input-"+test.Name+".json"))
			jtest.RequireNil(t, err)

			var req GameRequest
			err = json.NewDecoder(f).Decode(&req)
			jtest.RequireNil(t, err)

			var areas []string
			for _, m := range Moves {
				a := fill(req, 2*req.You.Length, req.You.Head.Move(m))
				areas = append(areas, fmt.Sprintf("%s:%d", m.String(), a.Size()))
			}

			require.EqualValues(t, test.Exp, areas, boardToViz(req))
		})
	}
}

func TestScoreMoves(t *testing.T) {

	type Exp map[Move]func(t *testing.T, m Move, scores map[Move]int)

	tests := []struct {
		Name string
		Exp  Exp
	}{
		{
			Name: "001",
			Exp:  Exp{Up: requireBack, Down: require2nd, Left: require1st, Right: require3rd},
		},
		{
			Name: "002",
			Exp:  Exp{Up: requireBack, Down: requireWall, Left: require1st, Right: requireBad},
		},
		{
			Name: "003",
			Exp:  Exp{Up: requireWall, Down: requireBack, Left: require1st, Right: require1st},
		},
		{
			Name: "004",
			Exp:  Exp{Up: requireWall, Down: require2nd, Left: require1st, Right: requireBack},
		},
		{
			Name: "005",
			Exp:  Exp{Up: requireBack, Down: require1st, Left: require3rd, Right: require2nd},
		},
		{
			Name: "006",
			Exp:  Exp{Up: require2nd, Down: requireBack, Left: requireBody, Right: require1st},
		},
		{
			Name: "007",
			Exp:  Exp{Up: requireBad, Down: require1st, Left: requireBack, Right: require2nd},
		},
		{
			Name: "008",
			Exp:  Exp{Up: requireWall, Down: requireBack, Left: require2nd, Right: require1st},
		},
		{
			Name: "009",
			Exp:  Exp{Up: require2nd, Down: requireWall, Left: require1st, Right: requireBack},
		},
		{
			Name: "010",
			Exp:  Exp{Up: requireBack, Down: require1st, Left: require2nd, Right: require3rd},
		},
		{
			Name: "011",
			Exp:  Exp{Up: require2nd, Down: requireWall, Left: require1st, Right: requireBack},
		},
		{
			Name: "012",
			Exp:  Exp{Up: requireBack, Down: requireWall, Left: require1st, Right: require2nd},
		},
		{
			Name: "013",
			Exp:  Exp{Up: require1st, Down: require2nd, Left: requireBack, Right: requireWall},
		},
		{
			Name: "014",
			Exp:  Exp{Up: requireBody, Down: require2nd, Left: requireBack, Right: require1st},
		},
		{
			Name: "015",
			Exp:  Exp{Up: requireBody, Down: require1st, Left: requireBack, Right: require2nd},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			f, err := os.Open(path.Join("testdata", "input-"+test.Name+".json"))
			jtest.RequireNil(t, err)

			var req GameRequest
			err = json.NewDecoder(f).Decode(&req)
			jtest.RequireNil(t, err)

			w := basicWeights

			t0 := time.Now()
			scores := make(map[Move]int)
			for _, m := range Moves {
				score, err := scoreMove(context.Background(), req, w,m, true)
				jtest.RequireNil(t, err)
				scores[m] = score
			}
			d := time.Since(t0)
			require.True(t, d < time.Millisecond*20, d)

			fmt.Println("duration")
			fmt.Println(scores)
			for move, fn := range test.Exp {
				fn(t, move, scores)
			}
		})
	}
}

func TestFill(t *testing.T) {
	for _, in := range inputs(t) {
		t0 := time.Now()
		a := fill(in.Req, in.Req.You.Length*2, in.Req.You.Head.Move(Up))
		fmt.Printf("%s %v\n%v\n", in.Path, time.Since(t0), a.Viz())
	}
}

func TestStack(t *testing.T) {
	parse := func(file string, msg interface{}) {
		f, err := os.Open(file)
		jtest.RequireNil(t, err)
		jtest.RequireNil(t, json.NewDecoder(f).Decode(msg))
	}
	var req GameRequest
	parse("testdata/external/11.board.json", &req)

	a := fill(req, 2, req.You.Head.Move(Up))
	require.Equal(t, -3, a[Coord{3,3}])
	ttl, you, ok := isBody(req, req.You.Head)
	require.True(t, ok)
	require.True(t, you)
	require.Equal(t, 3, ttl)
}

var reload = flag.Bool("reload", false, "")

func TestExternal(t *testing.T) {
	if *reload {
		reloadExternal(t)
	}

	files, err := filepath.Glob("testdata/external/*.board.json")
	jtest.RequireNil(t, err)

	for i, file := range files {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			parse := func(file string, msg interface{}) {
				f, err := os.Open(file)
				jtest.RequireNil(t, err)
				jtest.RequireNil(t, json.NewDecoder(f).Decode(msg))
			}
			var req GameRequest
			parse(file, &req)
			var moves []Move
			parse(strings.Replace(file, ".board.",".moves.", 1), &moves)

			w := basicWeights

			scores := make(map[Move]int)
			for _, m := range Moves {
				score, err := scoreMove(context.Background(), req, w,m,true)
				jtest.RequireNil(t, err)
				scores[m] = score
			}

			fmt.Printf("scores=%v\n", scores)
			fmt.Printf("moves=%v\n", moves)
			for _, move := range moves {
				if i == 5 && move == Right {
					// Not really a good move
					require2nd(t, move, scores)
					continue
				}
				require1st(t, move, scores)
			}
		})
	}
}

// reloadExternal some unit tests from BattlesnakeOfficial/coding-badly.
func reloadExternal(t *testing.T) {
	jtest.RequireNil(t, os.RemoveAll("testdata/external"))
	jtest.RequireNil(t, os.MkdirAll("testdata/external", 0755))

	for i := 1; i < 20; i++ {
		u := fmt.Sprintf("https://raw.githubusercontent.com/BattlesnakeOfficial/coding-badly/main/src/tests/%d.move", i)
		res, err := http.Get(u)
		jtest.RequireNil(t, err)
		if res.StatusCode == http.StatusNotFound {
			if i < 12 {
				require.Fail(t, "missing expected file", i, u)
			}
			return
		}
		require.Equal(t, res.StatusCode, http.StatusOK)

		d := json.NewDecoder(res.Body)

		exp := struct{ AcceptedMoves []string }{}
		jtest.RequireNil(t, d.Decode(&exp))
		var req GameRequest
		jtest.RequireNil(t, d.Decode(&req))
		jtest.RequireNil(t, res.Body.Close())

		req.You = fixSnake(req.You)
		for i := 0; i < len(req.Board.Snakes); i++ {
			req.Board.Snakes[i] = fixSnake(req.Board.Snakes[i])
		}

		js, err := json.MarshalIndent(req, "", " ")
		jtest.RequireNil(t, err)
		err = os.WriteFile(fmt.Sprintf("testdata/external/%02d.board.json", i), js, 0644)
		jtest.RequireNil(t, err)

		viz := boardToViz(req)
		err = os.WriteFile(fmt.Sprintf("testdata/external/%02d.board.txt", i), []byte(viz), 0644)
		jtest.RequireNil(t, err)

		fmt.Printf("JCR: exp=%#v\n", exp)
		empty := len(exp.AcceptedMoves) == 1 && exp.AcceptedMoves[0] == ""
		var moves []Move
		if !empty {
			for _, e := range exp.AcceptedMoves {
				for _, move := range Moves {
					if move.String() == e {
						moves = append(moves, move)
					}
				}
			}

			require.Equal(t, len(exp.AcceptedMoves), len(moves), "%v vs %v", exp.AcceptedMoves, moves)
		}

		js, err = json.MarshalIndent(moves, "", " ")
		jtest.RequireNil(t, err)
		err = os.WriteFile(fmt.Sprintf("testdata/external/%02d.moves.json", i), js, 0644)
		jtest.RequireNil(t, err)
	}
}

func flatSort(scores map[Move]int) (res []int) {
	for _, s := range scores {
		res = append(res, s)
	}
	sort.Ints(res)
	return res
}
func requireWall(t *testing.T, m Move, scores map[Move]int) {
	t.Helper()
	require.Equal(t, scoreWall, scores[m])
}
func requireBack(t *testing.T, m Move, scores map[Move]int) {
	t.Helper()
	require.Equal(t, scoreBack, scores[m])
}
func requireBody(t *testing.T, m Move, scores map[Move]int) {
	t.Helper()
	require.Equal(t, scoreBody, scores[m])
}
func require1st(t *testing.T, m Move, scores map[Move]int) {
	t.Helper()
	all := flatSort(scores)
	require.Equal(t, scores[m], all[len(all)-1])
}
func require2nd(t *testing.T, m Move, scores map[Move]int) {
	t.Helper()
	all := flatSort(scores)
	require.Equal(t, scores[m], all[len(all)-2])
}
func require3rd(t *testing.T, m Move, scores map[Move]int) {
	t.Helper()
	all := flatSort(scores)
	require.Equal(t, scores[m], all[len(all)-3])
}
func requireBad(t *testing.T, m Move, scores map[Move]int) {
	t.Helper()
	score := scores[m]
	require.True(t, score < 10, score)
}
func fixSnake(b Battlesnake) Battlesnake {
	b.Length = len(b.Body)
	b.Head = b.Body[0]
	return b
}
