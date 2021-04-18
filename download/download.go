package version

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/kkdai/youtube/v2"
	"github.com/schollz/progressbar/v3"
)

func createTmpDir(folderName string) string {
	tmpDir := fmt.Sprintf("/tmp/%s", folderName)
	if err := os.Mkdir(tmpDir, 0777); err != nil {
		if os.IsExist(err) {
			// fmt.Printf("Directory exists: %s\n", tmpDir)
			os.Chmod(tmpDir, 0777)
		}
	}
	return tmpDir
}

func startVideoDownload(videoID string) error {
	client := youtube.Client{}

	// GET video metadata
	video, err := client.GetVideo(videoID)
	if err != nil {
		log.Fatalf("ERROR getting video metadata: %v", err)
	}

	// PRINT
	printDownloadMetadata(video)

	// start download stream
	resp, err := client.GetStream(video, &video.Formats[0])
	if err != nil {
		log.Fatalf("ERROR downloading video: %v", err)
	}
	defer resp.Body.Close()

	// CREATE TMP dir
	downloadPath := createTmpDir("downloaded")

	// CREATE video file
	file, err := os.Create(fmt.Sprintf("%s/%s.mp4", downloadPath, video.Title))
	if err != nil {
		log.Fatalf("ERROR generating new video file: %v", err)
	}
	defer file.Close()

	// progress bar for download
	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"downloading",
	)

	// read from body -> file .mp4
	_, err = io.Copy(io.MultiWriter(file, bar), resp.Body)
	if err != nil {
		log.Fatalf("ERROR copying to video file: %v", err)
		return err
	}
	fmt.Println("Download complete at: ", file.Name())
	return nil
}

func printDownloadMetadata(video *youtube.Video) {
	fmt.Printf(" ---> DOWNLOADING: %s\n", video.Title)
}

// function that accepts URL and download video
func DownloadVideo(url string, totalThreads int) error {
	fmt.Println("\nDownloading video from: ", url)

	if isPlaylist(url) {
		downloadPlaylist(url, totalThreads)
	} else {
		videoID := getVideoID(url)
		err := startVideoDownload(videoID)
		return err
	}

	return nil
}

// Accepts URL and return video ID
func getVideoID(s string) string {
	u, err := url.Parse(s)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	return strings.TrimPrefix(u.RawQuery, "v=")
}

// download playlist
func downloadPlaylist(url string, totalThreads int) {
	playlistID := getVideoID(url)
	client := youtube.Client{}

	playlist, err := client.GetPlaylist(playlistID)
	if err != nil {
		panic(err)
	}

	/* ----- Enumerating playlist videos ----- */
	header := fmt.Sprintf("Playlist %s by %s", playlist.Title, playlist.Author)
	fmt.Println(header)
	fmt.Println(strings.Repeat("=", len(header)) + "\n")

	// Download with concurrency
	type result struct {
		id  string
		err error
	}

	// initialize a new channel
	maxGoroutines := totalThreads
	guard := make(chan struct{}, maxGoroutines)
	resultsCh := make(chan result)

	for k, v := range playlist.Videos {
		fmt.Printf("(%d) %s - '%s'\n", k+1, v.Author, v.Title)
		// instuct guard that a task is started so that when allocation is reached program is blocked
		guard <- struct{}{}

		// start download of each video
		go func(videoID string) {
			err := startVideoDownload(videoID)
			if err != nil {
				resultsCh <- result{
					err: err,
				}
			}
			// give results back to channel
			fmt.Println("Sending back to RESULTCH")
			resultsCh <- result{
				id: videoID,
			}
			fmt.Println("Sending back to GUARD")
			// once complete empty value from guard
			<-guard
		}(v.ID)
	}

	// pull all values from channel
	var results []result
	for i := 0; i < len(playlist.Videos); i++ {
		results = append(results, <-resultsCh)
	}
	fmt.Println("Total videos downloaded: %v", len(results))

}

// check if URL is for playlist or single video
func isPlaylist(url string) bool {
	videoID := getVideoID(url)
	client := youtube.Client{}

	// GET video metadata
	_, err := client.GetPlaylist(videoID)
	if err != nil {
		return false
	}
	return true
}
