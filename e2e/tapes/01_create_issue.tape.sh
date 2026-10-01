Output docs/assets/recordings/create-issue.gif

@clipboard_image e2e/fixtures/paste.png
@start
@wait 800

# select Platform Services project
@panel 0
Enter
Sleep 600ms
Set TypingSpeed 120ms
Type "plat"
Set TypingSpeed 0ms
Sleep 600ms
@confirm
@wait 800
@panel 2

# switch to Assigned tab
@switch_tab
@wait 800

# create a new issue
@create

# the form opens on Bug, the first issue type
Sleep 1200ms

# type the summary
Tab
Sleep 200ms
Set TypingSpeed 40ms
Type "Login page crashes on expired token refresh"
Set TypingSpeed 0ms
Sleep 300ms

# write the description in the form
Tab
Sleep 300ms
Set TypingSpeed 30ms
Type "Steps to reproduce:"
Enter
Type "1. Sign in and leave the tab open for an hour"
Enter
Type "2. Click any link"
Enter
Enter
Type "The page crashes instead of refreshing the token."
Enter
Enter
Set TypingSpeed 0ms

# paste the screenshot from the clipboard
Ctrl+V
Sleep 1500ms

# go to fields, past the Type row
Tab
Sleep 200ms
@down

# edit priority to High
@edit
@down
@confirm

# assignee: pick Demo User (ourselves so it shows in Assigned tab)
@down
@edit
@confirm

# labels: pick a few
@down
@edit
@toggle
@down
@toggle
@down
@toggle
@confirm

# components: pick API and Frontend
@down
@edit
@toggle
@down
@toggle
@confirm

# review: scroll through fields
@down
@wait 300

# submit
Ctrl+S
Sleep 600ms

# issue created: open it to show the description
@wait 800
@open
@wait 2000
@quit
