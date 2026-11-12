package videobuffer

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Timur1414/asvt/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Config struct {
	StreamURL       string
	BufferDir       string
	SegmentDuration int
	Retention       time.Duration
}

type VideoExtractor interface {
	ExtractClip(ctx context.Context, startTime, endTime time.Time) (string, error)
}

type HLSSegment struct {
	Filename  string
	StartTime time.Time
	Duration  time.Duration
}

type VideoBuffer struct {
	cfg        Config
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
}

func NewVideoBuffer(cfg Config) *VideoBuffer {
	if cfg.BufferDir == "" {
		cfg.BufferDir = "/app/buffer"
	}
	if cfg.SegmentDuration <= 0 {
		cfg.SegmentDuration = 2
	}
	if cfg.Retention <= 0 {
		cfg.Retention = 10 * time.Minute
	}
	return &VideoBuffer{
		cfg: cfg,
	}
}

func (obj *VideoBuffer) Start(parentCtx context.Context) {
	log := logger.GetLogger()
	if obj.cfg.StreamURL == "" {
		log.Warn("CAMERA_STREAM_URL is not set, video buffering is disabled")
		return
	}
	if err := os.MkdirAll(obj.cfg.BufferDir, 0755); err != nil {
		log.Error("failed to create buffer directory", zap.String("dir", obj.cfg.BufferDir), zap.Error(err))
		return
	}
	ctx, cancel := context.WithCancel(parentCtx)
	obj.cancelFunc = cancel
	obj.wg.Add(2)
	go obj.recordStreamLoop(ctx)
	go obj.cleanupLoop(ctx)
	log.Info("VideoBuffer service started",
		zap.String("stream", obj.cfg.StreamURL),
		zap.String("dir", obj.cfg.BufferDir),
		zap.Duration("retention", obj.cfg.Retention),
	)
}

func (obj *VideoBuffer) Stop() {
	if obj.cancelFunc != nil {
		obj.cancelFunc()
		obj.wg.Wait()
	}
}

func (obj *VideoBuffer) recordStreamLoop(ctx context.Context) {
	defer obj.wg.Done()
	log := logger.GetLogger()
	maxSegments := int(obj.cfg.Retention.Seconds()) / obj.cfg.SegmentDuration
	if maxSegments < 30 {
		maxSegments = 30
	}
	playlistPath := filepath.Join(obj.cfg.BufferDir, "playlist.m3u8")
	segmentPattern := filepath.Join(obj.cfg.BufferDir, "segment_%05d.ts")
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		log.Info("Starting ffmpeg HLS recorder process")
		args := []string{
			"-nostdin",
			"-hide_banner",
			"-loglevel", "warning",
			"-rtsp_transport", "tcp",
			"-i", obj.cfg.StreamURL,
			"-c", "copy",
			"-f", "hls",
			"-hls_time", strconv.Itoa(obj.cfg.SegmentDuration),
			"-hls_list_size", strconv.Itoa(maxSegments),
			"-hls_flags", "delete_segments+program_date_time",
			"-hls_segment_type", "mpegts",
			"-hls_segment_filename", segmentPattern,
			playlistPath,
		}
		cmd := exec.CommandContext(ctx, "ffmpeg", args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil && ctx.Err() == nil {
			log.Warn("FFmpeg recorder stopped, restarting in 3s",
				zap.Error(err),
				zap.String("stderr", stderr.String()),
			)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

func (obj *VideoBuffer) cleanupLoop(ctx context.Context) {
	defer obj.wg.Done()
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			files, err := os.ReadDir(obj.cfg.BufferDir)
			if err != nil {
				continue
			}
			cutoff := time.Now().Add(-obj.cfg.Retention * 2)
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				name := file.Name()
				if strings.HasPrefix(name, "segment_") || strings.HasPrefix(name, "alert_") || strings.HasPrefix(name, "concat_") {
					info, err := file.Info()
					if err == nil && info.ModTime().Before(cutoff) {
						_ = os.Remove(filepath.Join(obj.cfg.BufferDir, name))
					}
				}
			}
		}
	}
}

