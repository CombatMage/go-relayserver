ECHO ffmpeg -i Wildlife.wmv -f mpegts -codec:v mpeg1video -s 640x480 -r 30 -b:v 1000k -b 0 http://localhost:8081/secret1234

ffmpeg -i SampleVideo_1280x720_30mb.mp4 -f mpegts -codec:v mpeg1video -s 1280x720 -rtbufsize 2048M -r 30 -b:v 3000k  -q:v 6 http://localhost:8081/secret1234