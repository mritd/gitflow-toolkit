# GitFlow Toolkit Demo
# Usage: vhs demo.tape

Output demo.webm
Set FontSize 24
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set TypingSpeed 30ms
Set Padding 20
Set Framerate 30
Set CursorBlink true
Set Shell "bash"

# Prepare test repository (hidden)
Hide
Type "./setup_demo.sh && cd /tmp/testgit"
Enter
Sleep 1s
Show

# Show staged files
Type "git status"
Enter
Sleep 2s

# =============================================
# Part 1: Manual Commit
# =============================================
Type "git ci"
Enter
Sleep 1500ms

# Select 'feat' (first option, press Enter)
Enter
Sleep 800ms

# Input scope
Type "auth"
Tab
Sleep 300ms

# Input subject
Type "add login handler and token utilities"
Tab
Sleep 300ms

# Skip body and footer
Tab
Tab
Sleep 300ms

# Submit form
Enter
Sleep 1500ms

# Confirm commit
Enter
Sleep 2500ms

# =============================================
# Part 2: AI Auto Generate Commit
# =============================================

Sleep 2s

# Add more changes for AI demo
Hide
Type "echo '// Updated' >> auth/handler.go && echo '// Updated' >> auth/token.go && git add auth/"
Enter
Sleep 500ms
Show

Type "git status"
Enter
Sleep 2s

Type "git ci"
Enter
Sleep 1500ms

# Press Tab to switch to Auto Generate button
Tab
Sleep 1500ms

# Press Enter to trigger Auto Generate
Enter
Sleep 8s

# Confirm commit
Enter
Sleep 2s

# End with project URL
Type "# github.com/mritd/gitflow-toolkit"
Enter
Sleep 3s
