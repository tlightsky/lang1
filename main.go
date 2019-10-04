package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/tlightsky/lang1/parser"
)

func mapEval(l parser.SexpList, env string) parser.SexpList {
	var newList parser.SexpList
	for _, atom := range l {
		newList = append(newList, eval(atom, env))
	}
	return newList
}

// figure out value of each expression
func eval(exp parser.Sexp, env string) parser.Sexp {
	// exp.Dump(0)
	if l, isList := exp.I.(parser.SexpList); isList {
		op := eval(l[0], env)
		return apply(op, mapEval(l[1:], env))
	} else {
		// atom
		return exp
	}
}

// invoke op with arguments bind to env
func apply(op parser.Sexp, arguments parser.SexpList) parser.Sexp {
	operation := rune(op.I.(string)[0])
	result := 0.0
	var fn func(x, y float64) float64
	switch operation {
	case '+':
		fn = func(x, y float64) float64 { return x + y }
	case '-':
		fn = func(x, y float64) float64 { return x - y }
	case '*':
		fn = func(x, y float64) float64 { return x * y }
		result = 1.0
	case '/':
		fn = func(x, y float64) float64 { return x / y }
		result = 1.0
	default:
		panic(fmt.Errorf("unsupported op: %v", op))
	}

	for _, v := range arguments {
		switch vv := v.I.(type) {
		case float64:
			result = fn(result, vv)
		case int:
			result = fn(result, float64(vv))
		default:
			panic(fmt.Errorf("wrong atom: %v", v))
		}
	}
	return parser.Sexp{result}
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	for i := 0; ; i++ {
		fmt.Printf("%d =>", i)
		text, _ := reader.ReadString('\n')
		sexp, err := parser.ParseSexp(text)
		if err != nil {
			fmt.Println(err)
		}
		// sexp.Dump(0)
		result := eval(sexp, "")
		fmt.Printf("%s\n", result)
	}
}
