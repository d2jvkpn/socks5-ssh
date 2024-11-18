package bin

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func RunFileServer(args []string) {
	var (
		dir  string
		addr string
		err  error
		fSet *flag.FlagSet

		site       string
		fileServer http.Handler
	)

	fSet = flag.NewFlagSet("file_server", flag.ContinueOnError) // flag.ExitOnError

	fSet.StringVar(&addr, "addr", "127.0.0.1:1071", "file server listening address")
	fSet.StringVar(&dir, "dir", "./site", "local site directory path")
	fSet.StringVar(&site, "site", "/site", "http site path")

	fSet.Usage = func() {
		output := flag.CommandLine.Output()
		fmt.Fprintf(output, "Usage of file server:\n")
		fSet.PrintDefaults()
	}

	// fmt.Println("~~~ args:", args)
	if err = fSet.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "exit: %v\n", err)
		os.Exit(1)
		return
	}

	defer func() {
		if err != nil {
			log.Printf("Exit: %v\n", err)
			os.Exit(1)
		} else {
			log.Printf("Exit\n")
		}
	}()

	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = fmt.Errorf("Directory does not exist: %s", dir)
		return
	}

	fileServer = http.FileServer(http.Dir(dir))

	http.Handle(site, http.StripPrefix("/", fileServer))

	fmt.Printf(
		"==> Starting file server : dir=%q, address=%q, path=%s\n",
		dir, addr, site,
	)

	if err = http.ListenAndServe(addr, nil); err != nil {
		err = fmt.Errorf("Failed to start server: %w", err)
	}
}
