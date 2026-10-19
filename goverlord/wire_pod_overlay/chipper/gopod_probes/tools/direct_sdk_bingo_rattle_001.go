// direct_sdk_bingo_rattle_001.go - standalone, decoupled from Wire-Pod's own
// running process. Fires the Bingo sidecar's own rattle sound (audio/rattle.wav,
// streamed via ExternalAudioStreamPlayback) on one robot, over a direct
// connection to the vendored fforchino/vector-go-sdk - the same module
// chipper/go.mod already requires, no new dependency.
//
// Built by copying the proven connect->BehaviorControl-wait->act->release
// skeleton from direct_sdk_brobots_ready_001.go (DECOUPLED_DIRECT_SDK_GOLDEN_PATH_001.md
// names this file explicitly as the template to copy for a different action -
// "swap SayText for whatever other vectorpb call the vendored SDK exposes").
// The audio-chunking logic itself (1024-byte chunks, 30ms pacing, ffmpeg
// fallback conversion, Prepare/Chunk/Complete message shape) is copied
// verbatim from goverlord/runtime/bingo's own voicecommand_lottery.go
// playBingoSound()/loadBingoPCM() - the exact mechanism already proven safe
// in production, every Bingo draw, today - just pointed at an independent
// vector.NewWP(serial) connection instead of the sidecar's package-global
// sdk_wrapper.Robot.
//
// Deliberately does NOT call SayText, WriteColoredText, or any display
// function - VECTORX_BINGO_ERROR915_WRITECOLOREDTEXT_INVESTIGATION_001.md's
// best-evidenced cause for Error 915 (hypothesis H5) is audio-then-display
// racing on one connection with zero drain delay; this tool only ever does
// the audio call, matching the sidecar's own current (post-fix) production
// shape, which has no display call in it at all.
//
// Does not read sdk_wrapper.SDKConfig/GetAudioVolume - those depend on
// sdk_wrapper.InitSDK's own global config bootstrap, never designed to run
// standalone outside the full sidecar init flow. Volume defaults to 100
// (the ExternalAudioStreamPrepare field's own documented max) but is now
// overridable via an optional 3rd CLI arg (2026-07-22, phcal Part C) - see
// main()'s own usage/parsing below. Existing 2-arg callers (this song's own
// run_rattle(), run_bingo_capture_song_001.py) are unaffected: omitting the
// argument still means AudioVolume=100, byte-identical to before.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/gopod_probes/tools/gopoddirect"
)

const audioChunkBytes = 1024
const audioChunkPause = 30 * time.Millisecond

func loadRattlePCM(ctx context.Context, filename string) ([]byte, func(), error) {
	if strings.HasSuffix(filename, ".pcm") || strings.HasSuffix(filename, ".raw") {
		pcm, err := os.ReadFile(filename)
		if err != nil {
			return nil, nil, fmt.Errorf("read pcm audio: %w", err)
		}
		return pcm, nil, nil
	}

	tmpFile, err := os.CreateTemp("", "bingo-rattle-*.pcm")
	if err != nil {
		return nil, nil, fmt.Errorf("create temp pcm: %w", err)
	}
	tmpName := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		return nil, nil, fmt.Errorf("close temp pcm: %w", err)
	}
	cleanup := func() { _ = os.Remove(tmpName) }

	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-i", filename, "-f", "s16le", "-acodec", "pcm_s16le", "-ar", "16000", "-ac", "1", tmpName)
	if output, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("convert rattle audio: %w: %s", err, string(output))
	}

	pcm, err := os.ReadFile(tmpName)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("read converted audio: %w", err)
	}
	return pcm, cleanup, nil
}

