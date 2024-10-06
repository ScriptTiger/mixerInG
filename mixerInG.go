package mixerInG

import (
	"errors"
	"math"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

// Structure to hold information for a track
type TrackInfo struct {
	BitDepth int
	BufferSize int
	IntBuffer *audio.IntBuffer
	FloatBuffer *audio.FloatBuffer
}

// Function to create a new TrackInfo
func newTrack(format *audio.Format, bitDepth, bufferCap int) (newTrack *TrackInfo) {
	return &TrackInfo{
		BitDepth: bitDepth,
		BufferSize: -1,
		IntBuffer: &audio.IntBuffer{Format: format, Data: make([]int, bufferCap)},
		FloatBuffer: &audio.FloatBuffer{Format: format, Data: make([]float64, bufferCap)},
	}
}

// Function to mix TrackInfo float buffers to a provided mix float buffer, performing common scaling and attenuation operations as needed, and return length of longest buffer
func Mix(mixTrack *audio.FloatBuffer, sourceTracks []*TrackInfo, bitDepth int, attenuate bool) (mixBufferSize int) {
	ScaleFloatBuffers(sourceTracks, bitDepth)
	mixBufferSize = SumFloatBuffers(mixTrack, sourceTracks)
	if attenuate {AttenuateFloatBuffer(mixTrack, len(sourceTracks), mixBufferSize)}
	if mixBufferSize < cap(mixTrack.Data) {mixTrack.Data = mixTrack.Data[:mixBufferSize]}
	return mixBufferSize
}

// Function to sum TrackInfo float buffers to a mix float buffer and return buffer size of mix, equal to length of longest buffer
func SumFloatBuffers(mixTrack *audio.FloatBuffer, sourceTracks []*TrackInfo) (mixBufferSize int) {
	for _, sourceTrack := range sourceTracks {
		if sourceTrack.BufferSize == 0 {continue}
		if mixBufferSize == 0 {
			for m, _ := range mixTrack.Data {mixTrack.Data[m] = 0}
			for s, data := range sourceTrack.FloatBuffer.Data {mixTrack.Data[s] = data}
			mixBufferSize = sourceTrack.BufferSize
			continue
		}
		for s := 0; s < sourceTrack.BufferSize; s++ {
			mixTrack.Data[s] = mixTrack.Data[s]+sourceTrack.FloatBuffer.Data[s]
		}
		if sourceTrack.BufferSize > mixBufferSize {mixBufferSize = sourceTrack.BufferSize}
	}
	return mixBufferSize
}

// Function to attenuate linearly to prevent clipping in real-time without knowing/using true peak, RMS, LUFS, etc.
func AttenuateFloatBuffer(mixTrack *audio.FloatBuffer, numTracks, bufferSize int) {
	for i := 0; i < bufferSize; i++ {mixTrack.Data[i] = mixTrack.Data[i]/float64(numTracks)}
}

// Function to scale input bit depth to output bit depth
func ScaleFloatBuffers(tracks []*TrackInfo, bitDepth int) {
	for _, track := range tracks {
		if track.BufferSize != 0 && track.BitDepth != bitDepth {
			for i := 0; i < track.BufferSize; i++ {
				track.FloatBuffer.Data[i] = track.FloatBuffer.Data[i]*math.Pow(2, float64(bitDepth-track.BitDepth))
			}
		}
	}
}

// Function to read PCM data from a wav decoder into a TrackInfo's buffers, set bufferSize, and return length of longest buffer
func ReadWavsToBuffers(wavDecs []*wav.Decoder, tracks []*TrackInfo) (mixBufferSize int) {
	var err error
	for i, track := range tracks {
		if track.BufferSize == 0 {continue}
		if track.BufferSize != -1 && track.BufferSize < cap(track.IntBuffer.Data) {
			track.BufferSize = 0
			continue
		}
		track.BufferSize, err = wavDecs[i].PCMBuffer(track.IntBuffer)
		if track.BufferSize == 0 || err != nil {continue}
		if track.BufferSize < cap(track.IntBuffer.Data) {
			track.IntBuffer.Data = track.IntBuffer.Data[:track.BufferSize]
			track.FloatBuffer.Data = track.FloatBuffer.Data[:track.BufferSize]
		}
		track.FloatBuffer = track.IntBuffer.AsFloatBuffer()
		if track.BufferSize > mixBufferSize {mixBufferSize = track.BufferSize}
	}
	return mixBufferSize
}

// Function to mix wav decoders and write mix to output
func MixWavDecoders(wavDecs []*wav.Decoder, wavOut *os.File, bitDepth int, attenuate bool) (error) {

	var (
		format *audio.Format
		sampleRate int
		numChans int
	)

	index := make([]*TrackInfo, len(wavDecs))
	bufferCap := 8000

	// Validate tracks, populate index and convenience variables
	for i, wavDec := range wavDecs {

		if !wavDec.IsValidFile() {return errors.New("Invalid file")}

		if wavDec.WavAudioFormat == 3 {return errors.New("IEEE float is not currently supported")}
		if wavDec.WavAudioFormat == 6 {return errors.New("A-law is not currently supported")}
		if wavDec.WavAudioFormat == 7 {return errors.New("µ-law is not currently supported")}
		if wavDec.WavAudioFormat != 1 && wavDec.WavAudioFormat != 0xFFFE {return errors.New("Only PCM is currently supported")}

		if i == 0 {
			format = wavDec.Format()
			sampleRate = int(wavDec.SampleRate)
			numChans = int(wavDec.NumChans)
		} else if sampleRate != int(wavDec.SampleRate) {return errors.New("Sample rate mismatch")
		} else if numChans != int(wavDec.NumChans) {return errors.New("Channel layout mismatch")}

		index[i] = newTrack(format, int(wavDec.BitDepth), bufferCap)
	}

	// Initialize wav encoder
	wavEnc := wav.NewEncoder(
		wavOut,
		sampleRate,
		bitDepth,
		numChans,
		1,
	)
	defer wavEnc.Close()

	// Initialize mix buffer
	mixFloatBuffer := &audio.FloatBuffer{Format: format, Data: make([]float64, bufferCap)}

	// Systematically load buffers, scale if needed, sum all tracks to mix buffer, adjust gain if requested, and write to output
	for {
		mixBufferSize := ReadWavsToBuffers(wavDecs, index)
		if mixBufferSize == 0 {break}
		Mix(mixFloatBuffer, index, bitDepth, attenuate)
		wavEnc.Write(mixFloatBuffer.AsIntBuffer())
		if mixBufferSize < bufferCap {break}
	}

	return nil

}

// Function to mix wav files and write mix to output
func MixWavFiles(files []*string, outWavName *string, bitDepth int, attenuate bool) (error) {

	// Initialize wavDecs to number of tracks
	wavDecs := make([]*wav.Decoder, len(files))

	// Initialize slice for wave decoders
	for i, file := range files {
		wavFile, err := os.Open(*file)
		if err != nil {return err}
		defer wavFile.Close()
		wavDecs[i] = wav.NewDecoder(wavFile)
	}

	// Initialize output
	var wavOut *os.File
	var err error
	if *outWavName == "-" {wavOut = os.Stdout
	} else {
		wavOut, err = os.Create(*outWavName)
		if err != nil {return err}
		defer wavOut.Close()
	}

	// Mix decoders
	err = MixWavDecoders(wavDecs, wavOut, bitDepth, attenuate)
	if err != nil {return err}

	return nil
}