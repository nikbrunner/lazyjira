Output docs/assets/recordings/preview.gif

@start
@wait 1200

# Maximize the Issues panel and scroll through the long list
Type "+"
Sleep 800ms
@down 14
@wait 400
Type "G"
Sleep 800ms
Type "g"
Sleep 600ms

# Maximized Issue details, then step back out
@open
@wait 800
Ctrl+d
Sleep 800ms
@close
@wait 400
@close
@wait 600

# Issue tabs
@switch_tab
@wait 600
@switch_tab
@wait 600
Shift+Tab
Sleep 200ms
Shift+Tab
Sleep 600ms

# Filter picker: In Progress and In Review
Type "f"
Sleep 600ms
@select
@down
@select
@wait 600
@confirm
@wait 800
@close

# Mark a range of issues
Type "v"
Sleep 200ms
@down 2
Type "v"
Sleep 800ms
@close

# Issue details: description, then comments
Type "g"
Sleep 200ms
@open
@wait 600
Ctrl+d
Sleep 400ms
Ctrl+d
Sleep 600ms
@tab_next
@wait 1000

# Issue info
@panel 3
@down 4
@wait 400

# JQL search opens a temporary tab
@panel 2
Type "s"
Sleep 400ms
Set TypingSpeed 50ms
Type "priority = Critical"
Set TypingSpeed 0ms
Sleep 400ms
@confirm
@wait 1000
Type "x"
Sleep 400ms

# Switch project
@panel 0
Enter
Sleep 400ms
Set TypingSpeed 80ms
Type "plat"
Set TypingSpeed 0ms
Sleep 400ms
@confirm
@panel 2
@wait 600
@down 2
@wait 1000

# Help
Type "?"
Sleep 1200ms
Escape
Sleep 400ms

@quit
