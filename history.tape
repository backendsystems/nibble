Output history.gif

Set FontFamily "DejaVu Sans Mono,Noto Color Emoji"
Set FontSize 16
Set Width 1000
Set Height 570
Set Shell "bash"
Set TypingSpeed 0ms

# Run a demo scan first so history is populated
Type "./nibble --demo"
Enter
Sleep 1s

# Open history view
Type "r"
Sleep 300ms

# Expand first interface node
Right
Sleep 300ms

# Expand first network node
Right
Sleep 300ms

# Open scan detail view
Right
Sleep 300ms

# Scroll through hosts
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms

# Scan all ports
Right
Sleep 2s

# Show delete
Left
Sleep 600ms
Delete
Sleep 2s
Delete
Sleep 1s

Left
Left
Left