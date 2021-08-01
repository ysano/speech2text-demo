// Speech to Text Demo with GCP
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	speech "cloud.google.com/go/speech/apiv1"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

const (
	// ANSI Color
	colorRed   = "\033[0;31m"
	colorGreen = "\033[0;32m"
	colorBlue  = "\033[0;34m"
	colorNone  = "\033[0m"
)

func main() {
	// Usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <audiofile> [[word1] [word2]...]\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "<audiofile>: 16khz 16bit little endian only")
		fmt.Fprintf(os.Stderr, "[word1 word2..]: Search word(s)")
	}

	// Option Parse
	flag.Parse()
	if len(flag.Args()) == 0 {
		log.Fatal("Specify a local audiofile.")
	}
	args := flag.Args()
	audioFile := args[0]
	words := args[1:]

	// Context
	ctx := context.Background()

	// Make Speech Client
	// TODO: explicitly point to your service account file
	//       with option.WithCredentialsFile(jsonPath)
	client, err := speech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Make Stream
	stream, err := client.StreamingRecognize(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Get Transcript
	transcript := getTrans(stream, audioFile)

	// Output entire transcript
	if len(words) == 0 {
		fmt.Printf("%sSearch word is nil, output entire transcript%s\n", colorGreen, colorNone)
		fmt.Printf("%v\n", transcript)
		return
	}

	// Default word neighbors
	neighbors := 5

	// Output result per word
	for _, w := range words {
		results := parseSingle(transcript, w, neighbors)
		results = formatString(results, w)
		fmt.Printf("%sWord: %v%s\n", colorGreen, w, colorNone)
		for i, v := range results {
			fmt.Printf("%s%0d:%s %v\n", colorGreen, i, colorNone, v[0])
		}
	}
	return
}

func getTrans(stream speechpb.Speech_StreamingRecognizeClient, audioFile string) string {

	// Send configuration
	// TODO: Support for formats other than 16kHz, 16bit
	if err := stream.Send(&speechpb.StreamingRecognizeRequest{
		StreamingRequest: &speechpb.StreamingRecognizeRequest_StreamingConfig{
			StreamingConfig: &speechpb.StreamingRecognitionConfig{
				Config: &speechpb.RecognitionConfig{
					Encoding:        speechpb.RecognitionConfig_LINEAR16,
					SampleRateHertz: 16000,
					LanguageCode:    "ja-JP",
				},
			},
		},
	}); err != nil {
		log.Fatal(err)
	}

	// Open Audio File
	f, err := os.Open(audioFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Goroutine(like thread)
	go func() {
		buf := make([]byte, 1024) // Make slice
		for {
			n, err := f.Read(buf)
			if n > 0 {
				if err := stream.Send(&speechpb.StreamingRecognizeRequest{
					StreamingRequest: &speechpb.StreamingRecognizeRequest_AudioContent{
						AudioContent: buf[:n],
					},
				}); err != nil {
					log.Printf("Could not send audio: %v", err)
				}
			}
			if err == io.EOF {
				if err := stream.CloseSend(); err != nil {
					log.Fatalf("Could not close stream: %v", err)
				}
				return
			}
			if err != nil {
				log.Printf("Could not read from %s: %v", audioFile, err)
				continue
			}
		}
	}() // Invoke

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Cannot stream results: %v", err)
		}
		if err := resp.Error; err != nil {
			log.Fatalf("Could not recognize: %v", err)
		}

		// Print the results.
		for _, result := range resp.Results {
			for _, alt := range result.Alternatives {

				// With confidence
				// fmt.Fprintf(os.Stdout, "\"%v\" (confidence=%3f)\n", alt.Transcript, alt.Confidence)

				// Transcript only
				// fmt.Fprintf(os.Stdout, "%v\n", alt.Transcript)

				return alt.Transcript
			}
		}
	}
	return ""
}

// Full-text search with a single target word
func parseSingle(str string, target string, neighbors int) [][]string {

	// Add dummy space for edge case
	dummy := strings.Repeat(` `, neighbors)
	str = dummy + str + dummy

	// Find Target
	rWithNeighbors := regexp.MustCompile(fmt.Sprintf(`.{%d}%v.{%d}`, neighbors, target, neighbors))
	ret := rWithNeighbors.FindAllStringSubmatch(str, -1)

	// Remove dummy
	rDummy := regexp.MustCompile(`^\s+|\s+$`)
	for _, v := range ret {
		v[0] = rDummy.ReplaceAllString(v[0], "")
		// log.Printf(v[0])
	}

	return ret
}

// ANSI color
func formatString(results [][]string, target string) [][]string {

	rTarget := regexp.MustCompile(target)

	for _, v := range results {
		v[0] = rTarget.ReplaceAllString(v[0], colorRed+"$0"+colorNone)
		// log.Printf(v[0])
	}
	return results
}
