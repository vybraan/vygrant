package auth

import (
	"context"
	"fmt"
	"html"
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
<title>vygrant · Authentication Complete</title>
<style>
  :root {
    color-scheme: light dark;

    --bg: #f6f8fb;
    --bg-soft: #eef2f7;
    --card: rgba(255, 255, 255, 0.88);
    --card-border: rgba(15, 23, 42, 0.08);
    --text: #0f172a;
    --muted: #64748b;
    --muted-strong: #334155;
    --success: #16a34a;
    --success-soft: #dcfce7;
    --success-border: #bbf7d0;
    --shadow: 0 24px 80px rgba(15, 23, 42, 0.12);
  }

  @media (prefers-color-scheme: dark) {
    :root {
      --bg: #020617;
      --bg-soft: #0f172a;
      --card: rgba(15, 23, 42, 0.78);
      --card-border: rgba(148, 163, 184, 0.16);
      --text: #e5e7eb;
      --muted: #94a3b8;
      --muted-strong: #cbd5e1;
      --success: #22c55e;
      --success-soft: rgba(34, 197, 94, 0.12);
      --success-border: rgba(34, 197, 94, 0.28);
      --shadow: 0 24px 80px rgba(0, 0, 0, 0.42);
    }
  }

  * {
    box-sizing: border-box;
  }

  html, body {
    height: 100%%;
  }

  body {
    margin: 0;
    font-family:
      Inter,
      ui-sans-serif,
      system-ui,
      -apple-system,
      BlinkMacSystemFont,
      "Segoe UI",
      sans-serif;
    background:
      radial-gradient(circle at top left, rgba(34, 197, 94, 0.16), transparent 32rem),
      radial-gradient(circle at bottom right, rgba(59, 130, 246, 0.12), transparent 30rem),
      linear-gradient(135deg, var(--bg), var(--bg-soft));
    color: var(--text);
    display: grid;
    place-items: center;
    padding: 24px;
  }

  .shell {
    width: 100%%;
    max-width: 460px;
  }

  .brand {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    margin-bottom: 18px;
    color: var(--muted);
    font-size: 13px;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }

  .brand-mark {
    width: 9px;
    height: 9px;
    border-radius: 999px;
    background: var(--success);
    box-shadow: 0 0 0 6px var(--success-soft);
  }

  .card {
    position: relative;
    overflow: hidden;
    background: var(--card);
    border: 1px solid var(--card-border);
    border-radius: 24px;
    box-shadow: var(--shadow);
    backdrop-filter: blur(18px);
    padding: 34px;
    text-align: center;
  }

  .card::before {
    content: "";
    position: absolute;
    inset: 0;
    height: 4px;
    background: linear-gradient(90deg, transparent, var(--success), transparent);
  }

  .icon {
    width: 64px;
    height: 64px;
    margin: 0 auto 22px;
    border-radius: 20px;
    display: grid;
    place-items: center;
    color: var(--success);
    background: var(--success-soft);
    border: 1px solid var(--success-border);
  }

  .icon svg {
    width: 34px;
    height: 34px;
  }

  .status {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 14px;
    padding: 7px 12px;
    border-radius: 999px;
    color: var(--success);
    background: var(--success-soft);
    border: 1px solid var(--success-border);
    font-size: 13px;
    font-weight: 650;
  }

  h1 {
    margin: 0;
    font-size: clamp(26px, 5vw, 34px);
    line-height: 1.08;
    letter-spacing: -0.04em;
    font-weight: 760;
  }

  p {
    margin: 14px 0 0;
    color: var(--muted);
    font-size: 15px;
    line-height: 1.65;
  }

  strong {
    color: var(--muted-strong);
    font-weight: 700;
  }

  .account {
    margin-top: 20px;
    padding: 12px 14px;
    border-radius: 14px;
    background: rgba(148, 163, 184, 0.12);
    border: 1px solid var(--card-border);
    color: var(--muted-strong);
    font-family:
      ui-monospace,
      SFMono-Regular,
      Menlo,
      Monaco,
      Consolas,
      "Liberation Mono",
      monospace;
    font-size: 13px;
    overflow-wrap: anywhere;
  }

  .footer {
    margin-top: 24px;
    padding-top: 20px;
    border-top: 1px solid var(--card-border);
    color: var(--muted);
    font-size: 13px;
  }

  .footer code {
    color: var(--muted-strong);
    background: rgba(148, 163, 184, 0.14);
    padding: 2px 6px;
    border-radius: 7px;
    font-family:
      ui-monospace,
      SFMono-Regular,
      Menlo,
      Monaco,
      Consolas,
      "Liberation Mono",
      monospace;
  }
</style>
</head>
<body>
  <main class="shell">
    <div class="brand">
      <span class="brand-mark"></span>
      <span>vygrant authorization broker</span>
    </div>

    <section class="card" aria-labelledby="title">
      <div class="icon" aria-hidden="true">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.25" stroke-linecap="round" stroke-linejoin="round">
          <path d="M3.85 8.62a4 4 0 0 1 4.78-4.77 4 4 0 0 1 6.74 0 4 4 0 0 1 4.78 4.78 4 4 0 0 1 0 6.74 4 4 0 0 1-4.77 4.78 4 4 0 0 1-6.75 0 4 4 0 0 1-4.78-4.77 4 4 0 0 1 0-6.76Z"/>
          <path d="m9 12 2 2 4-4"/>
        </svg>
      </div>

      <div class="status">Authentication complete</div>

      <h1 id="title">You are signed in.</h1>

      <p>
        The account below has been linked successfully. You can close this browser tab and return to your terminal.
      </p>

      <div class="account">%s</div>

      <p class="footer">
        <code>vygrant</code> will keep handling token refresh in the background.
      </p>
    </section>
  </main>
</body>
</html>
`

const errorHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>vygrant · Authentication Failed</title>
<style>
  :root {
    color-scheme: light dark;

    --bg: #f8fafc;
    --bg-soft: #f1f5f9;
    --card: rgba(255, 255, 255, 0.9);
    --card-border: rgba(15, 23, 42, 0.08);
    --text: #0f172a;
    --muted: #64748b;
    --muted-strong: #334155;
    --danger: #dc2626;
    --danger-soft: #fee2e2;
    --danger-border: #fecaca;
    --link: #2563eb;
    --shadow: 0 24px 80px rgba(15, 23, 42, 0.12);
  }

  @media (prefers-color-scheme: dark) {
    :root {
      --bg: #020617;
      --bg-soft: #111827;
      --card: rgba(15, 23, 42, 0.8);
      --card-border: rgba(148, 163, 184, 0.16);
      --text: #e5e7eb;
      --muted: #94a3b8;
      --muted-strong: #cbd5e1;
      --danger: #f87171;
      --danger-soft: rgba(248, 113, 113, 0.12);
      --danger-border: rgba(248, 113, 113, 0.28);
      --link: #60a5fa;
      --shadow: 0 24px 80px rgba(0, 0, 0, 0.42);
    }
  }

  * {
    box-sizing: border-box;
  }

  html, body {
    height: 100%%;
  }

  body {
    margin: 0;
    font-family:
      Inter,
      ui-sans-serif,
      system-ui,
      -apple-system,
      BlinkMacSystemFont,
      "Segoe UI",
      sans-serif;
    background:
      radial-gradient(circle at top left, rgba(220, 38, 38, 0.14), transparent 32rem),
      radial-gradient(circle at bottom right, rgba(59, 130, 246, 0.10), transparent 30rem),
      linear-gradient(135deg, var(--bg), var(--bg-soft));
    color: var(--text);
    display: grid;
    place-items: center;
    padding: 24px;
  }

  .shell {
    width: 100%%;
    max-width: 500px;
  }

  .brand {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    margin-bottom: 18px;
    color: var(--muted);
    font-size: 13px;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }

  .brand-mark {
    width: 9px;
    height: 9px;
    border-radius: 999px;
    background: var(--danger);
    box-shadow: 0 0 0 6px var(--danger-soft);
  }

  .card {
    position: relative;
    overflow: hidden;
    background: var(--card);
    border: 1px solid var(--card-border);
    border-radius: 24px;
    box-shadow: var(--shadow);
    backdrop-filter: blur(18px);
    padding: 34px;
    text-align: center;
  }

  .card::before {
    content: "";
    position: absolute;
    inset: 0;
    height: 4px;
    background: linear-gradient(90deg, transparent, var(--danger), transparent);
  }

  .icon {
    width: 64px;
    height: 64px;
    margin: 0 auto 22px;
    border-radius: 20px;
    display: grid;
    place-items: center;
    color: var(--danger);
    background: var(--danger-soft);
    border: 1px solid var(--danger-border);
  }

  .icon svg {
    width: 34px;
    height: 34px;
  }

  .status {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 14px;
    padding: 7px 12px;
    border-radius: 999px;
    color: var(--danger);
    background: var(--danger-soft);
    border: 1px solid var(--danger-border);
    font-size: 13px;
    font-weight: 650;
  }

  h1 {
    margin: 0;
    font-size: clamp(26px, 5vw, 34px);
    line-height: 1.08;
    letter-spacing: -0.04em;
    font-weight: 760;
  }

  p {
    margin: 14px 0 0;
    color: var(--muted);
    font-size: 15px;
    line-height: 1.65;
  }

  .error-box {
    margin-top: 20px;
    padding: 14px 16px;
    border-radius: 14px;
    background: rgba(148, 163, 184, 0.12);
    border: 1px solid var(--card-border);
    color: var(--muted-strong);
    font-family:
      ui-monospace,
      SFMono-Regular,
      Menlo,
      Monaco,
      Consolas,
      "Liberation Mono",
      monospace;
    font-size: 13px;
    line-height: 1.55;
    text-align: left;
    overflow-wrap: anywhere;
  }

  .actions {
    margin-top: 24px;
    display: flex;
    justify-content: center;
    flex-wrap: wrap;
    gap: 10px;
  }

  a {
    color: inherit;
  }

  .button {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-height: 40px;
    padding: 0 15px;
    border-radius: 12px;
    background: var(--text);
    color: var(--bg);
    text-decoration: none;
    font-size: 14px;
    font-weight: 650;
  }

  .button-secondary {
    background: transparent;
    color: var(--link);
    border: 1px solid var(--card-border);
  }

  .footer {
    margin-top: 22px;
    padding-top: 20px;
    border-top: 1px solid var(--card-border);
    color: var(--muted);
    font-size: 13px;
  }

  .footer code {
    color: var(--muted-strong);
    background: rgba(148, 163, 184, 0.14);
    padding: 2px 6px;
    border-radius: 7px;
    font-family:
      ui-monospace,
      SFMono-Regular,
      Menlo,
      Monaco,
      Consolas,
      "Liberation Mono",
      monospace;
  }
</style>
</head>
<body>
  <main class="shell">
    <div class="brand">
      <span class="brand-mark"></span>
      <span>vygrant authorization broker</span>
    </div>

    <section class="card" aria-labelledby="title">
      <div class="icon" aria-hidden="true">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.25" stroke-linecap="round" stroke-linejoin="round">
          <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"/>
          <path d="M12 9v4"/>
          <path d="M12 17h.01"/>
        </svg>
      </div>

      <div class="status">Authentication failed</div>

      <h1 id="title">The sign-in could not be completed.</h1>

      <p>
        vygrant received a response from the browser, but the authorization flow ended with an error.
      </p>

      <div class="error-box">%s</div>

      <div class="actions">
        <a class="button" href="javascript:location.reload()">Try again</a>
        <a class="button button-secondary" href="https://github.com/vybraan/vygrant/issues/new/choose">Open GitHub Issue</a>
      </div>

      <p class="footer">
        You can close this tab and run the <code>vygrant</code> login command again from your terminal.
      </p>
    </section>
  </main>
</body>
</html>
`

func StartAuthFlow(w http.ResponseWriter, r *http.Request) {
	accountName := r.URL.Query().Get("account")
	acct, ok := LoadedAccounts[accountName]
	if !ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)

		safeError := html.EscapeString("Account '" + accountName + "' not found.")

		fmt.Fprintf(w, errorHTML, safeError)
		// fmt.Fprintf(w, errorHTML, "Account '"+accountName+"' not found.")
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

// writeErrorPage writes an HTML error response using the provided HTTP status and message.
// It sets the Content-Type to "text/html; charset=utf-8" and writes the body by formatting the package's errorHTML template with the message.
func writeErrorPage(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, errorHTML, message)
}

// HandleOAuthCallback returns an http.HandlerFunc that handles OAuth2 provider callbacks, exchanges the authorization code for a token, and stores that token keyed by account name.
//
// The handler validates the callback `state` to extract an account name, verifies the account is configured, and exchanges the `code` for an OAuth2 token. If `httpClient` is non-nil it is used for the token exchange. On success the token is saved into `tokenStore` under the account name and an HTML success page is written; on failure an error page with an appropriate HTTP status is returned.
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

		if err := tokenStore.Set(accountName, token); err != nil {
			log.Printf("failed to save token for account %s: %v", accountName, err)
			writeErrorPage(w, http.StatusInternalServerError, "Authentication succeeded but failed to save token.")
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		safeAccount := html.EscapeString(accountName)

		fmt.Fprintf(w, successHTML, safeAccount)
	}
}
