---
name: miqt-ui-design
description: >-
  Design, build or audit a desktop interface with Qt 6 Widgets in Go through miqt (github.com/mappu/miqt/qt6): window and dialog layout, widget construction and signal handlers, QSS palette and type scale, focus order, accessible names, localisation, tray icon, and the Qt-thread rules. Use when adding or changing a window, form, dialog, card, menu or tray icon, or when styling, translating or auditing an existing one.
metadata:
  author: miqt-ui-design
  version: "1.0"
  miqt-version: "0.15"
  qt-version: "6.x"
  category: conceptual
---
# miqt UI design

miqt wraps Qt 6 for Go. This skill covers both halves of a miqt interface:
the design decisions (layout, type, colour, focus, accessibility,
localisation) and the binding mechanics that make them compile and stay on the
Qt thread. It targets **Qt Widgets**; miqt also binds QML, but a QML interface
follows `qt-ui-design`'s QML guidance instead.

## 1. Confirm the brief before designing

Run these seven items against the conversation and the code. Ask only what is
genuinely missing, and when the user cannot answer, choose a sensible Qt
default and name it in the response.

1. **Target platform** — Linux, Windows, macOS, or an embedded board.
2. **Screen shape** — rectangle, square, or round.
3. **Resolution and DPI** — approximate is enough; Qt 6 scales by itself.
4. **Design system** — an existing palette, token set or platform style
   (Fusion, Windows 11, macOS). Reuse it; do not invent a parallel vocabulary.
5. **Content priority** — primary, secondary, tertiary on this screen.
6. **Viewing distance** — handheld ~30 cm, desk ~60 cm, panel ~1.5 m.
7. **Locale and input** — language, RTL need, keyboard, pointer, touch.

For a small edit ("move the button", "change this label") skip the intake and
apply the rules silently.

## 2. The binding in one page

- **Generated, one file per class.** Every wrapper lives in
  `$(go env GOMODCACHE)/github.com/mappu/miqt@<version>/qt6/gen_q<name>.go`.
  That file is the authority on constructors, method names and overload
  numbers; read it instead of guessing.
- **Constructors are numbered by C++ overload**: `NewQLabel3(text)`,
  `NewQPushButton3("")`, `NewQScrollArea2()`, `NewQPropertyAnimation2(target,
  propertyName)`. The index is positional, not semantic: a newer binding can
  shift it, so never copy a number from an old project.
- **Wrappers embed their base class.** `QWidget` embeds `*QObject` and
  `*QPaintDevice`, so pass `w.QWidget` where a `*QWidget` is wanted and
  `w.QWidget.QObject` where only the object is. A `*QWidget` does not convert
  to a `*QObject` argument by itself.
- **Enums**: C++ scoped enums become `Type__Value` constants
  (`QFrame__NoFrame`, `QLineEdit__Password`, `QDialogButtonBox__Ok`); enums in
  the Qt namespace are plain (`AlignVCenter`, `NoFocus`, `RightToLeft`). Flags
  combine with `|`.
- **Signals** are `On<Signal>(func(args...))`: `OnClicked`, `OnTriggered`,
  `OnTimeout`, `OnCurrentIndexChanged(int)`, `OnFocusChanged(*QWidget,
  *QWidget)`. The slot runs on the Qt thread. A nil wrapper argument is normal.
- **Virtual methods** are overridable through the same mechanism, with a
  `super` callback:
  `w.OnPaintEvent(func(super func(*QPaintEvent), e *QPaintEvent) { super(e) })`.
  This is how a custom-drawn widget hooks `paintEvent`, `keyPressEvent`,
  `resizeEvent` or `showEvent`.
- **`SetObjectName` takes a `QAnyStringView`**, not a string:
  `w.SetObjectName(*qt6.NewQAnyStringView3("primary"))`. Wrap it in a local
  helper; it allocates on every call.
- **Byte arrays**: properties and binary blobs take `[]byte`
  (`NewQPropertyAnimation2(w.QWidget.QObject, []byte("geometry"))`).
- **Strings are Go strings**; `QString`, `QVariant` and `QByteArray` appear
  only where the C++ API truly needs them.
- **Ownership**: the C++ parent owns and deletes its children. Keep the Go
  wrapper in a struct field for as long as you will update that widget; keep a
  strong reference to every top-level window. Deleting or losing the wrapper
  does not delete the C++ object, and using a wrapper after its C++ parent is
  destroyed is a crash.
