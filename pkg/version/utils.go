package version

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/kylelemons/godebug/diff"
)

func CueFromStrings(ss ...string) (cue.Value, error) {
	if len(ss) == 0 {
		return cue.Value{}, fmt.Errorf("no strings provided")
	}

	cwd, _ := os.Getwd()
	abs := func(path string) string {
		return filepath.Join(cwd, path)
	}
	var cueArgs []string
	overlays := make(map[string]load.Source)
	for idx, s := range ss {
		mock := fmt.Sprintf("./mock/%d.cue", idx)
		cueArgs = append(cueArgs, mock)
		overlays[abs(mock)] = load.FromString(s)
	}

	ctx := cuecontext.New()
	value := cue.Value{}
	cfg := &load.Config{Overlay: overlays}
	for idx, i := range load.Instances(cueArgs, cfg) {
		if i.Err != nil {
			return cue.Value{}, fmt.Errorf("load error [%d]: %w", idx, i.Err)
		}
		v := ctx.BuildInstance(i)
		if err := v.Err(); err != nil {
			return v, fmt.Errorf("value error [%d]: %w", idx, err)
		}
		value = value.Unify(v)
	}

	return value, nil
}

func PrintCueDiff(a, b cue.Value) {
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
