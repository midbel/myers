package explore

import (
	"errors"

	"github.com/midbel/myers"
)

type Forkable[T any] interface {
	myers.Sequence[T]
	Fork() Forkable[T]
}

type State int

const (
	StateSucceed State = iota
	StateFail
)

func (s State) String() string {
	switch s {
	case StateSucceed:
		return "success"
	case StateFail:
		return "failure"
	default:
		return "normal"
	}
}

func (s State) Succeed() bool {
	return s == StateSucceed
}

func (s State) Failure() bool {
	return s == StateFail
}

type Entry[T any] struct {
	Edit  int
	Op    myers.Op
	Value T
	State State
}

func Explore[T myers.Equaler[T]](fst, snd Forkable[T], do func(Entry[T]) error) error {
	err := walk(fst.Fork(), snd.Fork(), 0, do)
	if errors.Is(err, myers.ErrEnd) {
		err = nil
	}
	return err
}

func createEntry[T any](edit int) Entry[T] {
	return Entry[T]{
		Edit: edit,
		Op:   myers.EqualOp,
	}
}

func walk[T myers.Equaler[T]](fst, snd Forkable[T], edit int, do func(Entry[T]) error) error {
	v1, e1 := fst.Peek()
	v2, e2 := snd.Peek()

	entry := createEntry[T](edit)
	switch {
	case errors.Is(e1, myers.ErrEnd) && errors.Is(e2, myers.ErrEnd):
		entry.State = StateSucceed
		return do(entry)
	case e1 != nil || e2 != nil:
		if e1 != nil && !errors.Is(e1, myers.ErrEnd) {
			return e1
		}
		if e2 != nil && !errors.Is(e2, myers.ErrEnd) {
			return e2
		}
		entry.State = StateFail
		return do(entry)
	case v1.Equal(v2):
		fst.Next()
		snd.Next()
		entry.Value = v1
		if err := do(entry); err != nil {
			return err
		}
		if err := walk(fst, snd, edit, do); err != nil {
			return err
		}
	default:
		entry.Edit += 1

		delFst := fst.Fork()
		delSnd := snd.Fork()
		entry.Op = myers.DeleteOp
		if err := do(entry); err != nil {
			return err
		}

		delFst.Next()
		entry.Value = v1
		if err := walk(delFst, delSnd, edit+1, do); err != nil {
			return err
		}

		insFst := fst.Fork()
		insSnd := snd.Fork()
		entry.Op = myers.InsertOp
		entry.Value = v2
		if err := do(entry); err != nil {
			return err
		}

		insSnd.Next()
		if err := walk(insFst, insSnd, edit+1, do); err != nil {
			return err
		}
	}
	return nil
}
