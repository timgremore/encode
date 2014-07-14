package main

import (
  "os"
  "os/exec"
  "io/ioutil"
  // "bytes"
  "log"
  "github.com/codegangsta/cli"
  "fmt"
  "strings"
  "regexp"
  "path/filepath"
  "bitbucket.org/pkg/inflect"
)

func main() {
  app := cli.NewApp()
  app.Name = "encode"
  app.Usage = "Prepare one or more videos for HTML5"
  app.Commands = []cli.Command{
    {
      Name: "batch",
      Usage: "Encode all video files at the given path. Special thanks to http://diveintohtml5.info/video.html as this library is entirely guided by the recommendations in that article.",
      Flags: []cli.Flag{
        cli.StringFlag{"path", ".", "input path"},
        cli.StringFlag{"destination", "output", "output path"},
        cli.StringFlag{"formats", "mp4 webm ogg ogv wmv", "scan for and encode these file formats"},
        cli.BoolFlag{"pretend", "just pretend like you're going to encode"},
        cli.BoolFlag{"skip-ogg", "skip generating an Ogg Theora"},
        cli.BoolFlag{"skip-mp4", "skip generating a MP4"},
        cli.BoolFlag{"skip-webm", "skip generating a WebM"},
        cli.BoolFlag{"camelcase", "rename files using camel-case (myVideoFile)"},
        cli.BoolFlag{"skip-rename", "don't rename each video file"},
        cli.BoolFlag{"html", "generate index.html for each video"},
        cli.BoolFlag{"html-only", "generate index.html for each video and do not encode any videos"},
      },
      Action: encodeFiles,
    },
  }
  app.Run(os.Args)
}

func verifyPath(path string) (string, error) {
  // Verify that path is a valid path.
  _, err := os.Stat(path)

  return path, err
}

func createDirectory(path string, force bool) (string, error) {

  // Declare error return from os.Stat
  var err error

  // Does the path exist?
  // Then remove it if forced to do so
  // Else create it
  if _, err = os.Stat(path); err != nil {

    // Does the path already exist?
    // Then remove it if forced
    if os.IsExist(err) && force {
      return path, os.RemoveAll(path)
    }

    // Create directory and parents if path does not exist
    if os.IsNotExist(err) {
      return path, os.MkdirAll(path, 0777)
    }
  }

  return path, err
}

// Convert string into a regexp
// String is assumed to contain white space separated values
// that are to be considered optional in regexp form
func stringToRegex(value string) *regexp.Regexp {

  // Accept space delimited formats and replace with pipes
  formats := strings.Replace(value, " ", "|", -1)

  // Prepare formats for regex compilation
  formats = strings.Join([]string{"(?i)(", formats, ")$"}, "")

  // Match all accepted video file formats according to file extensions
  regexp := regexp.MustCompile(formats)

  return regexp
}

// Generate an html file with video tags
func createIndexFile(destination string, filename string) {

  // Build html
  html := `<!DOCTYPE html>
<html>
  <head>
    <title>` +  inflect.Titleize(filename) + `</title>
    <style>
      div {
        margin: 0 auto;
      }
    </style>
  </head>
  <body>
    <div>
    <video controls>
`

  // Collect all video source tags
  videoTags := make([]string, 0)

  // Traverse all directories and collect all videos that need encoding
  _ = filepath.Walk(destination, func(path string, _ os.FileInfo, _ error) error {

    if strings.HasSuffix(path, filename + ".mp4") {
      videoTags = append(videoTags, `      <source src="` + filename + `.mp4" type="video/mp4; codecs=,vorbis">`)
    }

    if strings.HasSuffix(path, filename + ".webm") {
      videoTags = append(videoTags, `      <source src="` + filename + `.webm" type="video/webm; codecs=vp8,vorbis">`)
    }

    if strings.HasSuffix(path, filename + ".ogg") {
      videoTags = append(videoTags, `      <source src="` + filename + `.ogg" type="video/ogg; codecs=theora,vorbis">`)
    }

    // return nil since filepath.Walk expects a return
    return nil
  })

  html = html + strings.Join(videoTags, "\n") + `
    </video>
    </div>
  </body>
</html>
`

  // Write html to index.html
  ioutil.WriteFile(filepath.Join(destination, "index.html"), []byte(html), 0755)
}

