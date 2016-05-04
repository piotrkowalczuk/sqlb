package sqlb

import (
	"fmt"
	"testing"
)

var (
	benchBuilder *Builder
	benchArgs    []interface{}
	benchString  string
)

func ExampleBuilder_Select() {
	b := New()
	b = b.Select(Name("*")).From(Name("user")).Where(
		And(
			Equal(Name("updated_at"), b.Arg(5)),
			Is(Name("updated_at"), NotNull()),
			Contains(Name("roles"), b.Arg(map[string]interface{}{"manager": 1})),
			IsContainedBy(Name("something"), b.Arg(`'{"a":1, "b":2}'`)),
			HasKey(Name("something"), b.Arg("key")),
		),
	)

	fmt.Println(b.String())
	fmt.Println(b.Args())

	// Output:
	// SELECT * FROM user WHERE  (updated_at = $1 AND updated_at IS  NOT NULL  AND roles @> $2 AND something <@ $3 AND something ? $4)
	// [5 map[manager:1] '{"a":1, "b":2}' key]
}

func ExampleBuilder_Insert() {
	b := New()
	b = b.Insert().Into("user").Columns("username", "first_name", "last_name").Values("john.snow@gmail.com", "John", "Snow")

	fmt.Println(b.String())
	for _, a := range b.Args() {
		fmt.Println(a)
	}

	// Output:
	// INSERT  INTO user (username, first_name, last_name)  VALUES($1, $2, $3)
	// john.snow@gmail.com
	// John
	// Snow
}

func BenchmarkBuilder_general(b *testing.B) {
	var (
		bb *Builder
		s  string
		a  []interface{}
	)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		bb = New()
		bb.Select(Name("*")).
			From(Name("user")).Where(
			And(
				Equal(Name("updated_at"), bb.Arg(5)),
				Is(Name("updated_at"), NotNull()),
				Contains(Name("roles"), bb.Arg(map[string]interface{}{"manager": 1})),
				IsContainedBy(Name("something"), bb.Arg(`'{"a":1, "b":2}'`)),
			),
		)
		s = bb.String()
		a = bb.Args()
	}
	benchString = s
	benchArgs = a
}

func benchmarkBuilder(b *testing.B) (bb *Builder) {
	bb = New().Select(Name("*")).From(Name("table"))
	exprs := make([]Expresser, 0, b.N)
	for n := 0; n < b.N; n++ {
		exprs = append(exprs, Equal(Name("column"), bb.Arg(n)))
	}
	bb.Where(And(exprs...))
	return bb
}

func BenchmarkBuilder_String(b *testing.B) {
	var s string
	bb := benchmarkBuilder(b)
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		s = bb.String()
	}
	benchString = s
}

func BenchmarkBuilder_Args(b *testing.B) {
	var args []interface{}
	bb := benchmarkBuilder(b)
	bb.String()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		args = bb.Args()
	}
	benchArgs = args
}

func BenchmarkBuilder_Select(b *testing.B) {
	expr := Name("*")
	benchBuilder = New()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchBuilder.Select(expr)
	}
}

func BenchmarkBuilder_Select_simple(b *testing.B) {
	bb := New()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchString = bb.Select(Name("id")).
			From(Name("tickets")).
			Where(
				And(
					Equal(Name("subdomain_id"), bb.Arg(1)),
					Or(
						Equal(Name("state"), bb.Arg("open")),
						Equal(Name("state"), bb.Arg("spam")),
					),
				),
			).String()

		b.StopTimer()
		bb.Reset()
		b.StartTimer()
	}
}

func BenchmarkBuilder_Select_conditional(b *testing.B) {
	bb := New()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		bb = bb.Select(Name("id")).
			From(Name("tickets")).
			Where(
				And(
					Equal(Name("subdomain_id"), bb.Arg(1)),
					Or(
						Equal(Name("state"), bb.Arg("open")),
						Equal(Name("state"), bb.Arg("spam")),
					),
				),
			)

		if n%2 == 0 {
			bb.GroupBy("subdomain_id").
				Having(Equal(Name("number"), bb.Arg(1))).
				OrderBy("state").
				Limit(7).
				Offset(8)
		}

		benchString = bb.String()
		b.StopTimer()
		bb.Reset()
		b.StartTimer()
	}
}

func BenchmarkBuilder_Select_complex(b *testing.B) {
	bb := New()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchString = bb.Select(Names("a", "b", "z", "y", "x")).
			Distinct().
			From(Name("c")).
			Where(
				And(
					Or(
						Equal(Name("d"), bb.Arg(1)),
						Equal(Name("e"), bb.Arg("wat")),
					),
					Equal(Name("f"), bb.Arg(2)),
					Equal(Name("x"), bb.Arg("hi")),
					Equal(Name("g"), bb.Arg(3)),
					In(Name("h"), List(bb.Arg(1), bb.Arg(2), bb.Arg(3))),
				),
			).
			GroupBy("i", "ii", "iii").
			Having(
				And(
					Equal(Name("j"), Name("k")),
					Equal(Name("jj"), bb.Arg(1)),
					Equal(Name("jjj"), bb.Arg(2)),
				),
			).
			OrderBy("l", "lll", "lll").
			Limit(7).
			Limit(8).
			String()

		b.StopTimer()
		bb.Reset()
		b.StartTimer()
	}
}

func BenchmarkBuilder_Select_subquery(b *testing.B) {
	bb := New()
	bbs := New()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		bbs = New().Select(Name("id")).From(Name("tickets")).Where(
			And(
				Equal(Name("subdomain_id"), bb.Arg(1)),
				Or(
					Equal(Name("state"), bb.Arg("open")),
					Equal(Name("state"), bb.Arg("spam")),
				),
			),
		)

		benchString = bb.Select(List(Name("a"), Name("b"), As(bbs.Expr(), "subq"))).
			From(Name("c")).
			Distinct().
			Where(
				And(
					Equal(Name("f"), bb.Arg(2)),
					Equal(Name("x"), bb.Arg("hi")),
					Equal(Name("g"), bb.Arg(3)),
				),
			).
			OrderBy("l", "l").
			Limit(7).
			Offset(8).
			String()

		b.StopTimer()
		bb.Reset()
		b.StartTimer()
	}
}

func BenchmarkBuilder_Insert(b *testing.B) {
	b.SkipNow()
	//bb := New()
	//
	//b.ResetTimer()
	//for n := 0; n < b.N; n++ {
	//
	//}
}
