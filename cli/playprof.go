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
		if err := profile.LoadProfileFromFiles(args[0], args[1], appId); err != nil {
			log.Fatalf("Error: %v", err)
		}
	}
}
