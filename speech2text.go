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

	speech "cloud.google.com/go/speech/apiv1"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

func main() {
	// Usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <audiofile>\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "<audiofile>: 16khz 16bit little endian only")
	}

	// Option Parse
	flag.Parse()
	if len(flag.Args()) != 1 {
		log.Fatal("Specify a local audiofile.")
	}
	audioFile := flag.Arg(0)

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
	fmt.Fprintf(os.Stdout, "%v\n", transcript)
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
