package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/vybraan/vygrant/internal/config"
	"github.com/vybraan/vygrant/internal/storage"
	"golang.org/x/oauth2"
)

var LoadedAccounts map[string]*config.Account

const successHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>Authentication Successful</title>
<style>
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen,
      Ubuntu, Cantarell, "Open Sans", "Helvetica Neue", sans-serif;
    background: #f0f4f8;
    margin: 0;
    padding: 0;
    display: flex;
    height: 100vh;
    align-items: center;
    justify-content: center;
  }
  .container {
    background: white;
    padding: 2rem 3rem;
    border-radius: 10px;
    box-shadow: 0 4px 12px rgb(0 0 0 / 0.1);
    max-width: 400px;
    text-align: center;
  }
  .success-icon {
    font-size: 3rem;
    color: #4BB543; /* nice green */
    margin-bottom: 1rem;
  }
  h1 {
    margin: 0 0 1rem 0;
    font-weight: 600;
  }
  p {
    font-size: 1rem;
    color: #333;
    margin-bottom: 0;
  }
</style>
</head>
<body>
  <div class="container">
    <div class="success-icon"><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-badge-check-icon lucide-badge-check"><path d="M3.85 8.62a4 4 0 0 1 4.78-4.77 4 4 0 0 1 6.74 0 4 4 0 0 1 4.78 4.78 4 4 0 0 1 0 6.74 4 4 0 0 1-4.77 4.78 4 4 0 0 1-6.75 0 4 4 0 0 1-4.78-4.77 4 4 0 0 1 0-6.76Z"/><path d="m9 12 2 2 4-4"/></svg></div>
    <h1>Authentication Successful</h1>
    <p>Your account <strong>%s</strong> has been authenticated successfully.</p>
    <p>You can safely close this tab now.</p>
    <p><em>vygrant</em> will continue handling your authentication tokens in the background.</p>
  </div>
</body>
</html>
`

const errorHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>Error</title>
<style>
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen,
      Ubuntu, Cantarell, "Open Sans", "Helvetica Neue", sans-serif;
    background: #f8d7da;
    margin: 0;
    padding: 0;
    display: flex;
    height: 100vh;
    align-items: center;
    justify-content: center;
  }
  .container {
    background: #fff0f0;
    padding: 2rem 3rem;
    border-radius: 10px;
    box-shadow: 0 4px 12px rgb(0 0 0 / 0.1);
    max-width: 400px;
    text-align: center;
    border: 1px solid #f5c6cb;
  }
  .error-icon {
    font-size: 3rem;
    color: #dc3545; /* bootstrap danger red */
    margin-bottom: 1rem;
  }
  h1 {
    margin: 0 0 1rem 0;
    font-weight: 600;
    color: #721c24;
  }
  p {
    font-size: 1rem;
    color: #721c24;
    margin-bottom: 0;
  }
</style>
</head>
<body>
  <div class="container">
    <div class="error-icon">
<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-triangle-alert-icon lucide-triangle-alert"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"/><path d="M12 9v4"/><path d="M12 17h.01"/></svg>
</div>
    <h1>Error Occurred</h1>
    <p>%s</p>
    <p>Please try again or consider opening an <a href="https://github.com/vybraan/vygrant/issues/new/choose">GitHub Issue</a> if the problem persists.  </p>
  </div>
</body>
</html>
`

func StartAuthFlow(w http.ResponseWriter, r *http.Request) {
	accountName := r.URL.Query().Get("account")
	acct, ok := LoadedAccounts[accountName]
	if !ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, errorHTML, "Account '"+accountName+"' not found.")
		return
	}
	oauthCfg := config.GetOAuth2Config(acct)

	state := "account:" + accountName
	authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)

	if hint, ok := acct.AuthURIFields["login_hint"]; ok {
		authURL += "&login_hint=" + url.QueryEscape(hint)
	}

	http.Redirect(w, r, authURL, http.StatusFound)
}

func writeErrorPage(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, errorHTML, message)
}

func HandleOAuthCallback(tokenStore storage.TokenStore, httpClient *http.Client) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		state := r.URL.Query().Get("state")
		if !strings.HasPrefix(state, "account:") {
			writeErrorPage(w, http.StatusBadRequest, "Invalid state parameter.")
			return
		}
		accountName := strings.TrimPrefix(state, "account:")

		acct, ok := LoadedAccounts[accountName]
		if !ok {
			writeErrorPage(w, http.StatusBadRequest, "Invalid Account")
			return
		}
		oauthCfg := config.GetOAuth2Config(acct)
		ctx := context.Background()
		if httpClient != nil {
			ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
		}

		code := r.URL.Query().Get("code")
		token, err := oauthCfg.Exchange(ctx, code)
		if err != nil {

			writeErrorPage(w, http.StatusInternalServerError, "failed to exchange token. Please try again.")

			log.Printf("token exchange error for account %s: %v", accountName, err)
			return
		}

		tokenStore.Set(accountName, token)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, successHTML, accountName)
	}
}
