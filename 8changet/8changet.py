#!C:\Python3\python.exe

import requests
import re
import os
import sys

def downloadFile(url, file):
	r = requests.get(url, stream = True)
	with open(file, 'wb') as f:
		for chunk in r.iter_content(chunk_size=8 * 1024): 
			if chunk:
				f.write(chunk)
		f.flush()

if len(sys.argv) != 2:
	print("usage: 8chget [url]")
	sys.exit(0)
url = sys.argv[1]

m = re.search(r"8ch.net/([^/]+)/res/\d+.html$", url)
if not m:
	print("not a valid 8chan url")
	sys.exit(0)
board = m.group(1)
r = requests.get(url)
for m in re.finditer(
	"File: <a href=\"(https?://media.8ch.net/[^\"]+)\">(.*?)</a>",
	r.text
):
	file = m.group(2)
	if os.path.exists("./" + file):
		print("file "+file+" already exists")
		continue
	print("downloading " + file)
	downloadFile(m.group(1), file)