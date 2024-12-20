package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: fi: [options] [files]\n options: -h, --help: Print this help message\n")
}

// from: https://parsiya.net/blog/2018-11-01-windows-filetime-timestamps-and-byte-wrangling-with-go/
func toTime(t [8]byte) time.Time {
	ft := &syscall.Filetime{
		LowDateTime:  binary.LittleEndian.Uint32(t[:4]),
		HighDateTime: binary.LittleEndian.Uint32(t[4:]),
	}
	return time.Unix(0, ft.Nanoseconds())
}

func printFOSInfo(bytecode []byte) {
	var u16 uint16
	var u32 uint32
	var f32 float32

	reader := bytes.NewReader(bytecode)
	mid := make([]byte, 12)
	binary.Read(reader, binary.BigEndian, &mid)
	binary.Read(reader, binary.LittleEndian, &u32)
	binary.Read(reader, binary.LittleEndian, &u32)
	fmt.Println("Engine Version:", u32)
	binary.Read(reader, binary.LittleEndian, &u32)
	fmt.Println("Save Number:", u32)
	binary.Read(reader, binary.LittleEndian, &u16)
	charName := make([]byte, u16)
	binary.Read(reader, binary.LittleEndian, &charName)
	fmt.Println("Character Name:", string(charName))
	binary.Read(reader, binary.LittleEndian, &u32)
	fmt.Println("Character Level:", u32)
	binary.Read(reader, binary.LittleEndian, &u16)
	charLocation := make([]byte, u16)
	binary.Read(reader, binary.LittleEndian, &charLocation)
	fmt.Println("Character Location:", string(charLocation))
	binary.Read(reader, binary.LittleEndian, &u16)
	charPlaytime := make([]byte, u16)
	binary.Read(reader, binary.LittleEndian, &charPlaytime)
	fmt.Println("Character Playtime:", string(charPlaytime))

	binary.Read(reader, binary.LittleEndian, &u16)
	charRace := make([]byte, u16)
	binary.Read(reader, binary.LittleEndian, &charRace)

	binary.Read(reader, binary.LittleEndian, &u16)
	if u16 == 0 {
		fmt.Println("Character Sex: Male")
	} else {
		fmt.Println("Character Sex: Female")
	}
	binary.Read(reader, binary.LittleEndian, &f32)
	binary.Read(reader, binary.LittleEndian, &f32)

	var filetime [8]byte
	binary.Read(reader, binary.LittleEndian, &filetime)
	t := toTime(filetime)
	fmt.Println("Filetime:", t)

	var snapshotWidth, snapshotHeight uint32
	binary.Read(reader, binary.LittleEndian, &snapshotWidth)
	binary.Read(reader, binary.LittleEndian, &snapshotHeight)
	snapshot := make([]uint8, snapshotWidth*snapshotHeight*4)
	binary.Read(reader, binary.LittleEndian, &snapshot)

	var version uint8
	binary.Read(reader, binary.LittleEndian, &version)
	fmt.Println("Format Version:", version)

	binary.Read(reader, binary.LittleEndian, &u16)
	gameVersion := make([]byte, u16)
	binary.Read(reader, binary.LittleEndian, &gameVersion)
	fmt.Println("Game Version:", string(gameVersion))

	binary.Read(reader, binary.LittleEndian, &u32)

	var pluginCount uint8
	binary.Read(reader, binary.LittleEndian, &pluginCount)
	for i := 0; i < int(pluginCount); i++ {
		binary.Read(reader, binary.LittleEndian, &u16)
		plugin := make([]byte, u16)
		binary.Read(reader, binary.LittleEndian, &plugin)
		fmt.Printf("Plugins [%03d]: %s\n", i, string(plugin))
	}

	// The pre Creation Club update version is version 67 and seems to
	// not have the light plugin data structure in its savefile
	if version == 68 {
		var lightPluginCount uint16
		binary.Read(reader, binary.LittleEndian, &lightPluginCount)
		for i := 0; i < int(lightPluginCount); i++ {
			binary.Read(reader, binary.LittleEndian, &u16)
			plugin := make([]byte, u16)
			binary.Read(reader, binary.LittleEndian, &plugin)
			fmt.Printf("Light Plugins [%05d]: %s\n", i, string(plugin))
		}
	}
}

func main() {
	arg_len := len(os.Args)
	for _, arg := range os.Args {
		if arg == "-h" || arg == "--help" {
			usage()
			os.Exit(1)
		}
	}

	if arg_len <= 1 {
		usage()
		os.Exit(1)
	}

	for i := 1; i < arg_len; i++ {
		arg := os.Args[i]
		fileInfo, err := os.Stat(arg)

		if err != nil {
			panic(err)
		}

		if fileInfo.IsDir() {
			fmt.Fprintf(os.Stderr, "%s: Permission denied", arg)
			continue
		}

		buffer, err := os.ReadFile(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't read file: %s", arg)
			continue
		}

		if strings.Compare(string(buffer[0:12]), "FO4_SAVEGAME") != 0 {
			fmt.Fprintf(os.Stderr, "not a Fallout 4 savefile")
			continue
		}

		fmt.Fprintf(os.Stdout, "== File \"%s\" ==\n", arg)
		printFOSInfo(buffer)
		println()
	}
}
