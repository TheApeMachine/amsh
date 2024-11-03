import os
import re

def parse_goconvey_tests(directory):
    data_objects = []

    for root, _, files in os.walk(directory):
        for file in files:
            if file.endswith("_test.go"):
                test_file_path = os.path.join(root, file)
                implementation_file_path = os.path.join(root, file.replace("_test.go", ".go"))
                
                # Make sure the corresponding implementation file exists
                if not os.path.exists(implementation_file_path):
                    continue

                with open(test_file_path, "r") as f:
                    test_content = f.read()

                    # Extract all Test functions
                    test_functions = re.findall(r'func (Test\w+)\(.*?\)\s*{(.*?)}', test_content, re.DOTALL)
                    for test_function_name, test_body in test_functions:
                        
                        # Extract method calls inside the test body
                        method_calls = extract_method_calls(test_body)

                        # Extract the corresponding implementations from the implementation file
                        implementations = []
                        for method_name in method_calls:
                            implementation_code = extract_function_implementation(implementation_file_path, method_name)
                            if implementation_code:
                                implementations.append(implementation_code)

                        # Skip if no matching implementations were found
                        if not implementations:
                            continue

                        # Generate the prompt using the test function name
                        prompt = f"Write the following functions: {', '.join(method_calls)}, with robust error handling."

                        # Add to data objects
                        data_objects.append({
                            "messages": [
                                {"role": "system", "content": "You are a Go developer who writes robust error-handling functions."},
                                {"role": "user", "content": prompt},
                                {"role": "assistant", "content": "\n\n".join(implementations)}
                            ]
                        })

    return data_objects

def extract_method_calls(test_body):
    """Extract all method calls from the body of the test function."""
    # Regex to match method calls: any word followed by an open parenthesis
    method_calls = re.findall(r'(\w+)\s*\(', test_body)
    
    # List of Go built-in functions, keywords, and common test functions to ignore
    go_keywords = {
        "Convey", "So", "t", "fmt", "errors", "func", "defer", "go", "return", "if", "else",
        "for", "switch", "case", "select", "break", "continue", "fallthrough", "var", "const",
        "type", "struct", "interface", "map", "range", "chan"
    }

    # Remove Go keywords and built-in functions that are not user-defined
    filtered_calls = [call for call in method_calls if call not in go_keywords]

    return list(set(filtered_calls))  # Return unique method calls

def extract_function_implementation(file_path, function_name):
    """Extract the function implementation from a given file by function name."""
    with open(file_path, "r") as f:
        content = f.read()

    # Regex to match the function and its body
    pattern = rf'func {function_name}\s*\(.*?\)\s*{{.*?^}}'
    match = re.search(pattern, content, re.DOTALL | re.MULTILINE)
    if match:
        return match.group()
    return None

# Directory to be scanned
directory_path = "./"
data_objects = parse_goconvey_tests(directory_path)

# Write the generated data objects to a file in JSONL format
import json
with open("training_data.jsonl", "w") as output_file:
    for obj in data_objects:
        output_file.write(json.dumps(obj) + "\n")
