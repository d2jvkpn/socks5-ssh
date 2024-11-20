package bin

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func RunFileServer(args []string) {
	var (
		dir  string
		addr string
		err  error
		fSet *flag.FlagSet

		site       string
		logger     *slog.Logger
		fileServer http.Handler
	)

	fSet = flag.NewFlagSet("file_server", flag.ContinueOnError) // flag.ExitOnError

	fSet.StringVar(&addr, "addr", "127.0.0.1:1099", "file server listening address")
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

	logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))

	defer func() {
		if err != nil {
			logger.Error("Exit", "error", err)
			os.Exit(1)
		} else {
			logger.Info("Exit")
		}
	}()

	if dir, err = filepath.Abs(dir); err != nil {
		err = fmt.Errorf("Can't get absolute path of %s", dir)
		return
	}

	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = fmt.Errorf("Directory does not exist: %s", dir)
		return
	}

	fileServer = http.FileServer(http.Dir(dir))

	site = "/" + strings.Trim(site, "/")
	// fmt.Println("~~~", site)
	http.Handle(site+"/", http.StripPrefix(site, fileServer))

	logger.Info(
		"==> Starting file server",
		"directory", dir,
		"address", addr,
		"site", site,
	)

	if err = http.ListenAndServe(addr, nil); err != nil {
		err = fmt.Errorf("Failed to start server: %w", err)
	}
}
