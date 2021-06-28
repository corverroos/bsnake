package board

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/luno/jettison/jtest"
)

func TestGenRand(t *testing.T) {
	var res strings.Builder
	res.WriteString("package board\n\nvar moveperms = [][]string{\n")
	perms := permutation([]int{0, 1, 2, 3})
	for _, perm := range perms {
		var moves []string
		for _, i := range perm {
			moves = append(moves, `"`+Moves[i]+`"`)
		}
		res.WriteString("\t\t{")
		res.WriteString(strings.Join(moves, ","))
		res.WriteString("},\n")
	}
	res.WriteString("\t}\n")
	res.WriteString("\n const perms = " + strconv.Itoa(len(perms)) + "\n\n")

	err := os.WriteFile("perms.go", []byte(res.String()), 0644)
	jtest.RequireNil(t, err)
}

func permutation(xs []int) (permuts [][]int) {
	var rc func([]int, int)
	rc = func(a []int, k int) {
		if k == len(a) {
			permuts = append(permuts, append([]int{}, a...))
		} else {
			for i := k; i < len(xs); i++ {
				a[k], a[i] = a[i], a[k]
				rc(a, k+1)
				a[k], a[i] = a[i], a[k]
			}
		}
	}
	rc(xs, 0)

	return permuts
}
