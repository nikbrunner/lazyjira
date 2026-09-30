Output docs/assets/recordings/create-issue.gif

@start

# select Platform Services project
@panel 0
Enter
Sleep 300ms
Type "plat"
@confirm
@panel 2

# switch to Assigned tab
@switch_tab

# create a new issue
@create

# the form opens on Bug, the first issue type
Sleep 600ms

# type the summary
Set TypingSpeed 40ms
Type "Login page crashes on expired token refresh"
Set TypingSpeed 0ms
Sleep 300ms

# write the description in $EDITOR
Tab
Sleep 300ms
Ctrl+G
Sleep 800ms
Type "i"
Set TypingSpeed 30ms
Type "Steps to reproduce:"
Enter
Type "1. Sign in and leave the tab open for an hour"
Enter
Type "2. Click any link"
Enter
Enter
Type "The page crashes instead of refreshing the token."
Set TypingSpeed 0ms
Sleep 600ms
Escape
Type ":wq"
Sleep 300ms
Enter
Sleep 800ms

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
