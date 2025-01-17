package repl

import (
	//"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/abiosoft/ishell/v2"
	"github.com/flipez/rocket-lang/ast"
	"github.com/flipez/rocket-lang/evaluator"
	"github.com/flipez/rocket-lang/lexer"
	"github.com/flipez/rocket-lang/object"
	"github.com/flipez/rocket-lang/parser"
)

const PROMPT = ">> "

var buildVersion = "v0.10.0"
var buildDate = "2021-12-27T21:13:44Z"

func Start(in io.Reader, out io.Writer) {
	shell := ishell.New()
	shell.SetHomeHistoryPath(".rocket_history")
	shell.SetOut(out)
	shell.SetPrompt("🚀 > ")

	env := object.NewEnvironment()
	imports := make(map[string]struct{})

	shell.Println(SplashScreen())
	shell.NotFound(func(ctx *ishell.Context) {

		l := lexer.New(strings.Join(ctx.RawArgs, " "))
		p := parser.New(l, imports)

		var program *ast.Program

		program, imports = p.ParseProgram()
		if len(p.Errors()) > 0 {
			printParserErrors(ctx, p.Errors())
			return
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			ctx.Println("=> " + evaluated.Inspect())
		}
	})

	shell.Run()
}

const ROCKET = `
   /\
  (  )     ___         _       _   _
  (  )    | _ \___  __| |_____| |_| |   __ _ _ _  __ _
 /|/\|\   |   / _ \/ _| / / -_)  _| |__/ _  | ' \/ _  |
/_||||_\  |_|_\___/\__|_\_\___|\__|____\__,_|_||_\__, |
              %10s | %-15s   |___/
`

func SplashScreen() string {
	return fmt.Sprintf(ROCKET, buildVersion, buildDate)
}

func SplashVersion() string {
	return fmt.Sprintf("rocket-lang version %s (%s)\n", buildVersion, buildDate)
}

func printParserErrors(ctx *ishell.Context, errors []string) {
	ctx.Println("🔥 Great, you broke it!")
	ctx.Println(" parser errors:")
	for _, msg := range errors {
		ctx.Printf("\t %s\n", msg)
	}
}
