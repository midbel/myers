package myers

import (
	"errors"
	"io"
	"slices"
)

type Equaler[T any] interface {
	Equal(T) bool
}

var ErrEnd = errors.New("end of sequence")

type Sequence[T any] interface {
	Next() (T, error)
	Peek() (T, error)
}

type Forkable[T any] interface {
	Sequence[T]
	Fork() Forkable[T]
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

func Diff[T Equaler[T]](w io.Writer, fst, snd []T) error {
	return nil
}

type Step struct {
	Op    Op
	Pos   int
	Count int
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
	return steps
}

func ExpandScriptFunc[T any](fst, snd []T, eq func(a, b T) bool) []Step {
	steps := ScriptFunc(fst, snd, eq)
	return Expand(steps)
}

func ExpandScript[T Equaler[T]](fst, snd []T) []Step {
	return ExpandScriptFunc(fst, snd, func(a, b T) bool {
		return a.Equal(b)
	})
}

func ScriptFunc[T any](fst, snd []T, eq func(a, b T) bool) []Step {
	var (
		script = buildPath(fst, snd, eq)
		steps  []Step
	)
	if script == nil {
		return nil
	}
	for {
		steps = append(steps, script.Step)
		script = script.parent
		if script == nil {
			break
		}
	}
	slices.Reverse(steps)
	return steps
}

func Script[T Equaler[T]](fst, snd []T) []Step {
	return ScriptFunc(fst, snd, func(a, b T) bool {
		return a.Equal(b)
	})
}

type Entry[T any] struct {
	Edit    int
	Op      Op
	First   T
	Second  T
	Success bool
	Failure bool
}

func Explore[T Equaler[T]](fst, snd Forkable[T], do func(Entry[T]) error) error {
	err := explore(fst.Fork(), snd.Fork(), 0, do)
	if errors.Is(err, ErrEnd) {
		err = nil
	}
	return err
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

func buildPath[T any](fst, snd []T, eq func(a, b T) bool) *path {
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

func advance[T any](x, y int, fst, snd []T, eq func(a, b T) bool) (int, int, int) {
	var count int
	for x < len(fst) && y < len(snd) && eq(fst[x], snd[y]) {
		x++
		y++
		count++
	}
	return x, y, count
}

func createEntry[T any](edit int, v1, v2 T) Entry[T] {
	return Entry[T]{
		Edit:   edit,
		Op:     EqualOp,
		First:  v1,
		Second: v2,
	}
}

func explore[T Equaler[T]](fst, snd Forkable[T], edit int, do func(Entry[T]) error) error {
	v1, e1 := fst.Peek()
	v2, e2 := snd.Peek()

	entry := createEntry(edit, v1, v2)
	switch {
	case errors.Is(e1, ErrEnd) && errors.Is(e2, ErrEnd):
		entry.Success = true
		return do(entry)
	case e1 != nil || e2 != nil:
		if e1 != nil && !errors.Is(e1, ErrEnd) {
			return e1
		}
		if e2 != nil && !errors.Is(e2, ErrEnd) {
			return e2
		}
		entry.Failure = true
		return do(entry)
	case v1.Equal(v2):
		fst.Next()
		snd.Next()

		if err := do(entry); err != nil {
			return err
		}
		if err := explore(fst, snd, edit, do); err != nil {
			return err
		}
	default:
		entry.Edit += 1

		delFst := fst.Fork()
		delSnd := snd.Fork()
		entry.Op = DeleteOp
		if err := do(entry); err != nil {
			return err
		}

		delFst.Next()
		if err := explore(delFst, delSnd, edit+1, do); err != nil {
			return err
		}

		insFst := fst.Fork()
		insSnd := snd.Fork()
		entry.Op = InsertOp
		if err := do(entry); err != nil {
			return err
		}

		insSnd.Next()
		if err := explore(insFst, insSnd, edit+1, do); err != nil {
			return err
		}
	}
	return nil
}
