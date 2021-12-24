package cmd

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

// Utility function to return a default file name for deployment YAML
func getDefaultFileName() string {
	const layout = "2006-01-02"
	t := time.Now()
	return "k8s-deployment-" + t.Format(layout)
}

// Utility function to prompt user to confirm
func askForConfirmation() bool {
	var response string

	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}

	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		fmt.Println("Please type (y)es or (n)o and then press enter:")
		return askForConfirmation()
	}
}

// Utility function to sanitize a string for K8s
func sanitizeValue(specialString string) string {
	re, err := regexp.Compile(`[&\/\\#,+()$~%.'":*?<>{}@]`)
	if err != nil {
		log.Fatal(err)
	}
	return re.ReplaceAllString(specialString, "-")
}
