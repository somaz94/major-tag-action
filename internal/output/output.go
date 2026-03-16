package output

import (
	"fmt"
	"os"
	"strings"
)

// SetOutput sets a GitHub Actions output variable.
func SetOutput(name, value string) error {
	outputFile := os.Getenv("GITHUB_OUTPUT")
	if outputFile == "" {
		fmt.Printf("::set-output name=%s::%s\n", name, value)
		return nil
	}

	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open GITHUB_OUTPUT: %w", err)
	}
	defer f.Close()

	if strings.Contains(value, "\n") {
		delimiter := "EOF"
		_, err = fmt.Fprintf(f, "%s<<%s\n%s\n%s\n", name, delimiter, value, delimiter)
	} else {
		_, err = fmt.Fprintf(f, "%s=%s\n", name, value)
	}

	return err
}

// LogInfo prints an info message in GitHub Actions format.
func LogInfo(msg string) {
	fmt.Printf("::notice::%s\n", msg)
}

// LogWarning prints a warning message in GitHub Actions format.
func LogWarning(msg string) {
	fmt.Printf("::warning::%s\n", msg)
}

// LogError prints an error message in GitHub Actions format.
func LogError(msg string) {
	fmt.Printf("::error::%s\n", msg)
}
