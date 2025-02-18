import os
import sys
import subprocess

def run_search_ai():
    package_dir = os.path.dirname(os.path.abspath(__file__))
    binary_path = os.path.join(package_dir, 'bin', 'search-ai')
    if not os.path.exists(binary_path):
        sys.exit(f"Error: Binary not found at {binary_path}")
    subprocess.run([binary_path] + sys.argv[1:])
