import QtQuick
import QtQuick.Layouts

RowLayout {
  id: root

  required property var theme
  required property var svc
  required property string selectedStatus

  spacing: theme.spacingNormal

  ColumnLayout {
    Layout.fillWidth: true
    Layout.fillHeight: true
    spacing: theme.spacingSmall

    Text { text: "Tokens"; font.family: theme.fontSans; font.pixelSize: theme.fontLarge; font.bold: true; color: theme.primary }
    Text { text: "Operate and inspect token output"; font.family: theme.fontSans; font.pixelSize: theme.fontSmall; color: theme.fgSurfaceVariant }

    Row {
      spacing: theme.spacingSmall
      ThemedButton { theme: root.theme; variant: "filled"; text: "Refresh"; onClicked: svc.runAction("token refresh") }
      ThemedButton { theme: root.theme; variant: "error"; text: "Delete"; onClicked: svc.runAction("token delete") }
      ThemedButton { theme: root.theme; variant: "success"; text: "Fetch"; onClicked: svc.fetchToken() }
      ThemedButton { theme: root.theme; variant: "tonal"; text: svc.tokenVisible ? "Mask" : "Reveal"; enabled: svc.tokenValue !== ""; onClicked: svc.tokenVisible = !svc.tokenVisible }
      ThemedButton { theme: root.theme; variant: "tonal"; text: "Copy"; enabled: svc.tokenValue !== ""; onClicked: svc.copyToken() }
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
      Text { text: selectedStatus; wrapMode: Text.WrapAnywhere; font.family: theme.fontSans; font.pixelSize: theme.fontSmall; color: theme.fgSurfaceVariant; width: parent.width }
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
