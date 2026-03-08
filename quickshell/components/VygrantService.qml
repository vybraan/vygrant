import Quickshell
import Quickshell.Io
import QtQuick

Scope {
  id: svc

  property string vygrantCmd: "vygrant"
  property string selectedAccount: ""
  property string infoText: "Loading daemon info..."
  property string actionText: "Idle"
  property string tokenValue: ""
  property bool tokenVisible: false
  property var accounts: []
  property var statuses: []
  property string socketPath: ""
  property string configPath: ""
  property string tokenStorage: ""
  property string legacyMigration: ""
  property string httpPort: ""
  property string httpsPort: ""
  property string httpsPublicKey: ""
  property bool infoLoaded: false
  property string lastRefresh: ""

  function shellQuote(value) {
    return "'" + value.replace(/'/g, "'\"'\"'") + "'"
  }

  function parseLines(raw) {
    let out = []
    let rows = raw.split("\n")
    for (let i = 0; i < rows.length; i++) {
      let line = rows[i].trim()
      if (line.length > 0) {
        out.push(line)
      }
    }
    return out
  }

  function parseInfo(raw) {
    infoText = raw.trim()
    socketPath = ""
    configPath = ""
    tokenStorage = ""
    legacyMigration = ""
    httpPort = ""
    httpsPort = ""
    httpsPublicKey = ""

    let lines = parseLines(raw)
    for (let i = 0; i < lines.length; i++) {
      let line = lines[i]
      if (line.indexOf("Socket path:") === 0) socketPath = line.substring("Socket path:".length).trim()
      else if (line.indexOf("Config file:") === 0) configPath = line.substring("Config file:".length).trim()
      else if (line.indexOf("Token storage:") === 0) tokenStorage = line.substring("Token storage:".length).trim()
      else if (line.indexOf("Legacy migration:") === 0) legacyMigration = line.substring("Legacy migration:".length).trim()
      else if (line.indexOf("HTTP Port:") === 0) httpPort = line.substring("HTTP Port:".length).trim()
      else if (line.indexOf("HTTPS Port:") === 0) httpsPort = line.substring("HTTPS Port:".length).trim()
      else if (line.indexOf("HTTPS public key:") === 0) httpsPublicKey = line.substring("HTTPS public key:".length).trim()
    }
    infoLoaded = lines.length > 0
    lastRefresh = Qt.formatTime(new Date(), "hh:mm:ss")
  }

  function refreshAll() {
    infoProcess.command = ["sh", "-lc", vygrantCmd + " info"]
    infoProcess.running = true
    accountsProcess.command = ["sh", "-lc", vygrantCmd + " accounts"]
    accountsProcess.running = true
    statusProcess.command = ["sh", "-lc", vygrantCmd + " status"]
    statusProcess.running = true
  }

  function runAction(cmd) {
    if (selectedAccount === "") {
      actionText = "Select an account first."
      return
    }
    actionProcess.command = ["sh", "-lc", vygrantCmd + " " + cmd + " " + shellQuote(selectedAccount)]
    actionProcess.running = true
  }

  function fetchToken() {
    if (selectedAccount === "") {
      actionText = "Select an account first."
      return
    }
    tokenProcess.command = ["sh", "-lc", vygrantCmd + " token get " + shellQuote(selectedAccount)]
    tokenProcess.running = true
  }

  function copyToken() {
    if (tokenValue === "") {
      actionText = "No token fetched yet."
      return
    }
    copyProcess.command = ["sh", "-lc", "if command -v wl-copy >/dev/null 2>&1; then printf %s " + shellQuote(tokenValue) + " | wl-copy; elif command -v xclip >/dev/null 2>&1; then printf %s " + shellQuote(tokenValue) + " | xclip -selection clipboard; else exit 13; fi"]
    copyProcess.running = true
  }

  function handleAccountsOutput(raw) {
    accounts = parseLines(raw)
    if (selectedAccount === "" && accounts.length > 0) {
      selectedAccount = accounts[0]
    } else if (selectedAccount !== "" && accounts.indexOf(selectedAccount) === -1) {
      selectedAccount = accounts.length > 0 ? accounts[0] : ""
    }
  }

  Timer {
    interval: 15000
    running: true
    repeat: true
    onTriggered: svc.refreshAll()
  }

  Process {
    id: infoProcess
    running: false
    command: []

    stdout: StdioCollector {
      onStreamFinished: svc.parseInfo(this.text)
    }
    stderr: StdioCollector {
      onStreamFinished: if (this.text.trim().length > 0) svc.actionText = "Info error: " + this.text.trim()
    }
  }

  Process {
    id: accountsProcess
    running: false
    command: []

    stdout: StdioCollector {
      onStreamFinished: svc.handleAccountsOutput(this.text)
    }
    stderr: StdioCollector {
      onStreamFinished: if (this.text.trim().length > 0) svc.actionText = "Accounts error: " + this.text.trim()
    }
  }

  Process {
    id: statusProcess
    running: false
    command: []

    stdout: StdioCollector {
      onStreamFinished: svc.statuses = svc.parseLines(this.text)
    }
    stderr: StdioCollector {
      onStreamFinished: if (this.text.trim().length > 0) svc.actionText = "Status error: " + this.text.trim()
    }
  }

  Process {
    id: actionProcess
    running: false
    command: []

    stdout: StdioCollector {
      onStreamFinished: {
        let out = this.text.trim()
        svc.actionText = out.length > 0 ? out : "Command finished."
        svc.tokenValue = ""
        svc.tokenVisible = false
        svc.refreshAll()
      }
    }
    stderr: StdioCollector {
      onStreamFinished: if (this.text.trim().length > 0) svc.actionText = "Error: " + this.text.trim()
    }
  }

  Process {
    id: tokenProcess
    running: false
    command: []

    stdout: StdioCollector {
      onStreamFinished: {
        svc.tokenValue = this.text.trim()
        svc.tokenVisible = false
        svc.actionText = svc.tokenValue.length > 0 ? "Token fetched (masked)." : "No token returned."
      }
    }
    stderr: StdioCollector {
      onStreamFinished: if (this.text.trim().length > 0) svc.actionText = "Get token error: " + this.text.trim()
    }
  }

  Process {
    id: copyProcess
    running: false
    command: []

    stdout: StdioCollector {
      onStreamFinished: svc.actionText = "Token copied to clipboard."
    }
    stderr: StdioCollector {
      onStreamFinished: svc.actionText = "Clipboard tool missing (install wl-copy or xclip)."
    }
  }
}
