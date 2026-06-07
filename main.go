package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: aivoice2-srt <directory>")
	}
	dir := os.Args[1]

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("failed to read directory: %v", err)
	}

	wavFiles := map[string]string{}
	txtFiles := map[string]string{}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := filepath.Ext(name)
		base := strings.TrimSuffix(name, ext)
		path := filepath.Join(dir, name)

		switch strings.ToLower(ext) {
		case ".wav":
			wavFiles[base] = path
		case ".txt":
			txtFiles[base] = path
		}
	}

	for base, wavPath := range wavFiles {
		txtPath, ok := txtFiles[base]
		if !ok {
			fmt.Fprintf(os.Stderr, "skip: no matching txt for %s.wav\n", base)
			continue
		}

		duration, err := wavDuration(wavPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip: %s: %v\n", wavPath, err)
			continue
		}

		text, err := os.ReadFile(txtPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip: %s: %v\n", txtPath, err)
			continue
		}

		sentences := splitSentences(string(text))
		if len(sentences) == 0 {
			continue
		}

		srt := generateSRT(sentences, duration)
		outPath := filepath.Join(dir, base+".srt")
		if err := os.WriteFile(outPath, []byte(srt), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error writing %s: %v\n", outPath, err)
			continue
		}
		fmt.Printf("created: %s\n", outPath)
	}
}

func wavDuration(path string) (float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var header struct {
		RiffID      [4]byte
		FileSize    uint32
		WaveID      [4]byte
		FmtID       [4]byte
		FmtSize     uint32
		AudioFormat uint16
		Channels    uint16
		SampleRate  uint32
		ByteRate    uint32
		BlockAlign  uint16
		BitsPerSamp uint16
	}

	if err := binary.Read(f, binary.LittleEndian, &header); err != nil {
		return 0, fmt.Errorf("read header: %w", err)
	}

	if string(header.RiffID[:]) != "RIFF" || string(header.WaveID[:]) != "WAVE" {
		return 0, fmt.Errorf("not a valid WAV file")
	}

	// Skip to "data" chunk
	offset := int64(12 + 8 + int64(header.FmtSize))
	if _, err := f.Seek(offset, 0); err != nil {
		return 0, err
	}

	for {
		var chunkID [4]byte
		var chunkSize uint32
		if err := binary.Read(f, binary.LittleEndian, &chunkID); err != nil {
			return 0, fmt.Errorf("find data chunk: %w", err)
		}
		if err := binary.Read(f, binary.LittleEndian, &chunkSize); err != nil {
			return 0, fmt.Errorf("read chunk size: %w", err)
		}
		if string(chunkID[:]) == "data" {
			seconds := float64(chunkSize) / float64(header.ByteRate)
			return seconds, nil
		}
		if _, err := f.Seek(int64(chunkSize), 1); err != nil {
			return 0, err
		}
	}
}

func splitSentences(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	delimiters := []rune{'。', '！', '？', '!', '?', '.'}
	var sentences []string
	var current strings.Builder

	for _, r := range text {
		current.WriteRune(r)
		if slices.Contains(delimiters, r) {
			s := strings.TrimSpace(current.String())
			if s != "" {
				sentences = append(sentences, s)
			}
			current.Reset()
		}
	}

	if s := strings.TrimSpace(current.String()); s != "" {
		sentences = append(sentences, s)
	}

	return sentences
}

func generateSRT(sentences []string, totalDuration float64) string {
	totalChars := 0
	for _, s := range sentences {
		totalChars += utf8.RuneCountInString(s)
	}

	var sb strings.Builder
	currentTime := 0.0

	for i, s := range sentences {
		charCount := utf8.RuneCountInString(s)
		segDuration := totalDuration * float64(charCount) / float64(totalChars)

		startTime := currentTime
		endTime := currentTime + segDuration
		currentTime = endTime

		fmt.Fprintf(&sb, "%d\n", i+1)
		fmt.Fprintf(&sb, "%s --> %s\n", formatTime(startTime), formatTime(endTime))
		sb.WriteString(s)
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func formatTime(seconds float64) string {
	totalMs := int(math.Round(seconds * 1000))
	h := totalMs / 3600000
	m := (totalMs % 3600000) / 60000
	s := (totalMs % 60000) / 1000
	ms := totalMs % 1000
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}
