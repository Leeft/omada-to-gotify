package main_test

import (
	"bytes"
	"log"
	"os"
	"testing"

	main "github.com/leeft/omada-to-gotify"
)

func TestInitMain(t *testing.T) {

	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "logger: ", log.Lshortfile)
	)

	t.Run("GOTIFY_URL is required", func(t *testing.T) {
		buf.Reset()
		_, _, _, err := main.InitMain(logger)
		if err.Error() != "GOTIFY_URL environment variable is required" {
			logger.Fatalf("Failed test whether GOTIFY_URL is required; log is `%v`", buf.String())
		}
	})

	os.Setenv("GOTIFY_URL", "http://foo:1337/")

	t.Run("GOTIFY_APP_TOKEN is required", func(t *testing.T) {
		buf.Reset()
		_, _, _, err := main.InitMain(logger)
		if err.Error() != "GOTIFY_APP_TOKEN environment variable is required" {
			logger.Fatalf("Failed test whether GOTIFY_APP_TOKEN is required; log is `%v`", buf.String())
		}
	})

	os.Setenv("GOTIFY_APP_TOKEN", "foo")

	t.Run("OMADA_SHARED_SECRET is required", func(t *testing.T) {
		buf.Reset()
		_, _, _, err := main.InitMain(logger)
		if err.Error() != "OMADA_SHARED_SECRET environment variable is required" {
			logger.Fatalf("Failed test whether OMADA_SHARED_SECRET is required; log is `%v`", buf.String())
		}
	})

	os.Setenv("OMADA_SHARED_SECRET", "foo")

	t.Run("Can initialise after environment variables are set", func(t *testing.T) {
		buf.Reset()

		gotifyClient, server, port, err := main.InitMain(logger)

		if err != nil {
			logger.Fatalf("Still failed to initialize main; log is %v", buf.String())
		}

		if gotifyClient.GotifyURL != "http://foo:1337/" {
			logger.Fatalf("Failed to initialize gotify client properly; GotifyURL is `%v`", gotifyClient.GotifyURL)
		}

		if port != "8080" {
			logger.Fatalf("Failed to initialize server port; PORT is `%v`", port)
		}

		if server == nil {
			logger.Fatal("The server wasn't created by the init call")
		}
	})
}
