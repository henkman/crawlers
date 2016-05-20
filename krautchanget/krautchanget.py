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
	print("usage: kcrunter [url]")
	sys.exit(0)
url = sys.argv[1]
m = re.search("krautchan.net/([^/]+)/thread-\d+\.html", url)
if not m:
	print("not a valid krautchan url")
	sys.exit(0)
r = requests.get(url)
for m in re.finditer(
	"<span class=\"filename\"><a href=\"/download/([^/]+)/([^\"]+)\"",
	r.text
):
	file = m.group(2)
	if os.path.exists("./" + file):
		print("file "+file+" already exists")
		continue
	print("downloading " + file)
	downloadFile("http://krautchan.net/files/"+m.group(1), file)