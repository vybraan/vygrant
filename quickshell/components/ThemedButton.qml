import QtQuick

Rectangle {
  id: root

  property string text: ""
  property bool enabled: true
  property string variant: "tonal" // filled|tonal|error|success|text
  property Theme theme

  signal clicked

  implicitWidth: Math.max(80, label.implicitWidth + 24)
  implicitHeight: 34
  radius: theme ? theme.radiusFull : 1000

  color: {
    if (!enabled) return theme.surfaceContainer
    if (variant === "filled") return theme.primary
    if (variant === "error") return theme.errorContainer
    if (variant === "success") return theme.successContainer
    if (variant === "text") return "transparent"
    return theme.secondaryContainer
  }

  border.width: 1
  border.color: {
    if (!enabled) return theme.outlineVariant
    if (variant === "filled") return Qt.darker(theme.primary, 1.2)
    if (variant === "error") return Qt.darker(theme.errorContainer, 1.2)
    if (variant === "success") return Qt.darker(theme.successContainer, 1.2)
    if (variant === "text") return "transparent"
    return theme.outline
  }

  Behavior on color {
    ColorAnimation { duration: 120 }
  }

  Text {
    id: label
    anchors.centerIn: parent
    text: root.text
    font.family: theme ? theme.fontSans : ""
    font.pixelSize: theme ? theme.fontNormal : 13
    color: {
      if (!enabled) return theme.fgSurfaceVariant
      if (variant === "filled") return theme.fgPrimary
      if (variant === "error") return theme.fgErrorContainer
      if (variant === "success") return theme.fgSuccessContainer
      return theme.fgSecondaryContainer
    }
  }

  Rectangle {
    anchors.fill: parent
    radius: parent.radius
    color: "transparent"
    border.width: 0
    opacity: mouse.pressed ? 0.14 : (mouse.containsMouse ? 0.08 : 0)
    Behavior on opacity {
      NumberAnimation { duration: 100 }
    }
  }

  MouseArea {
    id: mouse
    anchors.fill: parent
    enabled: root.enabled
    hoverEnabled: true
    cursorShape: root.enabled ? Qt.PointingHandCursor : Qt.ArrowCursor
    onClicked: root.clicked()
  }
}
