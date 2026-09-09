import re

with open("Note/prompt-qa.md", "r") as f:
    content = f.read()

# Fix the conflict in prompt-qa.md by keeping HEAD (which has the original, more complete rules)
content = re.sub(r'<<<<<<< HEAD\n(.*?)\n=======\n.*?\n>>>>>>> [^\n]*\n', r'\1\n', content, flags=re.DOTALL)

with open("Note/prompt-qa.md", "w") as f:
    f.write(content)
