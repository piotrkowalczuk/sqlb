package sqlb

import (
	"bytes"
	"fmt"
	"strings"
)

var (
	bufferPool = newBuffer()
)

type Builder struct {
	first, last Expresser
	counter     *counter
	buffer      *bytes.Buffer
}

func New() *Builder {
	return &Builder{
		counter: &counter{},
		buffer:  bytes.NewBuffer(nil),
	}
}

func (b *Builder) add(left, right Expresser) *Builder {
	b.last.SetRight(left)
	b.last = right

	return b

}

// String implements fmt Stringer interface.
func (b *Builder) String() string {
	if b.buffer.Len() == 0 {
		b.first.WriteTo(b.buffer)
	}

	return strings.TrimSpace(b.buffer.String())
}

func (b *Builder) Reset() {
	b.buffer.Reset()
	b.counter.reset()
}

func (b *Builder) Insert() *Builder {
	i := &node{
		body: " INSERT ",
	}
	b.first = i
	b.last = i

	return b
}

func (b *Builder) Into(s string) *Builder {
	i := &node{
		body: " INTO ",
		right: &node{
			body: s,
		},
	}

	return b.add(i, i.right)
}

func (b *Builder) Columns(cols ...string) *Builder {
	nn := &namesNode{
		node: node{
			body: ", ",
			left: &node{
				body: " (",
			},
			right: &node{
				body: ") ",
			},
		},
		names: cols,
	}
	return b.add(nn, nn.right)
}

func (b *Builder) Values(args ...interface{}) *Builder {
	nodes := make([]Expresser, 0, len(args))
	for _, a := range args {
		nodes = append(nodes, &argNode{arg: a, counter: b.counter})
	}

	ln := &listNode{
		expressions: nodes,
		node: node{
			body: ", ",
			left: &node{
				body: " VALUES(",
			},
			right: &node{
				body: ") ",
			},
		},
	}

	return b.add(ln, ln.right.Right())
}

func (b *Builder) Select(expr Expresser) *Builder {
	s := &node{
		body:  " SELECT ",
		right: expr,
	}
	b.first = s
	b.last = expr

	return b
}

func (b *Builder) From(expr Expresser) *Builder {
	f := &node{
		body:  " FROM ",
		right: expr,
	}

	b.last.SetRight(f)
	b.last = expr

	return b
}

func (b *Builder) Where(expr Expresser) *Builder {
	f := &node{
		body:  " WHERE ",
		right: expr,
	}
	b.last.SetRight(f)
	b.last = expr

	return b
}

func (b *Builder) Limit(l int64) *Builder {
	limit := &node{
		body: " LIMIT ",
		right: &argNode{
			arg:     l,
			counter: b.counter,
		},
	}
	b.last.SetRight(limit)
	b.last = limit.right

	return b
}

func (b *Builder) Offset(o int64) *Builder {
	offset := &node{
		body: " OFFSET ",
		right: &argNode{
			arg:     o,
			counter: b.counter,
		},
	}
	b.last.SetRight(offset)
	b.last = offset.right

	return b
}

func (b *Builder) GroupBy(names ...string) *Builder {
	offset := &node{
		body: " GROUP BY ",
		right: &namesNode{
			names: names,
		},
	}
	b.last.SetRight(offset)
	b.last = offset.right

	return b
}

func (b *Builder) Having(expr Expresser) *Builder {
	f := &node{
		body:  " HAVING ",
		right: expr,
	}
	b.last.SetRight(f)
	b.last = expr

	return b
}

func (b *Builder) OrderBy(names ...string) *Builder {
	offset := &node{
		body: " ORDER BY ",
		right: &namesNode{
			names: names,
		},
	}
	b.last.SetRight(offset)
	b.last = offset.right

	return b
}

func (b *Builder) Distinct() *Builder {
	offset := &node{
		body: " DISTINCT ",
	}
	b.last.SetRight(offset)
	b.last = offset

	return b
}

func (b *Builder) Arg(arg interface{}) Expresser {
	return &argNode{
		arg:     arg,
		counter: b.counter,
	}
}

// Args ...
func (b *Builder) Args() []interface{} {
	args := make([]interface{}, 0, b.counter.index)
	return b.args(b.first, args)
}

func (b *Builder) Expr() Expresser {
	return b.first
}

func (b *Builder) args(expr Expresser, args []interface{}) []interface{} {
	if expr.Left() != nil {
		args = b.args(expr.Left(), args)
	}
	switch et := expr.(type) {
	case *argNode:
		args = append(args, et.arg)
	case *listNode:
		for _, e := range et.expressions {
			args = b.args(e, args)
		}
	}
	if expr.Right() != nil {
		args = b.args(expr.Right(), args)
	}
	return args
}

// GoString ...
func (b *Builder) GoString() string {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, "query: %s \n", b.String())
	for i, a := range b.Args() {
		fmt.Fprintf(buf, "arg: %-5d %#v \n", i+1, a)
	}
	return buf.String()
}
