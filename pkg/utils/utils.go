package utils

import (
	"os"
	"os/exec"
	"runtime"
)

func OpenBrowser(url string) error {
	var err error

	// Try to open the URL in different OS
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = os.ErrNotExist
	}

	return err
}
