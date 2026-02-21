* Try downloading only the audio
* Try extracting the first 3 minutes and the last 1 minute in images.

  ``` bash
  yt-dlp --keep-video --extract-audio --audio-format mp3 https://www.youtube.com/watch?v=ID -o output.mp3

  DURATION=$(ffprobe -v error -show_entries format=duration \
    -of default=noprint_wrappers=1:nokey=1 output.webm | cut -d. -f1)

  mkdir -p frames

  ffmpeg -i output.webm -t 180 -vf fps=1 frames/first_%04d.png
  ffmpeg -ss $((DURATION - 60)) -i output.webm -vf fps=1 frames/last_%04d.png
  ```

* Try prompt with audio and images but write a rich md file: with:
  specs, context box, short summary and some other rich content.

* Uncomment and remove code in:
  - `internal/integrations/gemini/gemini.go`
  - `internal/worker/worker.go`
  - `internal/utils/utils.go`

* Write tests
* Add slug input to page
* Add form for new category as well as delete, edit category routes
* Delete source - should cascade and delete all the videos?

* Use logger
  https://go.dev/blog/slog

* Minify CSS and JS files during development.
  Calculate just etags on compile or on the fly in the route

* Refactor JS in functions and classes
  Use TypeScript