func (obj *VideoBuffer) ExtractClip(ctx context.Context, startTime, endTime time.Time) (string, error) {
	playlistPath := filepath.Join(obj.cfg.BufferDir, "playlist.m3u8")
	data, err := os.ReadFile(playlistPath)
	if err != nil {
		return "", fmt.Errorf("failed to read buffer playlist: %w", err)
	}
	segments, err := ParsePlaylist(string(data), obj.cfg.BufferDir)
	if err != nil || len(segments) == 0 {
		return "", fmt.Errorf("no segments found in playlist: %w", err)
	}
	startTime = startTime.UTC()
	endTime = endTime.UTC()
	if endTime.Before(startTime) {
		return "", fmt.Errorf("invalid interval: endTime before startTime")
	}
	var selected []HLSSegment
	for _, seg := range segments {
		segEnd := seg.StartTime.Add(seg.Duration)
		if segEnd.After(startTime) && seg.StartTime.Before(endTime) {
			if _, err := os.Stat(seg.Filename); err == nil {
				selected = append(selected, seg)
			}
		}
	}
	if len(selected) == 0 {
		return "", fmt.Errorf("no segments matching range %s - %s",
			startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))
	}

	firstSegStart := selected[0].StartTime
	offset := startTime.Sub(firstSegStart)
	if offset < 0 {
		offset = 0
	}
	duration := endTime.Sub(startTime)
	if duration <= 0 {
		duration = 5 * time.Second
	}
	concatFileName := filepath.Join(obj.cfg.BufferDir, fmt.Sprintf("concat_%s.txt", uuid.New().String()))
	var concatContent strings.Builder
	for _, seg := range selected {
		escapedPath := strings.ReplaceAll(seg.Filename, "\\", "/")
		concatContent.WriteString(fmt.Sprintf("file '%s'\n", escapedPath))
	}
	if err := os.WriteFile(concatFileName, []byte(concatContent.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write concat list: %w", err)
	}
	defer os.Remove(concatFileName)

	outputFileName := filepath.Join(obj.cfg.BufferDir, fmt.Sprintf("alert_%s.mp4", uuid.New().String()))
	args := []string{
		"-y",
		"-ss", fmt.Sprintf("%.3f", offset.Seconds()),
		"-f", "concat",
		"-safe", "0",
		"-i", concatFileName,
		"-t", fmt.Sprintf("%.3f", duration.Seconds()),
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		outputFileName,
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		_ = os.Remove(outputFileName)
		return "", fmt.Errorf("ffmpeg clip extraction error: %w, stderr: %s", err, stderr.String())
	}
	return outputFileName, nil
}

func ParsePlaylist(content string, bufferDir string) ([]HLSSegment, error) {
	var segments []HLSSegment
	var currentProgramTime time.Time
	var currentDuration time.Duration
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#EXT-X-PROGRAM-DATE-TIME:") {
			timeStr := strings.TrimPrefix(line, "#EXT-X-PROGRAM-DATE-TIME:")
			parsedTime, err := parseISOTime(strings.TrimSpace(timeStr))
			if err == nil {
				currentProgramTime = parsedTime
			}
		} else if strings.HasPrefix(line, "#EXTINF:") {
			inf := strings.TrimPrefix(line, "#EXTINF:")
			if idx := strings.Index(inf, ","); idx != -1 {
				inf = inf[:idx]
			}
			sec, err := strconv.ParseFloat(strings.TrimSpace(inf), 64)
			if err == nil {
				currentDuration = time.Duration(sec * float64(time.Second))
			}
		} else if !strings.HasPrefix(line, "#") {
			segFile := line
			if !filepath.IsAbs(segFile) {
				segFile = filepath.Join(bufferDir, segFile)
			}
			if !currentProgramTime.IsZero() {
				segments = append(segments, HLSSegment{
					Filename:  segFile,
					StartTime: currentProgramTime,
					Duration:  currentDuration,
				})
				currentProgramTime = currentProgramTime.Add(currentDuration)
			}
		}
	}
	return segments, nil
}

func parseISOTime(str string) (time.Time, error) {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05.999Z07:00",
		"2006-01-02T15:04:05.999",
		"2006-01-02T15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, str); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse ISO time: %s", str)
}
