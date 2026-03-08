import QtQuick
import QtQuick.Layouts

ColumnLayout {
  id: root

  required property var theme
  required property var svc
  required property string selectedStatus
  required property int validCount
  required property int missingCount

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
        Text { text: String(validCount); color: theme.fgSurface; font.family: theme.fontSans; font.pixelSize: 24; font.bold: true }
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
        Text { text: String(missingCount); color: theme.fgSurface; font.family: theme.fontSans; font.pixelSize: 24; font.bold: true }
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
          text: "Selected: " + (svc.selectedAccount === "" ? "<none>" : svc.selectedAccount) + " | " + selectedStatus
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

        ThemedButton { width: parent.width; theme: root.theme; variant: "filled"; text: "Refresh Selected Token"; onClicked: svc.runAction("token refresh") }
        ThemedButton { width: parent.width; theme: root.theme; variant: "success"; text: "Fetch Token"; onClicked: svc.fetchToken() }
        ThemedButton { width: parent.width; theme: root.theme; variant: "tonal"; text: "Copy Token"; enabled: svc.tokenValue !== ""; onClicked: svc.copyToken() }
        ThemedButton { width: parent.width; theme: root.theme; variant: "error"; text: "Delete Token"; onClicked: svc.runAction("token delete") }

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
