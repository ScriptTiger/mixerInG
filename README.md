[![Say Thanks!](https://img.shields.io/badge/Say%20Thanks-!-1EAEDB.svg)](https://docs.google.com/forms/d/e/1FAIpQLSfBEe5B_zo69OBk19l3hzvBmz3cOV6ol1ufjh0ER1q3-xd2Rg/viewform)

# Mixer in G
Mixer in G is a PCM audio stream mixer package written in pure Go using the [go-audio](https://github.com/go-audio) framework and does not require binding to non-native libraries, such as SoX or FFmpeg, as other mixers do.

To import Mixer in G into your project:  
`go get github.com/ScriptTiger/mixerInG`  
Then just `import "github.com/ScriptTiger/mixerInG"` and get started with using its functions.

Please refer to the dev package docs and reference implementation for more details and ideas on how to integrate Mixer in G into your project.  

Dev package docs:  
https://pkg.go.dev/github.com/ScriptTiger/mixerInG

Reference implementation:  
https://github.com/ScriptTiger/mixerInG/blob/main/ref/ref.go

# Reference Implementation

Usage: `mixerInG [options...]`

Argument               | Description
-----------------------|--------------------------------------------------------------------------------------------------------
 `-i <file>`           | Input WAV file (must be used for each input, for at least 2 inputs)
 `-o <file>`           | Destination WAV file of mix
 `-bits <number>`      | Bit depth of mix WAV file (16\|24\|32)
 `-attenuate`          | Attenuate linearly to prevent clipping, dividing by number of tracks mixed

If no output file is given, or `-` is given, the mix is written to standard output with a bit depth of 24 and can be piped into FFmpeg, FFplay, VLC, or other compatible media applications.

Attenuating linearly is best done when going from source tracks of lower bit depths to mixes of higher bit depths to diminish resolution loss. This will result in an output mix which will always be perceptually "quieter," but still retain better resolution than if it were a mix of the same or lower bit depth as the original tracks. This is because the mix itself, and attenuation division, is performed in 64-bit floating point and results in trailing digits behind a decimal which will be removed when the mix track is truncated to its destination bit depth, of 32 bits or lower, signed and not floating point. So, mixing to higher bit depths than the original will scale some of that data which would otherwise be behind the decimal to in front of the decimal and preserve it in the resulting mix.

However, it's also important to note that attenuating linearly is only recommended when mixing incohesive tracks, such as more raw and experimental tracks which were not previously mastered together to prevent clipping in the mix already. By not attenuating at all, there will never be any trailing digits behind a decimal since there would be no division occurring, unless scaling down from source tracks of higher bit depths to mixes of lower bit depths, which should also be avoided.

# Other projects using Mixer in G

FLACSFX:  
https://github.com/ScriptTiger/FLACSFX

# More About ScriptTiger

For more ScriptTiger scripts and goodies, check out ScriptTiger's GitHub Pages website:  
https://scripttiger.github.io/

[![Donate](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=MZ4FH4G5XHGZ4)
