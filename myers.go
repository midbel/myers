package myers

type Op uint8

const (
	EqualOp Op = iota
	InsertOp
	DeleteOp
)

func Diff(fst, snd []string) {
	var (
		x, y    int
		history []map[int]int
		done    bool
		edit    int
	)
	advance(&x, &y, fst, snd)

	paths := map[int]int{
		0: x,
	}
	history = append(history, paths)

	for edit = 1; !done; edit++ {
		next := make(map[int]int)
		for offset, x := range paths {
			y := x - offset
			if y < len(snd) {
				insX := x
				insY := y + 1

				advance(&insX, &insY, fst, snd)

				if insX == len(fst) && insY == len(snd) {
					x, y = insX, insY
					done = true
					break
				}

				insOffset := insX - insY
				if old, ok := next[insOffset]; !ok || insX > old {
					next[insOffset] = insX
				}
			}

			if x < len(fst) {
				delX := x + 1
				delY := y

				advance(&delX, &delY, fst, snd)

				if delX == len(fst) && delY == len(snd) {
					x, y = delX, delY
					done = true
					break
				}

				delOffset := delX - delY
				if old, ok := next[delOffset]; !ok || delX > old {
					next[delOffset] = delX
				}
			}
		}

		paths = next
		history = append(history, next)
		if done {
			break
		}
	}
	if x == len(fst) && y == len(snd) {
		return
	}

	ops, pos := collectOps(history, fst, snd)
	slices.Reverse(ops)
	slices.Reverse(pos)

	script := rebuildScript(fst, pos, ops)
	fmt.Println(ops, pos, script)
}

func advance(x, y *int, fst, snd []string) {
	for *x < len(fst) && *y < len(snd) && fst[*x] == snd[*y] {
		*x++
		*y++
	}
}

func collectOps(history []map[int]int, fst, snd []string) ([]Op, []int) {
	x, y := len(fst), len(snd)

	var (
		ops []Op
		pos []int
	)

	for e := len(history) - 2; e >= 0; e-- {
		var (
			offset = x - y
			found  bool
		)
		if prevX, ok := history[e][offset+1]; ok {
			prevY := prevX - (offset + 1)

			nextX := prevX
			nextY := prevY + 1

			advance(&nextX, &nextY, fst, snd)
			if nextX == x && nextY == y {
				found = true
				x, y = prevX, prevY
				ops = append(ops, InsertOp)
				pos = append(pos, x)
			}
		}
		if found {
			continue
		}
		if prevX, ok := history[e][offset-1]; ok {
			prevY := prevX - (offset - 1)

			nextX := prevX + 1
			nextY := prevY

			advance(&nextX, &nextY, fst, snd)
			if nextX == x && nextY == y {
				x, y = prevX, prevY
				ops = append(ops, DeleteOp)
				pos = append(pos, x)
			}
		}
	}
	return ops, pos
}

func rebuildScript(fst []string, pos []int, codes []Op) []Op {
	var (
		last   int
		script []Op
	)
	for i := range pos {
		for j := last; j < pos[i]; j++ {
			script = append(script, EqualOp)
		}
		last = pos[i]
		script = append(script, codes[i])
		if codes[i] == DeleteOp {
			last++
		}
	}
	for j := last; j < len(fst); j++ {
		script = append(script, EqualOp)
	}
	return script
}