func playRattleStream(ctx context.Context, robot *vector.Vector, filename string, volume uint32) error {
	pcm, cleanup, err := loadRattlePCM(ctx, filename)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		return err
	}

	audioClient, err := robot.Conn.ExternalAudioStreamPlayback(ctx)
	if err != nil {
		return fmt.Errorf("open audio stream: %w", err)
	}
	defer func() { _ = audioClient.CloseSend() }()

	if err := audioClient.Send(&vectorpb.ExternalAudioStreamRequest{
		AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamPrepare{
			AudioStreamPrepare: &vectorpb.ExternalAudioStreamPrepare{
				AudioFrameRate: 16000,
				AudioVolume:    volume,
			},
		},
	}); err != nil {
		return fmt.Errorf("prepare audio stream: %w", err)
	}

	for offset := 0; offset < len(pcm); offset += audioChunkBytes {
		end := offset + audioChunkBytes
		if end > len(pcm) {
			end = len(pcm)
		}
		chunk := pcm[offset:end]
		if err := audioClient.Send(&vectorpb.ExternalAudioStreamRequest{
			AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamChunk{
				AudioStreamChunk: &vectorpb.ExternalAudioStreamChunk{
					AudioChunkSizeBytes: uint32(len(chunk)),
					AudioChunkSamples:   chunk,
				},
			},
		}); err != nil {
			return fmt.Errorf("send audio chunk: %w", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(audioChunkPause):
		}
	}

	if err := audioClient.Send(&vectorpb.ExternalAudioStreamRequest{
		AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamComplete{
			AudioStreamComplete: &vectorpb.ExternalAudioStreamComplete{},
		},
	}); err != nil {
		return fmt.Errorf("complete audio stream: %w", err)
	}
	return nil
}

func assumeControlAndRattle(serial, rattlePath string, volume uint32) (ok bool, errMsg string, start, end time.Time) {
	ctx := context.Background()

	// 2026-08-08: connect/grant/release now come from the shared gopoddirect
	// package (PHCAL_SLEEP_HOLD_BARRIER_FIX_001.md's own follow-up) instead
	// of this file's own hand-copied skeleton - see that package's own
	// doc comment for why. Single robot, no barrier needed here.
	robot, stream, err := gopoddirect.ConnectAndGrant(ctx, serial)
	if err != nil {
		return false, err.Error(), start, end
	}

	// 2026-09-30: hold the wheels still before the rattle. The robot is free
	// between the song's release and this grant, and can start driving off its
	// charger (explore) - live, the wheels were heard under the rattle. Same
	// stop the wheel-nudge alias ends with (move_wheels lw=0 rw=0, brobots.sh).
	_, _ = robot.Conn.DriveWheels(ctx, &vectorpb.DriveWheelsRequest{})

	start = time.Now()
	rattleErr := playRattleStream(ctx, robot, rattlePath, volume)
	end = time.Now()

	gopoddirect.Release(stream)

	if rattleErr != nil {
		return false, fmt.Sprintf("rattle: %v", rattleErr), start, end
	}
	return true, "", start, end
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: direct_sdk_bingo_rattle_001 <serial> <rattle_wav_path> [volume 1-100]")
		os.Exit(2)
	}
	serial := os.Args[1]
	rattlePath := os.Args[2]

	// Optional 3rd arg, 2026-07-22 (phcal Part C) - defaults to 100 (the
	// field's own documented max) when omitted, so the existing 2-arg call
	// site (run_rattle(), run_bingo_capture_song_001.py) is byte-identical
	// in behavior, unchanged by this addition.
	volume := uint32(100)
	if len(os.Args) >= 4 {
		parsed, err := strconv.Atoi(os.Args[3])
		if err != nil || parsed < 1 || parsed > 100 {
			fmt.Fprintln(os.Stderr, "usage: direct_sdk_bingo_rattle_001 <serial> <rattle_wav_path> [volume 1-100] - volume must be an integer 1-100")
			os.Exit(2)
		}
		volume = uint32(parsed)
	}

	ok, errMsg, start, end := assumeControlAndRattle(serial, rattlePath, volume)
	status := "OK"
	if !ok {
		status = "FAIL:" + errMsg
	}
	fmt.Printf(
		"BINGO_RATTLE_DIRECT serial=%s start=%.6f end=%.6f elapsed=%.3fs status=%s\n",
		serial,
		float64(start.UnixNano())/1e9, float64(end.UnixNano())/1e9,
		end.Sub(start).Seconds(), status,
	)
	if !ok {
		os.Exit(1)
	}
}
