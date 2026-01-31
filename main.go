package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/gen2brain/malgo"
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	config := sherpa.VadModelConfig{}

	// Please download silero_vad.onnx from
	// https://github.com/k2-fsa/sherpa-onnx/releases/download/asr-models/silero_vad.onnx

	if FileExists("./silero_vad.onnx") {
		fmt.Println("Use silero-vad")
		config.SileroVad.Model = "./silero_vad.onnx"
		config.SileroVad.Threshold = 0.3
		config.SileroVad.MinSilenceDuration = 0.25
		config.SileroVad.MinSpeechDuration = 0.25
		config.SileroVad.MaxSpeechDuration = 10
		config.SileroVad.WindowSize = 512
	} else {
		fmt.Println("Please download ./silero_vad.onnx")
		return
	}

	config.SampleRate = 16000
	config.NumThreads = 1
	config.Provider = "cpu"
	config.Debug = 1

	windowSize := config.SileroVad.WindowSize

	var bufferSizeInSeconds float32 = 5

	vad := sherpa.NewVoiceActivityDetector(&config, bufferSizeInSeconds)
	defer sherpa.DeleteVoiceActivityDetector(vad)

	buffer := sherpa.NewCircularBuffer(10 * config.SampleRate)
	defer sherpa.DeleteCircularBuffer(buffer)

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>", message)
	})
	chk(err)

	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	var selectedDeviceID malgo.DeviceID
	var zeroDeviceID malgo.DeviceID

	// Detect microphone on Linux - Raspberry Pi
	if runtime.GOOS == "linux" {
		fmt.Println("Capture Devices")
		infos, err := ctx.Devices(malgo.Capture)
		chk(err)

		for i, info := range infos {
			// ignore HDMI input
			if strings.Contains(info.Name(), "generate zero samples (capture)") {
				continue
			}
			e := "ok"
			_, err := ctx.DeviceInfo(malgo.Capture, info.ID, malgo.Shared)
			if err != nil {
				e = err.Error()
			}

			selectedDeviceID = info.ID
			fmt.Printf("    %d: %v, %s, [%s]\n",
				i, info.ID, info.Name(), e)
		}
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = 16000
	deviceConfig.Alsa.NoMMap = 1
	if selectedDeviceID != zeroDeviceID {
		deviceConfig.Capture.DeviceID = selectedDeviceID.Pointer()
	}

	printed := false
	k := 0

	onRecvFrames := func(_, pSample []byte, frameCount uint32) {
		samples := samplesInt16ToFloat(pSample)
		buffer.Push(samples)
		for buffer.Size() >= windowSize {
			head := buffer.Head()
			s := buffer.Get(head, windowSize)
			buffer.Pop(windowSize)

			vad.AcceptWaveform(s)

			if vad.IsSpeech() && !printed {
				printed = true
				log.Print("Detected speech\n")
			}

			if !vad.IsSpeech() {
				printed = false
			}

			for !vad.IsEmpty() {
				speechSegment := vad.Front()
				vad.Pop()

				duration := float32(len(speechSegment.Samples)) / float32(config.SampleRate)

				audio := sherpa.GeneratedAudio{}
				audio.Samples = speechSegment.Samples
				audio.SampleRate = config.SampleRate

				filename := fmt.Sprintf("seg-%d-%.2f-seconds.wav", k, duration)
				ok := audio.Save(filename)
				if ok {
					log.Printf("Saved to %s", filename)
				}

				k += 1

				log.Printf("Duration: %.2f seconds\n", duration)
				log.Print("----------\n")
			}
		}
	}

	captureCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, captureCallbacks)
	chk(err)

	err = device.Start()
	chk(err)

	fmt.Println("Started. Please speak. Press ctrl + C  to exit")
	fmt.Scanln()
	device.Uninit()

}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

func samplesInt16ToFloat(inSamples []byte) []float32 {
	numSamples := len(inSamples) / 2
	outSamples := make([]float32, numSamples)

	for i := 0; i != numSamples; i++ {
		// Decode two bytes into an int16 using bit manipulation
		s16 := int16(inSamples[2*i]) | int16(inSamples[2*i+1])<<8
		outSamples[i] = float32(s16) / 32768
	}

	return outSamples
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return false
}
