import QtQuick

QtObject {
  // Spacing and sizing rhythm inspired by Caelestia's appearance config.
  readonly property int spacingSmall: 7
  readonly property int spacingNormal: 12
  readonly property int spacingLarge: 20
  readonly property int padSmall: 5
  readonly property int padNormal: 10
  readonly property int padLarge: 15

  readonly property int radiusSmall: 12
  readonly property int radiusNormal: 17
  readonly property int radiusLarge: 25
  readonly property int radiusFull: 1000

  readonly property string fontSans: "Rubik"
  readonly property int fontSmall: 11
  readonly property int fontNormal: 13
  readonly property int fontLarge: 18

  // Material-like tonal palette for a Caelestia-adjacent look.
  readonly property color bg: "#0F1218"
  readonly property color surface: "#161B24"
  readonly property color surfaceContainer: "#1D2430"
  readonly property color surfaceContainerHigh: "#252E3D"
  readonly property color outline: "#344052"
  readonly property color outlineVariant: "#2B3545"
  readonly property color primary: "#A8C7FA"
  readonly property color fgPrimary: "#0E223F"
  readonly property color secondaryContainer: "#374A67"
  readonly property color fgSecondaryContainer: "#D9E3F7"
  readonly property color errorContainer: "#5F3A39"
  readonly property color fgErrorContainer: "#F9DEDC"
  readonly property color successContainer: "#2F4F40"
  readonly property color fgSuccessContainer: "#D6F3E3"
  readonly property color fgSurface: "#E3E8F2"
  readonly property color fgSurfaceVariant: "#B7C1D3"
}
