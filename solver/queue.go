/******************************************************************************************[Heap.h]
Copyright (c) 2003-2006, Niklas Een, Niklas Sorensson
Copyright (c) 2007-2010, Niklas Sorensson

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
**************************************************************************************************/

package solver

// A heap implementation with support for decrease/increase key. This is
// strongly inspired from Minisat's mtl/Heap.h.

type queue struct {
	activity []float64 // Activity of each variable. This should be the solver's slice, not a copy.
	content  []int     // Actual content.
	indices  []int     // Reverse queue, i.e position of each item in content; -1 means absence.
}

func newQueue(activity []float64) queue {
	q := queue{
		activity: activity,
	}
	for i := range q.activity {
		q.insert(i)
	}
	return q
}

func (q *queue) lt(i, j int) bool {
	return q.activity[i] > q.activity[j]
}

// Traversal functions.
func left(i int) int   { return i*2 + 1 }
func right(i int) int  { return (i + 1) * 2 }
func parent(i int) int { return (i - 1) >> 1 }

func (q *queue) percolateUp(i int) {
	x := q.content[i]
	p := parent(i)
	for i != 0 && q.lt(x, q.content[p]) {
		q.content[i] = q.content[p]
		q.indices[q.content[p]] = i
		i = p
		p = parent(p)
	}
	q.content[i] = x
	q.indices[x] = i
}

func (q *queue) percolateDown(i int) {
	x := q.content[i]
	for left(i) < len(q.content) {
		var child int
		if right(i) < len(q.content) && q.lt(q.content[right(i)], q.content[left(i)]) {
			child = right(i)
		} else {
			child = left(i)
		}
		if !q.lt(q.content[child], x) {
			break
		}
		q.content[i] = q.content[child]
		q.indices[q.content[i]] = i
		i = child
	}
	q.content[i] = x
	q.indices[x] = i
}

func (q *queue) len() int    { return len(q.content) }
func (q *queue) empty() bool { return len(q.content) == 0 }

func (q *queue) contains(n int) bool {
	return n < len(q.indices) && q.indices[n] >= 0
}

func (q *queue) get(index int) int {
	return q.content[index]
}

func (q *queue) decrease(n int) {
	q.percolateUp(q.indices[n])
}

func (q *queue) increase(n int) {
	q.percolateDown(q.indices[n])
}

func (q *queue) update(n int) {
	if !q.contains(n) {
		q.insert(n)
	} else {
		q.percolateUp(q.indices[n])
		q.percolateDown(q.indices[n])
	}
}

func (q *queue) insert(n int) {
	for i := len(q.indices); i <= n; i++ {
		q.indices = append(q.indices, -1)
	}
	q.indices[n] = len(q.content)
	q.content = append(q.content, n)
	q.percolateUp(q.indices[n])
}

func (q *queue) removeMin() int {
	x := q.content[0]
	q.content[0] = q.content[len(q.content)-1]
	q.indices[q.content[0]] = 0
	q.indices[x] = -1
	q.content = q.content[:len(q.content)-1]
	if len(q.content) > 1 {
		q.percolateDown(0)
	}
	return x
}

// Rebuild the heap from scratch, using the elements in 'ns':
func (q *queue) build(ns []int) {
	for i := range q.content {
		q.indices[q.content[i]] = -1
	}
	q.content = q.content[:0]
	for i, val := range ns {
		q.indices[val] = i
		q.content = append(q.content, val)
	}
	for i := len(q.content)/2 - 1; i >= 0; i-- {
		q.percolateDown(i)
	}
}