- **Layouts reparent.** Adding a widget to a layout of another parent moves it
  there; you do not need `SetParent` for that.
- **Build is cgo.** Qt 6 development headers, `pkg-config` and a working C++
  toolchain are required. Pin the binding version against the system Qt minor:
  a binding generated for an older Qt can fail to compile against a newer one.
  The first build compiles the C++ shim and takes minutes; keep the Go build
  cache in CI.
- **One `QApplication` per process.** `QApplication_Exec` blocks until
  quit; code after it runs at exit. Create the application before any widget
  and set quit-on-last-window-closed explicitly if the app lives in the tray.
- **UI calls stay on the Qt thread.** A goroutine touches widgets only through
  `github.com/mappu/miqt/qt6/mainthread`: `Start(fn)` to post, `Wait(fn)` to
  post and block. A worker that must not block uses `Start`; a caller that
  needs the result uses `Wait`. Reaching a widget directly from a goroutine is
  undefined behaviour, not a race you can tune away.

## 3. Window, dialog, layout

- **Pick the shell first**: `QMainWindow` for a document-style app,
  `QDialog` for a modal task with OK/Cancel, a plain `QWidget` for an
  auxiliary panel. A tray-only app still needs one window for settings.
- **Layout hierarchy over absolute geometry.** `QVBoxLayout` and
  `QHBoxLayout` for stacks and rows, `QGridLayout` for aligned columns,
  `QFormLayout` for label/field pairs (`AddRow3(labelText, field)` handles the
  caption column and mnemonic). Fixed `x`/`y` breaks under font scaling and
  translation.
- **Use a spacing rhythm**: 4, 8, 12, 16, 24. Page margins 16–24 px, group
  spacing 8–16, inner padding 12–18. Consistent rhythm reads as a system;
  arbitrary numbers read as noise.
- **Let layouts size the content**: `SetSizePolicy2(qt6.QSizePolicy__Expanding,
  qt6.QSizePolicy__Preferred)` for the field that should grow,
  `AddWidget2(w, 1)` for stretch, `AddStretch()` for slack. `SetFixedWidth` /
  `SetFixedHeight` on anything holding user or translated text truncates.
- **Survive a small window.** Put the scrolling body in a `QScrollArea` with
  `SetWidgetResizable(true)`, keep the primary action outside it, and set a
  minimum size derived from the content. A window that cannot shrink below
  its labels is a bug on 1366×768.
- **Group related controls** with a `QGroupBox` or a styled `QFrame`; one
  heading per group, not a heading per field.
- **Dialogs use `QDialogButtonBox`**, which orders buttons the way the
  platform expects (OK/Cancel inverted on macOS). Mark the affirmative button
  `SetDefault(true)` and close on Escape.
- **Qt 6 handles high DPI**: QSS `px` and layout integers are logical pixels.
  Never multiply by a device pixel ratio by hand.
- **Design to a minimum width** (240 px for a panel; more for a form), and let
  secondary content collapse, wrap or scroll below it.

## 4. Style sheet: one palette, semantic roles

`application.SetStyleSheet(sheet)` applies one sheet to the whole tree. A
widget-level sheet fragments the palette and is a maintenance trap; put every
colour, size and border in the application sheet and select by type,
`#objectName` and pseudo-state.

- **Name roles, not colours.** A token set of ~10 roles is enough: surface,
  raised surface, ink, muted ink, border, accent, accent-on, danger, warning,
  success. Keep the hex values in one Go constant or one builder, so a theme
  change is one edit.
- **QSS is a Qt dialect, not CSS**: no custom properties, no flexbox, no
  transitions, no `line-height`. Supported pseudo-states include `:hover`,
  `:focus`, `:disabled`, `:selected`, `:checked`, `:pressed`. Selector
  specificity and the cascade behave differently from the web.
- **The box model differs from CSS**: `min-width`/`min-height` size the
  contents rectangle; padding, border and margin sit around it, which is why a
  "22 px" button renders taller.
- **Popups are separate top-level widgets.** `QComboBox QAbstractItemView`,
  `QMenu`, tooltips and scrollbars have their own rules; a rule on the base
  control does not reach them.
- **Focus, hover, disabled and selected each need a visible style.** Focus is
  not optional: a focus ring of 2–3 px accent, or an equivalent, at every
  interactive element.
- **Contrast**: text ≥ 4.5:1, large text and UI components/borders ≥ 3:1, in
  both the resting and the focus state, against the real background.
