package version

import (
	"bytes"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/kylelemons/godebug/diff"
)

func Cue(ss ...string) cue.Value {
	return cuecontext.New().CompileString(strings.Join(ss, "\n"))
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
