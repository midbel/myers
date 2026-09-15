package myers

import (
	"slices"
)

type Equaler[T any] interface {
	Equal(T) bool
}

type Op uint8

const (
	EqualOp Op = iota
	InsertOp
	DeleteOp
)

func (o Op) String() string {
	switch o {
	case EqualOp:
		return "equal"
	case InsertOp:
		return "insert"
	case DeleteOp:
		return "delete"
	default:
		return "???"
	}
}

func Same[T Equaler[T]](fst, snd []T) bool {
	script := Script(fst, snd)
	for i := range script {
		if script[i] != EqualOp {
			return false
		}
	}
	return true
}

func Diff[T Equaler[T]](fst, snd []T) {
}

func Script[T Equaler[T]](fst, snd []T) []Op {
	var (
		script = buildPath(fst, snd)
		ops    []Op
	)
	if script == nil {
		return nil
	}
	for {
		for i := 0; i < script.Count; i++ {
			ops = append(ops, script.Op)
		}
		script = script.parent
		if script == nil {
			break
		}
	}
	slices.Reverse(ops)
	return ops
}

func Explore[T Equaler[T]](fst, snd []T, do func(int, Op, T, T, bool, bool)) {
	explore(fst, snd, 0, 0, 0, do)
}

func explore[T Equaler[T]](fst, snd []T, x, y, edit int, do func(int, Op, T, T, bool, bool)) {
	for x < len(fst) && y < len(snd) && fst[x].Equal(snd[y]) {
		do(edit, EqualOp, fst[x], snd[y], false, false)
		x++
		y++
	}
	if x == len(fst) && y == len(snd) {
		do(edit, EqualOp, fst[x-1], snd[y-1], true, false)
		return
	}
	if x == len(fst) || y == len(snd) {
		do(edit, EqualOp, fst[x-1], snd[y-1], false, true)
		return
	}

	if x < len(fst) && y < len(snd) {
		do(edit, DeleteOp, fst[x], snd[y], false, false)
		explore(fst, snd, x+1, y, edit+1, do)
	}
	if y < len(snd) && x < len(fst) {
		do(edit, InsertOp, fst[x], snd[y], false, false)
		explore(fst, snd, x, y+1, edit+1, do)
	}
}

type Step struct {
	Edit   int
	Offset int // diagonale du step
	From   int // diagonale du parent, à Edit-1
	Op     Op
	X, Y   int
	Count  int
	Kept   bool
}

type path struct {
	Step
	parent *path
}

func buildPath[T Equaler[T]](fst, snd []T) *path {
	x, y, count := advance(0, 0, fst, snd)

	root := &path{
		Step: Step{
			Edit:   0,
			Offset: x - y,
			From:   0,
			Op:     EqualOp,
			X:      x,
			Y:      y,
			Count:  count,
		},
	}
	if x == len(fst) && y == len(snd) {
		return nil
	}

	var (
		paths  = map[int]*path{root.Offset: root}
		winner *path
	)

	for edit := 1; edit <= len(fst)+len(snd); edit++ {
		next := make(map[int]*path)
		for offset := -edit; offset <= edit; offset += 2 {
			var (
				pred int
				xp   int
				yp   int
				curr *path
				op   Op
			)
			switch {
			case offset == -edit:
				// insert only
				pred = offset + 1
				xp = paths[pred].X
				op = InsertOp
			case offset == edit:
				// delete only
				pred = offset - 1
				xp = paths[pred].X + 1
				op = DeleteOp
			case paths[offset-1].X+1 >= paths[offset+1].X:
				// delete
				pred = offset - 1
				xp = paths[pred].X + 1
				op = DeleteOp
			default:
				// insert
				pred = offset + 1
				xp = paths[pred].X
				op = InsertOp
			}
			yp = xp - offset
			curr = paths[pred]

			child := &path{
				parent: curr,
				Step: Step{
					Edit:   edit,
					Offset: offset,
					From:   curr.Offset,
					Op:     op,
					X:      xp,
					Y:      yp,
					Count:  1,
				},
			}
			nx, ny, count := advance(xp, yp, fst, snd)

			candidate := child

			if count > 0 {
				candidate = &path{
					parent: child,
					Step: Step{
						Edit:   edit,
						Offset: nx - ny,
						From:   offset,
						Op:     EqualOp,
						X:      nx,
						Y:      ny,
						Count:  count,
					},
				}
			}

			next[offset] = candidate
			if nx == len(fst) && ny == len(snd) {
				winner = candidate
				break
			}
		}
		paths = next
		if winner != nil {
			break
		}
	}
	return winner
}

func advance[T Equaler[T]](x, y int, fst, snd []T) (int, int, int) {
	var count int
	for x < len(fst) && y < len(snd) && fst[x].Equal(snd[y]) {
		x++
		y++
		count++
	}
	return x, y, count
}
