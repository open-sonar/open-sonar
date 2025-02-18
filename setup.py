from setuptools import setup, find_packages
import pathlib

here = pathlib.Path(__file__).parent.resolve()
long_description = (here / "README.md").read_text(encoding="utf-8")

setup(
    name="open-sonar",
    version="0.0.0",
    description="Still in development. Do not use till version 0.1.0.",
    long_description=long_description,
    long_description_content_type="text/markdown",
    author="Rohan Adwankar",
    author_email="rohan.adwankar@gmail.com",
    license="MIT",
    # url="https://github.com/RohanAdwankar/search-ai",
    packages=find_packages(),
    include_package_data=True,
    package_data={
        "search-ai": ["bin/search-ai"],
    },
    entry_points={
        "console_scripts": [
            "search-ai = search_ai:run_search_ai",
        ],
    },
)