# filename: check_plot.py
import os

# Check if the file 'plot.png' exists
file_name = 'plot.png'
if os.path.isfile(file_name):
    print(f"{file_name} exists.")
else:
    print(f"{file_name} does not exist.")