- **Type scale**: pick a modular ratio (1.125 compact, 1.2 general, 1.25 open,
  1.333 bold hierarchy), a base size (16 px at desk distance), and map each
  role to one step. Three or four sizes per screen; body never below 16 px at
  desktop distance. Use the system font unless a brand requires otherwise.
- **Runtime restyling needs a repolish**:
  `w.Style().Unpolish(w.QWidget); w.Style().Polish(w.QWidget)` after changing
  the object name or a dynamic property.
- **Motion**: QSS has none. Use `QPropertyAnimation` for a meaningful change,
  100–200 ms, ease-out on enter and ease-in on exit; animate `geometry` or
  `windowOpacity`, not colours. Honour a reduced-motion preference from your
  settings and fall back to instant changes.
- **Dark mode** is a second sheet with the same role selectors, chosen before
  or at runtime, plus contrast checks for both.

## 5. Keyboard, focus, accessibility

- **Tab order is explicit**: `qt6.QWidget_SetTabOrder(a, b)` for each adjacent
  pair, in reading order. Creation order is not reading order, and a stray
  focusable container makes the chain jump.
- **Focus policy**: `qt6.StrongFocus` or `TabFocus` on real controls,
  `qt6.NoFocus` on scrolling containers, labels and decorative panels.
- **Mnemonics**: `&File` in a `QLabel` or button plus `SetBuddy(widget)` binds
  Alt+key to the field. Add them where a keyboard user repeats the same field.
- **Shortcuts**: `qt6.NewQShortcut2(qt6.NewQKeySequence2("Ctrl+S"),
  w.QWidget.QObject)` with `OnActivated`, or `QAction.SetShortcut`. Use the
  platform's standard keys through `QKeySequence__StandardKey` where one
  exists.
- **Conventions**: Enter performs the group's primary action, Escape cancels
  or closes, Tab never traps inside a panel. A modal dialog returns focus to
  the control that opened it.
- **Accessible names are mandatory** for fields and lists:
  `SetAccessibleName` for what the control is, `SetAccessibleDescription` for
  its error or hint. Bind the caption to the field rather than leaving the
  screen reader to infer it. Update both on language change.
- **Errors are visible and described**: the message below the field, the same
  text as the accessible description, never only a red border.
- **No colour-only state**: status, validation and progress pair colour with
  text, an icon or a shape.
- **Hit targets**: at least 24×24 px (WCAG 2.2 SC 2.5.8); 44×44 px for touch.
  QSS padding counts toward the target.
- **Hover is an enhancement.** Facts revealed on hover also need a visible
  path: a label, a hint line, or a keyboard-reachable tooltip.
- **Release check**: keyboard-only tour, screen reader pass on the form,
  focus always visible, no focus stop on a decorative container.

## 6. Localisation and RTL

- **Choose one string source and own its refresh**: a Go catalogue applied by
  an `ApplyTexts`-style method, or Qt's `QTranslator` +
  `QCoreApplication_InstallTranslator` with `QCoreApplication_Translate` at
  render time. Mixing both leaves half the window in the old language.
- **Re-apply texts on a language change**, including window title, menu
  actions, placeholders, accessible names and tooltips. Build the layout to
  re-measure, not to cache widths.
- **Locale-aware formatting** for dates, numbers and addresses: `QLocale`
  (`NewQLocale()` is the system locale, `NewQLocale2(name)` parses a locale
  name), never a hardcoded `DD/MM/YYYY`.
- **Plan for expansion**: translated strings run 30–40% longer than English.
  No fixed-width controls for user-visible text, `SetWordWrap(true)` on prose,
  and a scroll area for the rest.
- **RTL**: `QWidget.SetLayoutDirection(qt6.RightToLeft)` (or set it once on
  the application) mirrors layouts; directional icons still need mirroring and
  custom paint code must read the direction. Test an Arabic or Hebrew locale,
  not just a longer German string.
- **No text in images.** Every user-visible string belongs in the catalogue.
- **Icons**: prefer universal glyphs; avoid gestures or symbols whose meaning
  changes across regions.

## 7. Async work, feedback, state

- **Keep the Qt thread free.** Any operation over ~100 ms goes to a goroutine;
  the completion comes back through `mainthread.Start` or a queued signal.
  Blocking the loop freezes every window and the tray.
