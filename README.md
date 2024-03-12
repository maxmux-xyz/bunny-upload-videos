# Upload videos to bunny stream quickly and easily

1. Provide a directory of videos to upload
2. We loop through dir, rename files to unique titles, and make an output csv file
   "username,title,extention"
3. Create worker group and upload videos to bunny stream
4. Save the video id and url to the csv file in output dir

Throw your access keys to bunny:

```
LIBRARYIDSTG=""
ACCESSKEYSTG=""

LIBRARYIDPROD=""
ACCESSKEYPROD=""
```

bunny docs:
Create video: https://docs.bunny.net/reference/video_createvideo
Upload video: https://docs.bunny.net/reference/video_uploadvideo
