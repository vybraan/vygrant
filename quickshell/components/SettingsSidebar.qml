import QtQuick
import QtQuick.Layouts

Rectangle {
  id: root

  required property var theme
  required property int activePane

  signal paneSelected(int index)
  signal reloadRequested()

  Layout.fillHeight: true
  Layout.preferredWidth: 118
  radius: theme.radiusNormal
  color: theme.surfaceContainer
  border.color: theme.outlineVariant

  Column {
    anchors.fill: parent
    anchors.margins: 8
    spacing: theme.spacingSmall

    Rectangle {
      width: parent.width
      height: 50
      radius: theme.radiusNormal
      color: theme.surface
      border.color: theme.outlineVariant

      Text {
        anchors.centerIn: parent
        text: "Vygrant"
        font.family: theme.fontSans
        font.pixelSize: 15
        font.bold: true
        color: theme.fgSurface
      }
    }

    Item { width: 1; height: 6 }

    Repeater {
      model: [
        { label: "Overview", icon: "home", index: 0 },
        { label: "Tokens", icon: "key", index: 1 }
      ]

      delegate: Item {
        id: navItem
        required property var modelData
        width: parent.width
        height: 74

        Rectangle {
          id: iconWrap
          width: 42
          height: 42
          radius: 21
          anchors.horizontalCenter: parent.horizontalCenter
          anchors.top: parent.top
          color: activePane === modelData.index ? theme.secondaryContainer : theme.surface
          border.color: activePane === modelData.index ? theme.primary : theme.outlineVariant

          Rectangle {
            anchors.fill: parent
            radius: parent.radius
            color: theme.primary
            opacity: navMouse.containsMouse ? (activePane === modelData.index ? 0.14 : 0.08) : 0
            Behavior on opacity { NumberAnimation { duration: 120 } }
          }

          Text {
            anchors.centerIn: parent
            text: modelData.icon
            font.family: "Material Symbols Rounded"
            font.pixelSize: 22
            color: activePane === modelData.index ? theme.fgSecondaryContainer : theme.fgSurfaceVariant
          }
        }

        Text {
          anchors.top: iconWrap.bottom
          anchors.topMargin: 6
          anchors.horizontalCenter: parent.horizontalCenter
          text: modelData.label
          font.family: theme.fontSans
          font.pixelSize: theme.fontSmall
          color: activePane === modelData.index ? theme.fgSurface : theme.fgSurfaceVariant
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

    Rectangle {
      width: 42
      height: 42
      radius: 21
      anchors.horizontalCenter: parent.horizontalCenter
      color: theme.surface
      border.color: theme.outlineVariant

      Rectangle {
        anchors.fill: parent
        radius: parent.radius
        color: theme.primary
        opacity: reloadMouse.containsMouse ? 0.10 : 0
        Behavior on opacity { NumberAnimation { duration: 120 } }
      }

      Text {
        anchors.centerIn: parent
        text: "refresh"
        font.family: "Material Symbols Rounded"
        font.pixelSize: 22
        color: theme.fgSurfaceVariant
      }

      MouseArea {
        id: reloadMouse
        anchors.fill: parent
        hoverEnabled: true
        cursorShape: Qt.PointingHandCursor
        onClicked: root.reloadRequested()
      }
    }

    Text {
      anchors.horizontalCenter: parent.horizontalCenter
      text: "Reload"
      font.family: theme.fontSans
      font.pixelSize: 10
      color: theme.fgSurfaceVariant
    }
  }
}
