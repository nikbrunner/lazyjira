Output docs/assets/recordings/hero.gif

@start
@wait 800

# Browse issues; details follow the cursor
@down 3
@wait 300

# Switch issue tabs
@switch_tab
@wait 500
Shift+Tab
Sleep 500ms

# Maximize the list, scroll, open an issue, step back
Type "+"
Sleep 500ms
@down 6
@up 6
@open
@wait 1200
@close
@wait 300
@close
@wait 800

@quit
