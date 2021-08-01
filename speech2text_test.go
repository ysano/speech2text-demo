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

func TestParseSngle(t *testing.T) {

	str := "私が先生と知り合いになったのは鎌倉かまくらである。その時私はまだ若々しい書生であった。暑中休暇を利用して海水浴に行った友達からぜひ来いという端書はがきを受け取ったので、私は多少の金を工面くめんして、出掛ける事にした。私は金の工面に二に、三日さんちを費やした。ところが私が鎌倉に着いて三日と経たたないうちに、私を呼び寄せた友達は、急に国元から帰れという電報を受け取った。電報には母が病気だからと断ってあったけれども友達はそれを信じなかった。友達はかねてから国元にいる親たちに勧すすまない結婚を強しいられていた。彼は現代の習慣からいうと結婚するにはあまり年が若過ぎた。それに肝心かんじんの当人が気に入らなかった。それで夏休みに当然帰るべきところを、わざと避けて東京の近くで遊んでいたのである。彼は電報を私に見せてどうしようと相談をした。私にはどうしていいか分らなかった。けれども実際彼の母が病気であるとすれば彼は固もとより帰るべきはずであった。それで彼はとうとう帰る事になった。せっかく来た私は一人取り残された。"

	// middle case 1
	res, indices := parseSingle(str, "鎌倉", 5)
	if got, want := res[0], "なったのは鎌倉かまくらで"; got != want {
		t.Errorf("Text: got %q; want %q", got, want)
	}
	if got, want := indices[0], 15; got != want {
		t.Errorf("Text: got %v; want %v", got, want)
	}

	// middle case 2
	res, indices = parseSingle(str, "鎌倉", 5)
	if got, want := res[1], "ころが私が鎌倉に着いて三"; got != want {
		t.Errorf("Text: got %q; want %q", got, want)
	}
	if got, want := indices[1], 135; got != want {
		t.Errorf("Text: got %v; want %v", got, want)
	}

	// edge case
	res, indices = parseSingle(str, "私", 5)

	if got, want := res[0], "私が先生と知"; got != want {
		t.Errorf("Text: got %q; want %q", got, want)
	}
	if got, want := indices[0], 0; got != want {
		t.Errorf("Text: got %v; want %v", got, want)
	}

	// // partial case
	res, indices = parseSingle(str, "先生", 5)
	if got, want := res[0], "私が先生と知り合い"; got != want {
		t.Errorf("Text: got %q; want %q", got, want)
	}
	if got, want := indices[0], 2; got != want {
		t.Errorf("Text: got %v; want %v", got, want)
	}
}
