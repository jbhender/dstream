package dstream

import (
	"fmt"
	"os"
)

// Join combines several Dstreams that have been segmented by id
// variables.  Calls to Next always advance the first Dstream (Data[0]
// )by one chunk.  The other elements of Data are advanced until their
// id variable is equal to (if possible) or greater than the id
// variable of Data[0].  If equality is achieved, the corresponding
// element of Status is set to true.  Status[0] is always false and
// has no meaning.
//
// The Data values are assumed to be segmented so that the id
// variable is constant within chunks, and increases with subsequent
// calls to Next.
type Join struct {

	// A sequence of segmented Dstreams to advance in unison.
	Data []Dstream

	// Status[j] means that the id variable for Data value j is
	// equal to the id variable for Data value 0.  Status[0] is
	// not used.
	Status []bool

	inames []string

	ipos []int

	id [][]uint64
}

// NewJoin creates a Join of the given Dstreams, using the variable
// names in inames as ids. The Dstreams in data must be segmented by
// the inames variables before calling NewJoin.
func NewJoin(data []Dstream, inames []string) *Join {
	w := &Join{
		Data:   data,
		inames: inames,
	}

	for k, da := range data {
		names := da.Names()
		f := false
		for j, na := range names {
			if na == inames[k] {
				w.ipos = append(w.ipos, j)
				f = true
				break
			}
		}
		if !f {
			msg := fmt.Sprintf("Can't find index variable %s in %dth dataset", inames[k], k)
			os.Stderr.WriteString(msg)
			os.Exit(1)
		}
	}

	w.Status = make([]bool, len(data))
	w.id = make([][]uint64, len(data))

	return w
}

func (w *Join) needsadvance(j int) bool {
	if w.id[j] == nil || len(w.id[j]) == 0 {
		return true
	}

	if w.id[j][0] < w.id[0][0] {
		return true
	}

	return false
}

func (w *Join) clearstatus() {
	for j, _ := range w.Status {
		w.Status[j] = false
	}
}

func (w *Join) Next() bool {

	f := w.Data[0].Next()
	if !f {
		return false
	}
	w.id[0] = w.Data[0].GetPos(w.ipos[0]).([]uint64)

	w.clearstatus()
	for j := 1; j < len(w.Data); j++ {

		for w.needsadvance(j) {
			f := w.Data[j].Next()
			if !f {
				break
			}
			w.id[j] = w.Data[j].GetPos(w.ipos[j]).([]uint64)
		}

		w.Status[j] = w.id[j][0] == w.id[0][0]
	}

	return true
}
