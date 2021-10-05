package version

import (
	"bytes"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"github.com/kylelemons/godebug/diff"
)

func Cue(ss ...string) cue.Value {
	return cuecontext.New().CompileString(strings.Join(ss, "\n"))
}

func logCueErrors(err error) {
	// fmt.Printf("%#v\n", cueErr.Position().String())
	// for _, pos := range errors.Positions(cueErr) {
	// 	fmt.Printf("%#v\n", pos.String())
	// }
	// format, args := cueErr.Msg()
	// fmt.Printf(format+"\n", args...)
	for _, e := range errors.Errors(err.(errors.Error)) {
		// fmt.Printf("%#v\n", e.Position())
		// fmt.Printf("%#v\n", e.InputPositions())
		// format, args := e.Msg()
		// fmt.Printf(format, args...)
		logger.Error(e, e.Position().String())
	}
}

func CueDiff(a, b cue.Value) {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	aLines := strings.Split(aStr, "\n")
	bLines := strings.Split(bStr, "\n")

	chunks := diff.DiffChunks(aLines, bLines)

	buf := new(bytes.Buffer)
	for _, c := range chunks {
		for _, d := range c.Added {
			fmt.Fprintf(buf, "\033[32m+%s\n", d)
		}
		for _, d := range c.Deleted {
			fmt.Fprintf(buf, "\033[31m-%s\n", d)
		}
		fmt.Fprintf(buf, "\033[1;34m\033[0m")
	}
	fmt.Println(strings.TrimRight(buf.String(), "\n"))
}
