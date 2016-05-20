#!C:\Python3\python.exe

import requests
import re
import os
import sys
import json

if len(sys.argv) != 2:
	print("usage: hitboxget [url]")
	sys.exit(0)
url = sys.argv[1]

m = re.search(r"hitbox.tv/([^/]+)/", url)
if not m:
	print("not a valid hitbox url")
	sys.exit(0)
username = m.group(1)
r = requests.get("https://www.hitbox.tv/api/media/video/{0}/list?filter=&limit=0&media=true&size=mid&start=0".format(username))
vidjson = json.loads(r.text)
for vid in vidjson["video"]:
	filename = re.sub(r'[/\\?%*:|"<>]', '_', vid["media_title"]) + ".ts"
	if os.path.exists(filename):
		print(filename + " already exists")
		continue
	profiles = json.loads(vid["media_profiles"])
	if len(profiles) < 1:
		continue
	media = profiles[0]
	vodurl = "http://hitboxht.vo.llnwd.net/e2/static/videos/vods{0}".format(media["url"])
	ls = vodurl.rfind("/")
	vodbase = vodurl[:ls]
	r = requests.get(vodurl)
	vod = r.text
	print("downloading " + filename)
	with open(filename, 'wb') as f:
		for m in re.finditer("index\d+\.ts", vod):
			r = requests.get(vodbase+"/"+m.group(0), stream = True)
			for chunk in r.iter_content(chunk_size=8 * 1024): 
				if chunk:
					f.write(chunk)
			f.flush()