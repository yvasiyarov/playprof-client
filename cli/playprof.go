package main

import (
	"flag"
	"log"

	"github.com/yvasiyarov/playprof-client/profile"
)

var (
	appId int64
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.Int64Var(&appId, "app_id", 0, "Your app ID at playprof.com")
}

/*
func LoadSymbols(exe string) (*report.Reporter, error) {
	reporter := new(report.Reporter)
	err := reporter.SetExecutable(exe)
	return reporter, err
}
*/

/*
func PrintGraph(w io.Writer, r *report.Reporter, exe string) {
	col := report.ColCPU
	switch {
	case allocObj:
		col = report.ColAllocObj
	case allocSpace:
		col = report.ColAllocBytes
	case inuseObj:
		col = report.ColLiveObj
	case inuseSpace:
		col = report.ColLiveBytes
	}
	g := r.GraphByFunc(col)
	report := report.GraphReport{
		Prog:  exe,
		Total: r.Total(col),
		Unit:  "samples",
		Graph: g,

		NodeFrac: .005,
		EdgeFrac: .001,
	}
	report.WriteTo(w)
}
*/

func main() {
	flag.Parse()
	args := flag.Args()
	profile := profile.NewProfile()
	if len(args) == 1 {
		// pprof http://...
		if err := profile.ProfileByUrl(args[0], appId); err != nil {
			log.Fatalf("Error: %v", err)
		}
	} else {
		/*
			exe, prof := args[0], args[1]
			r, err := LoadSymbols(exe)
			if err != nil {
				log.Fatal(err)
			}
			f, err := os.Open(prof)
			if err != nil {
				log.Fatal(err)
			}
			LoadProfile(r, f)

			PrintGraph(os.Stdout, r, exe)
		*/
	}
}
