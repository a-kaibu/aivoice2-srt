package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
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

	var bases []string
	for base := range wavFiles {
		if _, ok := txtFiles[base]; ok {
			bases = append(bases, base)
		} else {
			fmt.Fprintf(os.Stderr, "skip: no matching txt for %s.wav\n", base)
		}
	}
	sort.Slice(bases, func(i, j int) bool {
		return extractNum(bases[i]) < extractNum(bases[j])
	})

	if len(bases) == 0 {
		log.Fatal("no matching wav/txt pairs found")
	}

	var sb strings.Builder
	currentTime := 0.0

	for i, base := range bases {
		duration, err := wavDuration(wavFiles[base])
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip: %s: %v\n", wavFiles[base], err)
			continue
		}

		text, err := os.ReadFile(txtFiles[base])
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip: %s: %v\n", txtFiles[base], err)
			continue
		}

		content := strings.TrimSpace(string(text))
		if content == "" {
			continue
		}

		endTime := currentTime + duration
		fmt.Fprintf(&sb, "%d\n%s --> %s\n%s\n\n", i+1, formatTime(currentTime), formatTime(endTime), content)
		currentTime = endTime
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("failed to resolve path: %v", err)
	}
	dirName := filepath.Base(absDir)
	outPath := filepath.Join(dir, dirName+".srt")
	if err := os.WriteFile(outPath, []byte(sb.String()), 0644); err != nil {
		log.Fatalf("error writing %s: %v", outPath, err)
	}
	fmt.Printf("created: %s\n", outPath)
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

func extractNum(base string) int {
	idx := strings.Index(base, "_")
	if idx <= 0 {
		return math.MaxInt
	}
	n, err := strconv.Atoi(base[:idx])
	if err != nil {
		return math.MaxInt
	}
	return n
}

func formatTime(seconds float64) string {
	totalMs := int(math.Round(seconds * 1000))
	h := totalMs / 3600000
	m := (totalMs % 3600000) / 60000
	s := (totalMs % 60000) / 1000
	ms := totalMs % 1000
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}
