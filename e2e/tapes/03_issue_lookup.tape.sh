Output docs/assets/recordings/issue-lookup.gif

@start
@wait 1000

# Open the lookup; it starts with the active project
Type "#"
Sleep 1200ms

# Switch to another project by its name
Ctrl+u
Sleep 400ms
Set TypingSpeed 150ms
Type "pl"
Set TypingSpeed 0ms
Sleep 1000ms
Tab
Sleep 1000ms

# Type the number and open the issue in the Issue Edit View
Set TypingSpeed 150ms
Type "3"
Set TypingSpeed 0ms
Sleep 1200ms
@confirm
@wait 2500

# esc returns to where you were
@close
@wait 1500

@quit
