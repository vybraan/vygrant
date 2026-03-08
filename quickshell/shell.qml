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

        SettingsSidebar {
          theme: theme
          activePane: root.activePane
          onPaneSelected: function(index) { root.activePane = index }
          onReloadRequested: svc.refreshAll()
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
            currentIndex: root.activePane

            OverviewPane {
              theme: theme
              svc: svc
              selectedStatus: root.selectedStatusLine()
              validCount: root.countBySuffix("token valid")
              missingCount: root.countBySuffix("missing or expired")
            }

            TokensPane {
              theme: theme
              svc: svc
              selectedStatus: root.selectedStatusLine()
            }
          }
        }
      }
    }
  }
}
