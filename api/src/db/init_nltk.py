#!/usr/bin/env python
import subprocess
import sys


def install_nltk():
    subprocess.call([sys.executable, "-m", "pip", "install", 'nltk'])


def download_corpora():
    subprocess.call([
        sys.executable,
        "-m",
        "nltk.downloader",
        "-d",
        './src/db/nltk_data',
        'gutenberg',
    ])


if __name__ == '__main__':
    install_nltk()
    download_corpora()
