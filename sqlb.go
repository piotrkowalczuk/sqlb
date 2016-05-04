package sqlb

import (
	"bytes"
	"strconv"
)

// Expresser ...
type Expresser interface {
	Left() Expresser
	Right() Expresser
	SetLeft(Expresser)
	SetRight(Expresser)
	WriteTo(*bytes.Buffer) (int64, error)
}

type node struct {
	left, right Expresser
	body        string
}

// String implements fmt Stringer interface.
func (n *node) String() string {
	return n.body
}

// Name ...
func Name(s string) Expresser {
	return &node{
		body: s,
	}
}

// Left ...
func (n *node) Left() Expresser {
	return n.left
}

// Right ...
func (n *node) Right() Expresser {
	return n.right
}

// SetLeft ...
func (n *node) SetLeft(e Expresser) {
	n.left = e
}

// SetRight ...
func (n *node) SetRight(e Expresser) {
	n.right = e
}

// WriteTo ...
func (n *node) WriteTo(b *bytes.Buffer) (int64, error) {
	var err error
	if n.left != nil {
		if _, err = n.left.WriteTo(b); err != nil {
			return 0, err
		}
	}
	b.WriteString(n.body)
	if n.right != nil {
		if _, err = n.right.WriteTo(b); err != nil {
			return 0, err
		}
	}
	return 0, nil
}

func Is(left, right Expresser) Expresser {
	return &node{
		body:  " IS ",
		right: right,
		left:  left,
	}
}

// NotNull ...
func NotNull() Expresser {
	return &node{
		body: " NOT NULL ",
	}
}

// Equal ...
func Equal(left, right Expresser) Expresser {
	return &node{
		body:  " = ",
		left:  left,
		right: right,
	}
}

func Contains(left, right Expresser) Expresser {
	return &node{
		body:  " @> ",
		left:  left,
		right: right,
	}
}

func IsContainedBy(left, right Expresser) Expresser {
	return &node{
		body:  " <@ ",
		left:  left,
		right: right,
	}
}

func In(left, right Expresser) Expresser {
	return &node{
		body:  " IN ",
		left:  left,
		right: right,
	}
}

func As(left Expresser, right string) Expresser {
	return &node{
		body: " AS ",
		left: left,
		right: &node{
			body: right,
		},
	}
}

func HasKey(left, right Expresser) Expresser {
	return &node{
		body:  " ? ",
		left:  left,
		right: right,
	}
}

type argNode struct {
	node
	counter *counter
	arg     interface{}
}

func (an *argNode) WriteTo(b *bytes.Buffer) (int64, error) {
	if _, err := b.WriteString("$"); err != nil {
		return 0, err
	}
	if _, err := b.WriteString(strconv.FormatInt(an.counter.get(), 10)); err != nil {
		return 0, err
	}
	return 0, nil
}

type listNode struct {
	node
	expressions []Expresser
}

func (ln *listNode) WriteTo(b *bytes.Buffer) (int64, error) {
	if ln.left != nil {
		if _, err := ln.left.WriteTo(b); err != nil {
			return 0, err
		}
	}
	for i, e := range ln.expressions {
		if i > 0 {
			b.WriteString(ln.body)
		}
		if _, err := e.WriteTo(b); err != nil {
			return 0, err
		}
	}
	if ln.right != nil {
		if _, err := ln.right.WriteTo(b); err != nil {
			return 0, err
		}
	}

	return 0, nil
}

func And(exprs ...Expresser) Expresser {
	return &listNode{
		node: node{
			body: " AND ",
			left: &node{
				body: " (",
			},
			right: &node{
				body: ") ",
			},
		},
		expressions: exprs,
	}
}

func Or(exprs ...Expresser) Expresser {
	return &listNode{
		node: node{
			body: " OR ",
			left: &node{
				body: " (",
			},
			right: &node{
				body: ") ",
			},
		},
		expressions: exprs,
	}
}

// List ...
func List(exprs ...Expresser) Expresser {
	return &listNode{
		node: node{
			body: ", ",
		},
		expressions: exprs,
	}
}

type namesNode struct {
	node
	names []string
}

// Names ...
func Names(names ...string) Expresser {
	return &namesNode{
		names: names,
	}
}

// WriteTo ...
func (nn *namesNode) WriteTo(b *bytes.Buffer) (int64, error) {
	if nn.left != nil {
		if _, err := nn.left.WriteTo(b); err != nil {
			return 0, err
		}
	}
	for i, n := range nn.names {
		if i > 0 {
			b.WriteString(nn.body)
		}
		b.WriteString(n)
	}
	if nn.right != nil {
		if _, err := nn.right.WriteTo(b); err != nil {
			return 0, err
		}
	}

	return 0, nil
}
