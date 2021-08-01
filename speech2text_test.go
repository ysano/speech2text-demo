package main

import (
	speech "cloud.google.com/go/speech/apiv1"
	"context"
	"strings"
	"testing"
)

func TestRecognize(t *testing.T) {

	// Context
	ctx := context.Background()

	// Make Speech Client
	client, err := speech.NewClient(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Make Stream
	stream, err := client.StreamingRecognize(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Test getTrans
	str := getTrans(stream, "./testdata/common_voice_ja_25371066.wav")
	if str == "" {
		t.Fatal("no response")
	}
	if got, want := str, "ノートパソコンがない"; !strings.Contains(got, want) {
		t.Errorf("Transcript: got %q; want %q", got, want)
	}
}
