package main

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func getVisualisers() (visualisers []string) {
	if userBrowser := os.Getenv("BROWSER"); userBrowser != "" {
		visualisers = append(visualisers, userBrowser)
	}
	switch runtime.GOOS {
	case "darwin":
		visualisers = append(visualisers, "/usr/bin/open")
	case "windows":
		visualisers = append(visualisers, "cmd /c start")
	default:
		visualisers = append(visualisers, []string{"chrome", "google-chrome", "chromium", "firefox", "sensible-browser"}...)
		if os.Getenv("DISPLAY") != "" {
			// xdg-open is only for use in a desktop environment.
			visualisers = append(visualisers, "xdg-open")
		}
	}
	return
}

func browser(fileName string) error {
	for _, v := range getVisualisers() {
		args := strings.Split(v, " ")
		if len(args) == 0 {
			continue
		}
		viewer := exec.Command(args[0], append(args[1:], fileName)...)
		viewer.Stderr = os.Stderr
		err := viewer.Start()
		if err == nil {
			return viewer.Wait()
		}
	}
	return errors.New("cannot start the browser, you can open this link manually: " + fileName)
}
