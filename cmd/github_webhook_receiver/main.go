// github_webhook_receiver is a proof of concept for creating a github
// webhook that automatically adds and removes labels from issues.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/stephen-soltesz/github-webhook-poc/githubx/webhook"
	"github.com/stephen-soltesz/github-webhook-poc/local"

	// "github.com/kr/pretty"

	"golang.org/x/crypto/acme/autocert"
)

const (
	usage = `
USAGE:

  GitHub webhook receiver is configured through environment variables.

  Set the webhook secret to match the secret used to register the receiver:
  - GITHUB_WEBHOOK_SECRET

  For personal access token authentication:
  - GITHUB_AUTH_TOKEN

  For Github App authentication:
  - GITHUB_PRIVATE_KEY - the path to filename containing private key.
  - GITHUB_APP_ID - the application ID from registering the Github App.

  For Let's Encrypt TLS certificate, you may provide a hostname:
  - WEBHOOK_HOSTNAME

PERSONAL ACCESS TOKENS:

  Allocate a "Personal Access Token" by visiting github.com:
  - https://github.com/settings/tokens

REGISTER WEBHOOK:

  Register the webhook by visiting your repo on github.com. Click "Settings"
  and then "Webhooks". You should land on a URL like:

   * https://github.com/<owner>/<repo>/settings/hooks

  Click "Add Webhook".

  Use the payload URL (note the "/event_handler" path):

   * Payload URL: https://<service-url>/event_handler
   * Secret: value matching the environment variable GITHUB_WEBHOOK_SECRET
   * Select "Let me select individual events."
   * Check "Issues".
   * Uncheck "Pushes".
   * Click the green "Add Webhook" button.

  If the registration was successful, there should be a green checkmark. If
  registration failed, there will be a red "X".

FLAGS:

`
)

var (
	authToken     string
	webhookSecret string
	privateKey    string
	hostname      string
	fListenAddr   string
)

func init() {
	authToken = os.Getenv("GITHUB_AUTH_TOKEN")
	webhookSecret = os.Getenv("GITHUB_WEBHOOK_SECRET")
	privateKey = os.Getenv("GITHUB_PRIVATE_KEY")
	hostname = os.Getenv("WEBHOOK_HOSTNAME")
	flag.StringVar(&fListenAddr, "addr", ":3000", "The github user or organization name.")

	log.SetFlags(log.LstdFlags | log.LUTC | log.Lshortfile)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
}

func usageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", usage)
	flag.CommandLine.SetOutput(w)
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	if (authToken == "" && privateKey == "") || webhookSecret == "" {
		flag.Usage()
		os.Exit(1)
	}

	config := local.NewConfig(time.Second)

	eventHandler := &webhook.Handler{
		WebhookSecret:                 webhookSecret,
		IssuesEvent:                   config.IssuesEvent,
		InstallationEvent:             local.InstallationEvent,
		InstallationRepositoriesEvent: local.InstallationRepositoriesEvent,
		//ProjectCardEvent:              local.ProjectCardEvent,
		//ProjectColumnEvent:            local.ProjectColumnEvent,
		//ProjectEvent:                  local.ProjectEvent,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", usageHandler)
	mux.Handle("/event_handler", eventHandler)

	log.Println("Starting listeners")
	if hostname != "" {
		log.Fatal(http.Serve(autocert.NewListener(hostname), mux))
	} else {
		fmt.Println("Listening on ", fListenAddr)
		log.Fatal(http.ListenAndServe(fListenAddr, mux))
	}
}
