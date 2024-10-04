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
https://github.com/ScriptTiger/mixerInG/tree/main/ref

# Reference Implementation

Usage: `mixerInG [options...]`

Argument               | Description
-----------------------|--------------------------------------------------------------------------------------------------------
 `-i <file>`           | Input WAV file (must be used for each input, for at least 2 inputs)
 `-o <file>`           | Destination WAV file of mix
 `-b <number>`         | Bit depth of mix WAV file (16\|24\|32)

If no output file is given, or `-` is given, the mix is written to standard output with a bit depth of 24 and can be piped into VLC or other compatible media applications.

# More About ScriptTiger

For more ScriptTiger scripts and goodies, check out ScriptTiger's GitHub Pages website:  
https://scripttiger.github.io/

[![Donate](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=MZ4FH4G5XHGZ4)
