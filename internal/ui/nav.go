package ui

// GoToLinearMsg asks the app to switch to the Linear view and select the issue
// with the given identifier (e.g. "SRE-3686"). A view emits it as a command
// result; the root model routes it to whichever view can select issues.
type GoToLinearMsg struct{ Identifier string }
