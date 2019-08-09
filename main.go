package main

import (
	"fmt"
	//"io"
	"io/ioutil"
	"os"
)

type Stream struct {
	data []byte
	position int
}

func (s *Stream) ReadByte() byte {
	s.position += 1
	return s.data[s.position-1]
}

func (s *Stream) ReadInt16() int {
	s.position += 2
	return int(s.data[s.position-2]) + int(s.data[s.position-1]) << 8
}

func (s *Stream) ReadInt32() int {
	s.position += 4
	return int(s.data[s.position-4]) | int(s.data[s.position-3]) << 8 | int(s.data[s.position-2]) << 16 | int(s.data[s.position-1]) << 24
}

func (s *Stream) ReadInt64() int {
	s.position += 8
	return int(s.data[s.position-8]) | int(s.data[s.position-7]) << 8 | int(s.data[s.position-6]) << 16 | int(s.data[s.position-5]) << 24 | int(s.data[s.position-4]) << 32 | int(s.data[s.position-3]) << 40 | int(s.data[s.position-2]) << 48 | int(s.data[s.position-1]) << 56
}

func (s *Stream) ReadVarInt() int {
	count := 0
	shift := uint(0)
	b := byte(0)
	for ok := true; ok; ok = ((b & 0x80) != 0) {
		b = s.ReadByte()
		count |= int((b & 0x7F) << shift)
		shift += 7
	}
	return count
}

func (s *Stream) ReadString() string {
	if s.ReadByte() != byte(0x0b) {
		return ""
	}
	len := s.ReadVarInt()
	b := make([]byte, len)
	copy(b, s.data[s.position:s.position+len])
	s.position += len
	return string(b)
}

func (s *Stream) ReadBytes(count int) []byte {
	b := make([]byte, count)
	copy(b, s.data[s.position:s.position+count])
	return b
}

func main() {
	args := os.Args[1:]
	
	for i := 0; i < len(args); i++ {
		r, err := ReadReplay(args[i])
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		
		err = ioutil.WriteFile("Raw-" + args[i], r.ReplayData, 0644)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		fmt.Println(fmt.Sprintf("Processed Replay: %v (%d/%d)", args[i], i+1, len(args)))
	}
	
	/*fmt.Println(r.Lifebar)
	fmt.Println(r.TimeTicks)
	fmt.Println(r.ReplayLength)
	fmt.Println(r.ReplayData)
	fmt.Println(r.ReplayId)*/
}

type Replay struct {
	Gamemode byte
	OsuVersion int
	MapHash, PlayerName, ReplayHash string
	Count300, Count100, Count50, CountGeki, CountKatu, CountMiss, Combo int
	Score, Mods int
	FullCombo bool
	
	Lifebar string
	TimeTicks int
	ReplayLength int
	ReplayData []byte
	ReplayId int
}

func (r Replay) String() string {
	return fmt.Sprintf(`Gamemode: %d
osu! Version: %d
Map Hash: %v
Player: %v
Score Hash: %v
%d 300s, %d 100s, %d 50s
%d Geki, %d Katu, %d Misses
Score: %d
Combo: %d, Full Combo: %t
Mods used: %d`, r.Gamemode, r.OsuVersion, r.MapHash, r.PlayerName, r.ReplayHash, r.Count300, r.Count100, r.Count50, r.CountGeki, r.CountKatu, r.CountMiss, r.Score, r.Combo, r.FullCombo, r.Mods)
}

func ReadReplay(path string) (Replay, error) {
	var r Replay
	
	f, err := ioutil.ReadFile(path) // f == byte array
	
	if err != nil {
		return r, err
	}

	var ms = Stream{f, 0}
	
	r.Gamemode = ms.ReadByte()
	r.OsuVersion = ms.ReadInt32()
	r.MapHash = ms.ReadString()
	r.PlayerName = ms.ReadString()
	r.ReplayHash = ms.ReadString()
	r.Count300 = ms.ReadInt16()
	r.Count100 = ms.ReadInt16()
	r.Count50 = ms.ReadInt16()
	r.CountGeki = ms.ReadInt16()
	r.CountKatu = ms.ReadInt16()
	r.CountMiss = ms.ReadInt16()
	r.Score = ms.ReadInt32()
	r.Combo = ms.ReadInt16()
	r.FullCombo = (ms.ReadByte() == byte(1))
	r.Mods = ms.ReadInt32()
	
	r.Lifebar = ms.ReadString()
	r.TimeTicks = ms.ReadInt64()
	r.ReplayLength = ms.ReadInt32()
	r.ReplayData = f[ms.position:len(f)-8]
	ms.position += len(r.ReplayData)
	r.ReplayId = ms.ReadInt64()
	return r, nil
}