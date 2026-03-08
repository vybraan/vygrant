import QtQuick
import QtQuick.Layouts

Rectangle {
  id: root

  required property var theme
  required property int activePane

  signal paneSelected(int index)
  signal reloadRequested()

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
          onClicked: root.paneSelected(modelData.index)
        }
      }
    }

    Item { width: 1; height: 1 }

    ThemedButton {
      width: parent.width
      theme: root.theme
      variant: "text"
      text: "Reload Data"
      onClicked: root.reloadRequested()
    }
  }
}
