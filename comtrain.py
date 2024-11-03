import os
import re

def parse_comments_and_functions(directory):
    data_objects = []

    for root, _, files in os.walk(directory):
        for file in files:
            if file.endswith(".go") and not file.endswith("_test.go"):
                file_path = os.path.join(root, file)
                
                with open(file_path, "r") as f:
                    content = f.read()

                    # Regex to match multi-line comments followed by functions or types
                    pattern_multi = r'(?P<comment>/\*.*?\*/)\s*(?P<code>type\s+\w+\s+struct\s*{.*?^}|func\s+\w+\s*\(.*?\)\s*{.*?^})'
                    matches_multi = re.finditer(pattern_multi, content, re.DOTALL | re.MULTILINE)

                    # Process multi-line comments with the corresponding function/type
                    for match in matches_multi:
                        comment = match.group('comment').strip()
                        code_block = match.group('code').strip()

                        prompt = clean_comment_for_prompt(comment)
                        response = code_block

                        data_objects.append({
                            "messages": [
                                {"role": "system", "content": "You are a Go developer who writes clear, readable, and idiomatic code."},
                                {"role": "user", "content": prompt},
                                {"role": "assistant", "content": response}
                            ]
                        })

                    # Regex to match single-line comments inside functions or methods
                    pattern_single = r'(?P<comment>//.*?$)(?P<code>.*?(\{.*?\}|^.*$))'
                    matches_single = re.finditer(pattern_single, content, re.MULTILINE)

                    # Process single-line comments with the corresponding line or block of code
                    for match in matches_single:
                        comment = match.group('comment').strip()
                        code_line = match.group('code').strip()

                        prompt = clean_comment_for_prompt(comment)
                        response = code_line

                        data_objects.append({
                            "messages": [
                                {"role": "system", "content": "You are a Go developer who writes clear, readable, and idiomatic code."},
                                {"role": "user", "content": prompt},
                                {"role": "assistant", "content": response}
                            ]
                        })

    return data_objects

def clean_comment_for_prompt(comment):
    """Clean the comment to make it suitable for use as a natural language prompt."""
    # Remove any stars, slashes, or other formatting characters
    comment = re.sub(r'//+', '', comment)
    comment = re.sub(r'/\*|\*/', '', comment)
    comment = comment.strip()
    return comment

# Directory to be scanned
directory_path = "./"
data_objects = parse_comments_and_functions(directory_path)

# Write the generated data objects to a file in JSONL format
import json
with open("comment_based_training_data.jsonl", "w") as output_file:
    for obj in data_objects:
        output_file.write(json.dumps(obj) + "\n")
