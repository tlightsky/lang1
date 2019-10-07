package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/tlightsky/lang1/parser"
)

type (
	Fn struct {
		params []string
		body   []parser.Sexp
		env    *Environment
	}

	Environment struct {
		Env       map[string]parser.Sexp
		ParentEnv *Environment
	}
)

func mapEval(l parser.SexpList, env *Environment) parser.SexpList {
	var newList parser.SexpList
	for _, atom := range l {
		newList = append(newList, eval(atom, env))
	}
	return newList
}

// figure out value of each expression
func eval(exp parser.Sexp, env *Environment) parser.Sexp {
	// exp.Dump(0)

	if l, isList := exp.I.(parser.SexpList); isList {
		// variable?
		// quoted?
		// assignment?
		// definition?
		// if?

		// lambda?
		if isDefn(l[0]) {
			args := l[1:]
			if len(args) < 3 {
				return parser.Sexp{errors.New("argument too less for defn")}
			}
			return evalDefn(args[0].I.(string), args[1].I.(parser.SexpList), args[2:], env)
		}
		op := eval(l[0], env)
		return apply(op, mapEval(l[1:], env))
	} else {
		if key, ok := exp.I.(string); ok {
			if val, ok := getEnv(key, env); ok {
				return val
			}
		}
		// self eval?
		// atom
		return exp
		// return parser.Sexp{errors.New(fmt.Sprintf("wrong eval", exp))
	}
}

func evalSequence(body parser.SexpList, env *Environment) (result parser.Sexp) {
	for _, exp := range body {
		result = eval(exp, env)
	}
	return
	// return parser.Sexp{errors.New(fmt.Sprintf("wrong evalSequence", body))}
}

func isPrimitiveOp(op parser.Sexp) bool {
	operation, ok := op.I.(string)
	if !ok {
		return false
	}
	switch rune(operation[0]) {
	case '+':
		return true
	case '-':
		return true
	case '*':
		return true
	case '/':
		return true
	}
	return false
}

func isDefn(op parser.Sexp) bool {
	if operation, ok := op.I.(string); ok {
		if operation == "defn" {
			return true
		}
	}
	return false
}

func evalDefn(name string, fnParams parser.SexpList, body parser.SexpList, env *Environment) parser.Sexp {
	var paramSybols []string
	for _, v := range fnParams {
		paramSybols = append(paramSybols, v.I.(string))
	}
	fn := parser.Sexp{Fn{paramSybols, body, env}}
	env.Env[name] = fn
	return fn
}

func getFn(op parser.Sexp, env Environment) *Fn {
	if operation, ok := op.I.(string); ok {
		if fnExp, ok := env.Env[operation]; ok {
			if fn, ok := fnExp.I.(Fn); ok {
				return &fn
			}
		}
	}
	return nil
}

// func applyFn(fn *Fn, args parser.SexpList) {
// 	// eval body in new environment
// 	newEnv := extendEnv(env, fn.arguments, args)
// 	return eval(fn.body, newEnv)
// }

func extendEnv(parentEnv *Environment, params []string, arguemtns parser.SexpList) *Environment {
	env := newEnv(parentEnv)
	for i := range params {
		env.Env[params[i]] = arguemtns[i]
	}
	return env
}

func newEnv(parentEnv *Environment) *Environment {
	return &Environment{
		Env:       make(map[string]parser.Sexp),
		ParentEnv: parentEnv,
	}
}

func getEnv(name string, env *Environment) (parser.Sexp, bool) {
	for {
		if val, ok := env.Env[name]; ok {
			return val, true
		}
		env = env.ParentEnv
		if env == nil {
			break
		}
	}
	return parser.Sexp{}, false
}

func applyStringJoin(args parser.SexpList) parser.Sexp {
	var sb strings.Builder
	for _, str := range args {
		sb.WriteString(string(str.I.(parser.QString)))
	}
	return parser.Sexp{sb.String()}
}

func applyPrimitiveOp(op parser.Sexp, arguments parser.SexpList) parser.Sexp {
	operation := rune(op.I.(string)[0])
	result := 0.0
	var fn func(x, y float64) float64
	switch operation {
	case '+':
		if len(arguments) > 1 && reflect.TypeOf(arguments[0].I).Kind() == reflect.String {
			return applyStringJoin(arguments)
		}
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

// invoke op with arguments bind to env
func apply(op parser.Sexp, args parser.SexpList) parser.Sexp {
	if isPrimitiveOp(op) {
		return applyPrimitiveOp(op, args)
	}

	// if is fn, value is Fn
	// call an fn
	if fn, ok := op.I.(Fn); ok {
		return evalSequence(fn.body, extendEnv(fn.env, fn.params, args))
	}

	// match fn
	return parser.Sexp{errors.New(fmt.Sprintf("unknown apply method: %s", op))}
}

func main() {
	globalEnv := newEnv(nil)

	reader := bufio.NewReader(os.Stdin)
	for i := 0; ; i++ {
		fmt.Printf("%d =>", i)
		text, _ := reader.ReadString('\n')
		sexp, err := parser.ParseSexp(text)
		if err != nil {
			fmt.Println(err)
		}
		// sexp.Dump(0)
		result := eval(sexp, globalEnv)
		fmt.Printf("%s\n", result)
	}
}
