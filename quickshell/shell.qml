import Quickshell
import QtQuick
import QtQuick.Layouts
import "components"

PanelWindow {
  id: root

  anchors {
    top: true
    bottom: true
    left: true
    right: true
  }

  color: "transparent"
  property int activePane: 0

  Theme { id: theme }

  VygrantService {
    id: svc
    Component.onCompleted: refreshAll()
  }

  function selectedStatusLine() {
    if (svc.selectedAccount === "") return "No account selected"
    for (let i = 0; i < svc.statuses.length; i++) {
      let line = svc.statuses[i]
      if (line.indexOf(svc.selectedAccount + ":") === 0) return line
    }
    return "Status unavailable"
  }

  function countBySuffix(suffix) {
    let c = 0
    for (let i = 0; i < svc.statuses.length; i++) {
      if (svc.statuses[i].indexOf(suffix) >= 0) c++
    }
    return c
  }

  Rectangle {
    anchors.fill: parent
    color: Qt.rgba(0, 0, 0, 0.40)

    Item {
      anchors.fill: parent
      focus: true
      Keys.onEscapePressed: Qt.quit()
    }

    MouseArea {
      anchors.fill: parent
      onClicked: Qt.quit()
    }

    Rectangle {
      id: shellCard
      width: Math.min(root.width - 120, 980)
      height: Math.min(root.height - 120, 640)
      anchors.centerIn: parent
      radius: theme.radiusLarge
      color: theme.surface
      border.color: theme.outline

      MouseArea {
        anchors.fill: parent
        onClicked: function(mouse) { mouse.accepted = true }
      }

      RowLayout {
        anchors.fill: parent
        anchors.margins: theme.padNormal
        spacing: theme.spacingNormal

        Rectangle {
          Layout.fillHeight: true
          Layout.preferredWidth: 224
          radius: theme.radiusNormal
          color: theme.surfaceContainer
          border.color: theme.outlineVariant

          Column {
            anchors.fill: parent
            anchors.margins: theme.padNormal
            spacing: theme.spacingSmall

            Rectangle {
              width: parent.width
              height: 66
              radius: theme.radiusNormal
              color: theme.surface
              border.color: theme.outlineVariant

              Column {
                anchors.fill: parent
                anchors.margins: 10
                spacing: 2
                Text {
                  text: "vygrant"
                  font.family: theme.fontSans
                  font.pixelSize: theme.fontLarge
                  font.bold: true
                  color: theme.primary
                }
                Text {
                  text: "settings"
                  font.family: theme.fontSans
                  font.pixelSize: theme.fontSmall
                  color: theme.fgSurfaceVariant
                }
              }
            }

            Item { width: 1; height: theme.spacingSmall }

            Text {
              text: "General"
              font.family: theme.fontSans
              font.pixelSize: theme.fontSmall
              color: theme.fgSurfaceVariant
              leftPadding: 6
            }

            Repeater {
              model: [
                { label: "Overview", icon: "home", index: 0 },
                { label: "Tokens", icon: "key", index: 1 }
              ]

              delegate: Rectangle {
                id: navItem
                required property var modelData
                width: parent.width
                height: 44
                radius: theme.radiusFull
                color: activePane === modelData.index ? theme.secondaryContainer : theme.surfaceContainer
                border.color: activePane === modelData.index ? theme.primary : theme.outlineVariant

                Rectangle {
                  anchors.fill: parent
                  radius: parent.radius
                  color: theme.primary
                  opacity: navMouse.containsMouse ? (activePane === modelData.index ? 0.12 : 0.08) : 0
                  Behavior on opacity { NumberAnimation { duration: 120 } }
                }

                Rectangle {
                  visible: activePane === modelData.index
                  width: 4
                  height: parent.height - 14
                  radius: 2
                  anchors.left: parent.left
                  anchors.leftMargin: 8
                  anchors.verticalCenter: parent.verticalCenter
                  color: theme.primary
                }

                Row {
                  anchors.centerIn: parent
                  spacing: 10
                  Text {
                    text: modelData.icon
                    font.family: "Material Symbols Rounded"
                    font.pixelSize: theme.fontLarge
                    color: activePane === modelData.index ? theme.fgSecondaryContainer : theme.fgSurfaceVariant
                  }
                  Text {
                    text: modelData.label
                    font.family: theme.fontSans
                    font.pixelSize: theme.fontNormal
                    color: activePane === modelData.index ? theme.fgSecondaryContainer : theme.fgSurfaceVariant
                  }
                }

                MouseArea {
                  id: navMouse
                  anchors.fill: parent
                  hoverEnabled: true
                  cursorShape: Qt.PointingHandCursor
                  onClicked: activePane = modelData.index
                }
              }
            }

            Item { width: 1; height: 1 }

            ThemedButton {
              width: parent.width
              theme: theme
              variant: "text"
              text: "Reload Data"
              onClicked: svc.refreshAll()
            }
          }
        }

        Rectangle {
          Layout.fillWidth: true
          Layout.fillHeight: true
          radius: theme.radiusNormal
          color: theme.surfaceContainer
          border.color: theme.outlineVariant

          StackLayout {
            anchors.fill: parent
            anchors.margins: theme.padNormal
            currentIndex: activePane

            // Overview
            ColumnLayout {
              spacing: theme.spacingNormal

              RowLayout {
                Layout.fillWidth: true
                Text {
                  text: "Overview"
                  font.family: theme.fontSans
                  font.pixelSize: theme.fontLarge
                  font.bold: true
                  color: theme.primary
                }
                Item { Layout.fillWidth: true }
                Rectangle {
                  radius: theme.radiusFull
                  color: svc.infoLoaded ? theme.successContainer : theme.errorContainer
                  border.color: theme.outlineVariant
                  implicitHeight: 28
                  implicitWidth: daemonBadge.implicitWidth + 18
                  Text {
                    id: daemonBadge
                    anchors.centerIn: parent
                    text: svc.infoLoaded ? "Daemon reachable" : "Daemon unavailable"
                    font.family: theme.fontSans
                    font.pixelSize: theme.fontSmall
                    color: svc.infoLoaded ? theme.fgSuccessContainer : theme.fgErrorContainer
                  }
                }
              }

              RowLayout {
                Layout.fillWidth: true
                spacing: theme.spacingNormal

                Rectangle {
                  Layout.fillWidth: true
                  Layout.preferredHeight: 82
                  radius: theme.radiusNormal
                  color: theme.surface
                  border.color: theme.outlineVariant
                  Column {
                    anchors.fill: parent
                    anchors.margins: theme.padNormal
                    spacing: 4
                    Text { text: "Accounts"; color: theme.fgSurfaceVariant; font.family: theme.fontSans; font.pixelSize: theme.fontSmall }
                    Text { text: String(svc.accounts.length); color: theme.fgSurface; font.family: theme.fontSans; font.pixelSize: 24; font.bold: true }
                  }
                }

                Rectangle {
                  Layout.fillWidth: true
                  Layout.preferredHeight: 82
                  radius: theme.radiusNormal
                  color: theme.surface
                  border.color: theme.outlineVariant
                  Column {
                    anchors.fill: parent
                    anchors.margins: theme.padNormal
                    spacing: 4
                    Text { text: "Valid"; color: theme.fgSurfaceVariant; font.family: theme.fontSans; font.pixelSize: theme.fontSmall }
                    Text { text: String(countBySuffix("token valid")); color: theme.fgSurface; font.family: theme.fontSans; font.pixelSize: 24; font.bold: true }
                  }
                }

                Rectangle {
                  Layout.fillWidth: true
                  Layout.preferredHeight: 82
                  radius: theme.radiusNormal
                  color: theme.surface
                  border.color: theme.outlineVariant
                  Column {
                    anchors.fill: parent
                    anchors.margins: theme.padNormal
                    spacing: 4
                    Text { text: "Missing / Expired"; color: theme.fgSurfaceVariant; font.family: theme.fontSans; font.pixelSize: theme.fontSmall }
                    Text { text: String(countBySuffix("missing or expired")); color: theme.fgSurface; font.family: theme.fontSans; font.pixelSize: 24; font.bold: true }
                  }
                }
              }

              RowLayout {
                Layout.fillWidth: true
                Layout.fillHeight: true
                spacing: theme.spacingNormal

                Rectangle {
                  Layout.fillWidth: true
                  Layout.fillHeight: true
                  radius: theme.radiusNormal
                  color: theme.surface
                  border.color: theme.outlineVariant

                  Column {
                    anchors.fill: parent
                    anchors.margins: theme.padNormal
                    spacing: theme.spacingSmall
                    Text { text: "Daemon Details"; font.family: theme.fontSans; font.pixelSize: theme.fontNormal; font.bold: true; color: theme.primary }

                    GridLayout {
                      width: parent.width
                      columns: 2
                      rowSpacing: 4
                      columnSpacing: 10

                      Text { text: "Socket"; color: theme.fgSurfaceVariant; font.family: theme.fontSans; font.pixelSize: theme.fontSmall }
                      Text { text: svc.socketPath === "" ? "<unknown>" : svc.socketPath; color: theme.fgSurface; font.family: theme.fontSans; font.pixelSize: theme.fontSmall; elide: Text.ElideMiddle; Layout.fillWidth: true }

                      Text { text: "Config"; color: theme.fgSurfaceVariant; font.family: theme.fontSans; font.pixelSize: theme.fontSmall }
                      Text { text: svc.configPath === "" ? "<unknown>" : svc.configPath; color: theme.fgSurface; font.family: theme.fontSans; font.pixelSize: theme.fontSmall; elide: Text.ElideMiddle; Layout.fillWidth: true }

                      Text { text: "Storage"; color: theme.fgSurfaceVariant; font.family: theme.fontSans; font.pixelSize: theme.fontSmall }
                      Text { text: svc.tokenStorage === "" ? "<unknown>" : svc.tokenStorage; color: theme.fgSurface; font.family: theme.fontSans; font.pixelSize: theme.fontSmall; Layout.fillWidth: true }

                      Text { text: "HTTP / HTTPS"; color: theme.fgSurfaceVariant; font.family: theme.fontSans; font.pixelSize: theme.fontSmall }
                      Text { text: (svc.httpPort === "" ? "disabled" : svc.httpPort) + " / " + (svc.httpsPort === "" ? "disabled" : svc.httpsPort); color: theme.fgSurface; font.family: theme.fontSans; font.pixelSize: theme.fontSmall; Layout.fillWidth: true }
                    }

                    Rectangle { width: parent.width; height: 1; color: theme.outlineVariant }
                    Text {
                      width: parent.width
                      text: "Selected: " + (svc.selectedAccount === "" ? "<none>" : svc.selectedAccount) + " | " + selectedStatusLine()
                      wrapMode: Text.WrapAnywhere
                      color: theme.fgSurfaceVariant
                      font.family: theme.fontSans
                      font.pixelSize: theme.fontSmall
                    }
                  }
                }

                Rectangle {
                  Layout.preferredWidth: 285
                  Layout.fillHeight: true
                  radius: theme.radiusNormal
                  color: theme.surface
                  border.color: theme.outlineVariant

                  Column {
                    anchors.fill: parent
                    anchors.margins: theme.padNormal
                    spacing: theme.spacingSmall
                    Text { text: "Quick Actions"; font.family: theme.fontSans; font.pixelSize: theme.fontNormal; font.bold: true; color: theme.primary }

                    Flickable {
                      width: parent.width
                      height: 36
                      contentWidth: chips.implicitWidth
                      clip: true
                      Row {
                        id: chips
                        spacing: 6
                        Repeater {
                          model: svc.accounts
                          delegate: Rectangle {
                            required property string modelData
                            height: 28
                            width: chipText.implicitWidth + 16
                            radius: theme.radiusFull
                            color: svc.selectedAccount === modelData ? theme.secondaryContainer : theme.surfaceContainer
                            border.color: svc.selectedAccount === modelData ? theme.primary : theme.outlineVariant
                            Text {
                              id: chipText
                              anchors.centerIn: parent
                              text: modelData
                              color: svc.selectedAccount === modelData ? theme.fgSecondaryContainer : theme.fgSurfaceVariant
                              font.family: theme.fontSans
                              font.pixelSize: theme.fontSmall
                            }
                            MouseArea {
                              anchors.fill: parent
                              hoverEnabled: true
                              cursorShape: Qt.PointingHandCursor
                              onClicked: svc.selectedAccount = modelData
                            }
                          }
                        }
                      }
                    }

                    ThemedButton { width: parent.width; theme: theme; variant: "filled"; text: "Refresh Selected Token"; onClicked: svc.runAction("token refresh") }
                    ThemedButton { width: parent.width; theme: theme; variant: "success"; text: "Fetch Token"; onClicked: svc.fetchToken() }
                    ThemedButton { width: parent.width; theme: theme; variant: "tonal"; text: "Copy Token"; enabled: svc.tokenValue !== ""; onClicked: svc.copyToken() }
                    ThemedButton { width: parent.width; theme: theme; variant: "error"; text: "Delete Token"; onClicked: svc.runAction("token delete") }

                    Rectangle { width: parent.width; height: 1; color: theme.outlineVariant }
                    Text {
                      width: parent.width
                      text: "Last activity: " + svc.actionText
                      wrapMode: Text.WrapAnywhere
                      font.family: theme.fontSans
                      font.pixelSize: theme.fontSmall
                      color: theme.fgSurfaceVariant
                    }
                  }
                }
              }
            }

            // Tokens
            RowLayout {
              spacing: theme.spacingNormal

              ColumnLayout {
                Layout.fillWidth: true
                Layout.fillHeight: true
                spacing: theme.spacingSmall

                Text { text: "Tokens"; font.family: theme.fontSans; font.pixelSize: theme.fontLarge; font.bold: true; color: theme.primary }
                Text { text: "Operate and inspect token output"; font.family: theme.fontSans; font.pixelSize: theme.fontSmall; color: theme.fgSurfaceVariant }

                Row {
                  spacing: theme.spacingSmall
                  ThemedButton { theme: theme; variant: "filled"; text: "Refresh"; onClicked: svc.runAction("token refresh") }
                  ThemedButton { theme: theme; variant: "error"; text: "Delete"; onClicked: svc.runAction("token delete") }
                  ThemedButton { theme: theme; variant: "success"; text: "Fetch"; onClicked: svc.fetchToken() }
                  ThemedButton { theme: theme; variant: "tonal"; text: svc.tokenVisible ? "Mask" : "Reveal"; enabled: svc.tokenValue !== ""; onClicked: svc.tokenVisible = !svc.tokenVisible }
                  ThemedButton { theme: theme; variant: "tonal"; text: "Copy"; enabled: svc.tokenValue !== ""; onClicked: svc.copyToken() }
                }

                Rectangle { Layout.fillWidth: true; height: 1; color: theme.outlineVariant }

                Text {
                  Layout.fillWidth: true
                  text: "Account: " + (svc.selectedAccount === "" ? "<none>" : svc.selectedAccount)
                  font.family: theme.fontSans
                  font.pixelSize: theme.fontSmall
                  color: theme.fgSurfaceVariant
                }

                Rectangle {
                  Layout.fillWidth: true
                  Layout.fillHeight: true
                  radius: theme.radiusNormal
                  color: theme.surface
                  border.color: theme.outlineVariant

                  Column {
                    anchors.fill: parent
                    anchors.margins: theme.padNormal
                    spacing: theme.spacingSmall
                    Text { text: "Token Output"; font.family: theme.fontSans; font.pixelSize: theme.fontNormal; font.bold: true; color: theme.primary }
                    Text {
                      width: parent.width
                      wrapMode: Text.WrapAnywhere
                      text: svc.tokenValue === "" ? "No token fetched." : (svc.tokenVisible ? svc.tokenValue : "*".repeat(Math.min(svc.tokenValue.length, 180)))
                      font.family: theme.fontSans
                      font.pixelSize: theme.fontSmall
                      color: theme.fgSurface
                    }
                    Rectangle { width: parent.width; height: 1; color: theme.outlineVariant }
                    Text {
                      width: parent.width
                      wrapMode: Text.WrapAnywhere
                      text: "Activity: " + svc.actionText
                      font.family: theme.fontSans
                      font.pixelSize: theme.fontSmall
                      color: theme.fgSurfaceVariant
                    }
                  }
                }
              }

              Rectangle {
                Layout.preferredWidth: 300
                Layout.fillHeight: true
                radius: theme.radiusNormal
                color: theme.surface
                border.color: theme.outlineVariant

                Column {
                  anchors.fill: parent
                  anchors.margins: theme.padNormal
                  spacing: theme.spacingSmall
                  Text { text: "Account Selection"; font.family: theme.fontSans; font.pixelSize: theme.fontNormal; font.bold: true; color: theme.primary }
                  Text { text: "Selected: " + (svc.selectedAccount === "" ? "<none>" : svc.selectedAccount); font.family: theme.fontSans; font.pixelSize: theme.fontSmall; color: theme.fgSurfaceVariant; elide: Text.ElideRight; width: parent.width }
                  Text { text: selectedStatusLine(); wrapMode: Text.WrapAnywhere; font.family: theme.fontSans; font.pixelSize: theme.fontSmall; color: theme.fgSurfaceVariant; width: parent.width }
                  Rectangle { width: parent.width; height: 1; color: theme.outlineVariant }

                  Flickable {
                    width: parent.width
                    height: parent.height - 88
                    contentHeight: tokenAccountColumn.implicitHeight
                    clip: true

                    Column {
                      id: tokenAccountColumn
                      width: parent.width
                      spacing: theme.spacingSmall

                      Repeater {
                        model: svc.accounts
                        delegate: Rectangle {
                          required property string modelData
                          width: tokenAccountColumn.width
                          height: 36
                          radius: theme.radiusNormal
                          color: svc.selectedAccount === modelData ? theme.secondaryContainer : theme.surfaceContainer
                          border.color: svc.selectedAccount === modelData ? theme.primary : theme.outlineVariant
                          Text {
                            anchors.centerIn: parent
                            text: modelData
                            font.family: theme.fontSans
                            font.pixelSize: theme.fontNormal
                            color: svc.selectedAccount === modelData ? theme.fgSecondaryContainer : theme.fgSurface
                          }
                          MouseArea {
                            anchors.fill: parent
                            hoverEnabled: true
                            cursorShape: Qt.PointingHandCursor
                            onClicked: svc.selectedAccount = modelData
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