- **State has one owner.** Render widgets from the model, never read them back
  as the source of truth. A guard flag (`updating`) prevents a programmatic
  `SetText`/`SetCurrentIndex` from re-entering its own signal handler.
- **Feedback within 400 ms** (Doherty threshold). Longer work shows a progress
  indicator; work over ~10 s shows progress and a way to cancel.
- **Disable the action that is running** and re-enable it in the completion
  path, so a double click cannot start the work twice.
- **Errors reach the user in place**: next to the field for validation, in a
  status line for a failed operation. A dialog only for an error that stops
  the task.
- **Timers**: `QTimer` with `OnTimeout` and `Start(msec)`; stop the timer when
  its owner hides or closes.
- **Closing semantics**: decide the app's exit model up front (last window,
  tray, explicit Quit) and make it consistent across every window and the
  menu.

## 8. Tray and system integration

- **Check availability**: `qt6.QSystemTrayIcon_IsSystemTrayAvailable()` before
  relying on the tray; on Linux the icon is a StatusNotifierItem and can be
  absent on some desktops.
- **Icon sizes**: 22–32 px logical, one PNG per state or an SVG; the desktop
  scales whichever you give it.
- **`QSystemTrayIcon` + `QMenu`**: `SetContextMenu`, `SetToolTip`,
  `OnActivated` for `Trigger`/`DoubleClick`. The tooltip repeats the state in
  words; the icon is the fast read, not the only one.
- **Keep an escape hatch**: if the tray is the only path to Quit or Settings,
  a desktop without a tray leaves the app unmanageable. Duplicate the
  important actions in a window.
- **The menu stays short**: a handful of actions, separators between groups,
  verbs as labels, translated like every other string.

## 9. Verifying a miqt interface

1. `gofmt -l .` clean, `go vet ./...`, `go test ./...` green.
2. Build with cgo against the pinned Qt; a clean build on the CI image catches
   the header/version drift that a warm local cache hides.
3. **Guard the design tokens with tests.** A Go test can parse the style
   sheet string for font sizes and colour pairs, assert the body minimum and
   the 4.5:1 / 3:1 ratios, and fail when a new role skips the check. This is
   the cheapest way to keep a palette honest.
4. **Headless smoke test**: `QT_QPA_PLATFORM=offscreen go test ./...`
   constructs the window without a display, which catches constructor and
   wiring errors in CI.
5. **Screenshot pass**: `QWidget.Grab()` renders the window to a `QPixmap`;
   save it and review the layout at the target size.
6. **Manual pass on the real desktop**: 640–1366 px widths, 125% and 150%
   scaling, keyboard-only tour, language switch, dark mode if present, tray
   states, an error in the right place.
7. Review the result against the project's design direction; the checks above
   prove correctness, not taste.

## 10. Pitfalls

- **Constructor numbers** are overload indexes. When an example does not
  compile, the overload changed: open the generated file.
- **`SetObjectName` needs a `QAnyStringView`**; passing a `string` does not
  compile, and the conversion allocates.
- **Nil wrappers** arrive from signals (`previous` on the first focus change,
  for example) and from lookups; check before use.
- **`mainthread.Wait` blocks the caller**: never call it from a slot, and
  never while holding a lock the worker needs.
- **`QApplication_Exec` blocks**; anything after it is shutdown code.
- **QSS is not CSS**: no variables, flex, transitions, `line-height`, or
  `calc`. Popups need their own rules.
- **A hidden widget still owns its children**; hiding is not deleting, and a
  deleted C++ object leaves a dangling Go wrapper.
- **Enums are ints**: the zero value is a real policy (`NoFocus`,
  `QFrame__NoFrame`), so pass the constant you mean.
- **Qt beyond the binding's version** fails to compile rather than warn; pin
  both together.
- **The first build is slow** because it compiles the C++ shim; do not
  interpret a several-minute build as a hang.

## 11. References

- miqt: https://github.com/mappu/miqt
- Qt Widgets: https://doc.qt.io/qt-6/qtwidgets-index.html
- Qt layouts: https://doc.qt.io/qt-6/layout.html
- Qt style sheets: https://doc.qt.io/qt-6/stylesheet-reference.html
- Qt accessibility: https://doc.qt.io/qt-6/accessible.html
- Qt internationalisation: https://doc.qt.io/qt-6/internationalization.html
- Modular scale: https://www.modularscale.com/
- WCAG 2.2: https://www.w3.org/TR/WCAG22/
- Laws of UX: https://lawsofux.com/