func encodeFiles(c *cli.Context) {

  // Initialize the path of the files to be encoded
  path, _ := verifyPath(c.String("path"))

  // Default destination for new files
  destination := "./output"

  // Check if destination flag was sent and set accordingly
  if c.IsSet("destination") {
    destination, _ = createDirectory(c.String("destination"), c.Bool("force"))
  }

  // Initialize slice to store all files to be encoded
  files := make([]string, 0)

  // Initialize the file formats to scan for and encode
  formats := "mp4 webm ogg ogv wmv"

  // Check if formats flag was sent and set formats accordingly
  if c.IsSet("formats") {
    formats = c.String("formats")
  }

  // Check for presence of the ffmpeg2theora
  ffmpeg2theoraPath, ffmpeg2theoraErr := exec.LookPath("ffmpeg2theora")
  if ffmpeg2theoraErr != nil  {
    log.Fatal("Please ensure ffmpeg2theora is in your path. brew install ffmpeg2theora if you enjoy Homebrew.")
  }
  fmt.Printf("ffmpeg2theora was found at %s\n", ffmpeg2theoraPath)

  ffmpegPath, ffmpegErr := exec.LookPath("ffmpeg")
  if ffmpegErr != nil  {
    log.Fatal("Please ensure ffmpeg is in your path. brew install ffmpeg if you enjoy Homebrew.")
  }
  fmt.Printf("ffmpeg was found at %s\n", ffmpegPath)

  // Convert formats string into regexp
  regx := stringToRegex(formats)

  // Traverse all directories and collect all videos that need encoding
  _ = filepath.Walk(path, func(path string, _ os.FileInfo, _ error) error {

    // Append this file to the list of files if there is a match
    if regx.MatchString(path) {
      files = append(files, path)
    }

    // return nil since filepath.Walk expects a return
    return nil
  })

  // Loop through each entry in files and process accordingly
  for _, value := range files {

    // Remember the original file name
    originalFileName := filepath.Base(value)

    // Remember the original file extension
    originalFileExtension := filepath.Ext(originalFileName)

    // Extract the file name from the path
    newFileBaseName := strings.TrimSuffix(filepath.Base(value), originalFileExtension)

    // Build the new file name according to flags sent
    if c.IsSet("camelcase") {
      newFileBaseName = inflect.CamelizeDownFirst(newFileBaseName)
    } else if !c.IsSet("skip-rename") {
      newFileBaseName = inflect.Parameterize(newFileBaseName)
    }

    // Build the destination to store this file's encoded files
    encodedFileDestination := filepath.Join(destination, newFileBaseName)

    // Create the new directory to store each file
    _, _ = createDirectory(encodedFileDestination, c.Bool("force"))

    // If request is to generate html only, skip encoding
    if !c.IsSet("html-only") {

      // Encode ogg unless we are to skip ogg
      if !c.IsSet("skip-ogg") {

        // Use ffmpeg2theora to generate an ogg file
        // --videoquality [0-10] (default 6)
        // --audioquality [-2-10] (default 1)
        execute("ffmpeg2theora",
                c.Bool("pretend"),
                "--videoquality",
                "7",
                "--audioquality",
                "7",
                "--output",
                filepath.Join(encodedFileDestination, newFileBaseName + ".ogg"),
                value)
      }

      // Encode mp4 unless we are to skip MP4
      if !c.IsSet("skip-mp4") {

        // Use ffmpeg to generate an mp4 file
        // Settings overview here: https://trac.ffmpeg.org/wiki/Encode/H.264
        execute("ffmpeg",
                c.Bool("pretend"),
                "-i",
                value,
                "-c:v",
                "libx264",
                "-preset",
                "slow",
                "-crf",
                "10",
                "-c:a",
                "copy",
                "-y",
                filepath.Join(encodedFileDestination, newFileBaseName + ".mp4"))
      }

      // Encode WebM unless we are to skip WebM
      if !c.IsSet("skip-webm") {

        // Use ffmpeg to generate a WebM file
        // Settings overview here: https://trac.ffmpeg.org/wiki/Encode/VP8
        execute("ffmpeg",
                c.Bool("pretend"),
                "-i",
                value,
                "-c:v",
                "libvpx",
                "-crf",
                "10",
                "-b:v",
                "1M",
                "-c:a",
                "libvorbis",
                "-y",
                filepath.Join(encodedFileDestination, newFileBaseName + ".webm"))
      }
    }

    // If html, then generate an index.html file for this video
    if c.IsSet("html") || c.IsSet("html-only") {
      createIndexFile(encodedFileDestination, newFileBaseName)
    }
  }

  s := []string{"Encode video files within path: ", path}
  fmt.Println(strings.Join(s, " "))
}

func execute(path string, pretend bool, args ...string) {

  // Build the Command
  cmd := exec.Command(path, args...)

  // Print the command
  fmt.Println(cmd)

  // If pretend was sent, just print the command, don't run it
  if !pretend {
    output, _ := cmd.CombinedOutput()
    fmt.Printf("%s\n", output)
  }
}
