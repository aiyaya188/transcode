/opt/local/ffmpeg/bin/ffmpeg -i  test.mp4 -i  water.jpg  -filter_complex overlay="(main_w/2)-(overlay_w/2):(overlay_h)"  out1.mp4
/opt/local/ffmpeg/bin/ffmpeg -y -i out1.mp4 -filter:v drawtext="/usr/share/fonts/chinese/simsun.ttc: text=123456 :fontcolor=red:fontsize=42:y=h-line_h-30:x=(tw-mod(5*n\,w+tw*1.8)): shadowx=5: shadowy=5" -codec:v libx264  -codec:a  copy -y  out13.mp4

/opt/local/ffmpeg/bin/ffmpeg -i out13.mp4 -c:v libx264 -c:a aac -strict -2 -hls_list_size 0 -f hls m3u8/output.m3u8
/opt/local/ffmpeg/bin/ffmpeg -y -i  out1.mp4  . '/out1.mp4 -filter:v drawtext="/usr/share/fonts/chinese/simsun.ttc: text='.$text.':fontcolor=red:fontsize=42:y=h-line_h-30:x=(tw-mod(5*n\,w+tw*1.8)): shadowx=5: shadowy=5" -codec:v libx264  -codec:a  copy -y   ' . $mp4Path .'/out13.mp4' ;

