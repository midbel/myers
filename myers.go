package myers

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"slices"
)

type Formatter[T any] interface {
	Equal(io.Writer, []T) error
	Insert(io.Writer, []T) error
	Delete(io.Writer, []T) error
}

type Equaler[T any] interface {
	Equal(T) bool
}

var ErrEnd = errors.New("end of sequence")

type Sequence[T any] interface {
	Next() (T, error)
	Peek() (T, error)
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

func DiffFunc[T any](w io.Writer, f Formatter[T], fst, snd []T, eq func(T, T) bool) error {
	steps := ScriptFunc(fst, snd, eq)
	if len(steps) == 0 {
		return nil
	}
	ws := bufio.NewWriter(w)

	var x, y int
	for _, s := range steps {
		switch s.Op {
		case EqualOp:
			if _, err := ws.WriteString("= "); err != nil {
				return err
			}
			if err := f.Equal(ws, snd[y:y+s.Count]); err != nil {
				return err
			}
			x += s.Count
			y += s.Count
		case InsertOp:
			if _, err := ws.WriteString("+ "); err != nil {
				return err
			}
			if err := f.Insert(ws, snd[y:y+s.Count]); err != nil {
				return err
			}
			y += s.Count
		case DeleteOp:
			if _, err := ws.WriteString("- "); err != nil {
				return err
			}
			if err := f.Delete(ws, fst[x:x+s.Count]); err != nil {
				return err
			}
			x += s.Count
		default:
			continue
		}
		if _, err := ws.WriteString("\n"); err != nil {
			return err
		}
	}
	return ws.Flush()
}

type Step struct {
	Op    Op
	Pos   int
	Count int
}

func (s Step) String() string {
	return fmt.Sprintf("%s(%d)", s.Op, s.Count)
}

func Expand(steps []Step) []Step {
	all := make([]Step, 0, len(steps))
	for i := range steps {
		for j := 0; j < steps[i].Count; j++ {
			s := steps[i]
			s.Count = 1
			all = append(all, s)
		}
	}
	return all
}

func ExpandScriptFunc[T any](fst, snd []T, eq func(T, T) bool) []Step {
	steps := ScriptFunc(fst, snd, eq)
	return Expand(steps)
}

func ExpandScript[T Equaler[T]](fst, snd []T) []Step {
	return ExpandScriptFunc(fst, snd, func(a, b T) bool {
		return a.Equal(b)
	})
}

func ScriptFunc[T any](fst, snd []T, eq func(T, T) bool) []Step {
	return buildScript(fst, snd, true, eq)
}

func Script[T Equaler[T]](fst, snd []T) []Step {
	return ScriptFunc(fst, snd, func(a, b T) bool {
		return a.Equal(b)
	})
}

func buildScript[T any](fst, snd []T, compress bool, eq func(T, T) bool) []Step {
	var (
		before = commonPrefix(fst, snd, eq)
		after  = commonSuffix(fst[before:], snd[before:], eq)
	)
	fst = fst[before : len(fst)-after]
	snd = snd[before : len(snd)-after]

	if steps := tryScript(fst, snd, before, after); len(steps) >= 1 {
		return compactScript(steps)
	}

	var (
		script = buildPath(fst, snd, eq)
		steps  []Step
	)
	for offset := 0; script != nil; {
		if compress && offset > 0 && len(steps) > 0 && script.Op == steps[offset-1].Op {
			steps[offset-1].Count += script.Count
		} else {
			steps = append(steps, script.Step)
			offset++
		}
		script = script.parent
	}
	slices.Reverse(steps)
	if before > 0 {
		steps = append([]Step{equalStep(0, before)}, steps...)
	}
	if after > 0 {
		steps = append(steps, equalStep(before+len(fst), after))
	}
	return compactScript(steps)
}

func equalStep(pos, count int) Step {
	return createStep(EqualOp, pos, count)
}

func createStep(op Op, pos, count int) Step {
	return Step{
		Op:    op,
		Pos:   pos,
		Count: count,
	}
}

type path struct {
	Step
	parent *path
}

func tryScript[T any](fst, snd []T, before, after int) []Step {
	switch {
	case len(fst) == 0 && len(snd) > 0:
		return []Step{
			equalStep(0, before),
			createStep(InsertOp, before, len(snd)),
			equalStep(len(snd), after),
		}
	case len(fst) > 0 && len(snd) == 0:
		return []Step{
			equalStep(0, before),
			createStep(DeleteOp, before, len(fst)),
			equalStep(len(snd), after),
		}
	case len(fst) == 0 && len(snd) == 0:
		// equal
		return []Step{equalStep(0, before+after)}
	default:
		return nil
	}
}

func compactScript(steps []Step) []Step {
	return slices.DeleteFunc(steps, func(s Step) bool {
		return s.Count == 0
	})
}

func buildPath[T any](fst, snd []T, eq func(T, T) bool) *path {
	x, y, count := advance(0, 0, fst, snd, eq)

	root := &path{
		Step: equalStep(x, count),
	}
	if x == len(fst) && y == len(snd) {
		return root
	}

	var (
		paths  = map[int]*path{0: root}
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
				xp = paths[pred].Pos
				op = InsertOp
			case offset == edit:
				// delete only
				pred = offset - 1
				xp = paths[pred].Pos + 1
				op = DeleteOp
			case paths[offset-1].Pos+1 < paths[offset+1].Pos:
				// insert
				pred = offset + 1
				xp = paths[pred].Pos
				op = InsertOp
			default:
				// delete
				pred = offset - 1
				xp = paths[pred].Pos + 1
				op = DeleteOp
			}
			yp = xp - offset
			curr = paths[pred]

			child := &path{
				parent: curr,
				Step:   createStep(op, xp, 1),
			}
			nx, ny, count := advance(xp, yp, fst, snd, eq)

			candidate := child

			if count > 0 {
				candidate = &path{
					parent: child,
					Step:   equalStep(nx, count),
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

func advance[T any](x, y int, fst, snd []T, eq func(T, T) bool) (int, int, int) {
	var count int
	for x < len(fst) && y < len(snd) && eq(fst[x], snd[y]) {
		x++
		y++
		count++
	}
	return x, y, count
}

func commonPrefix[T any](fst, snd []T, eq func(T, T) bool) int {
	var x, y, count int
	for x < len(fst) && y < len(snd) && eq(fst[x], snd[y]) {
		x++
		y++
		count++
	}
	return count
}

func commonSuffix[T any](fst, snd []T, eq func(T, T) bool) int {
	var (
		x     = len(fst) - 1
		y     = len(snd) - 1
		count int
	)
	for x >= 0 && y >= 0 && eq(fst[x], snd[y]) {
		x--
		y--
		count++
	}
	return count
}
