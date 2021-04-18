package cmd

import (
	"log"

	download "github.com/karanbirsingh7/go-youtube-dl/download"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download the video from given URL",
	Long:  `Download the video from given URL`,
	Run: func(cmd *cobra.Command, args []string) {
		url, err := cmd.Flags().GetString("url")
		totalThreads, _ := cmd.Flags().GetInt("threads")
		if err != nil {
			log.Fatal(err)
		}
		if url != "" {
			download.DownloadVideo(url, totalThreads)
		}
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.PersistentFlags().String("url", "", "A direct URL for youtube video to download from")
	downloadCmd.PersistentFlags().Int("threads", 1, "Total number of threads to run")
}
