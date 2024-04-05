package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func getRandomBinarySegment(length int) string {
	segment := ""
	for i := 0; i < length; i++ {
		segment += strconv.Itoa(rand.Intn(2))
	}
	return segment
}

func expandIPv6Segments(segments []string) []string {
	emptySegmentIndex := -1
	for i, segment := range segments {
		if segment == "" {
			emptySegmentIndex = i
			break
		}
	}

	if emptySegmentIndex != -1 {
		emptySegmentCount := 8 - len(segments) + 1
		emptySegments := make([]string, emptySegmentCount)
		for i := 0; i < emptySegmentCount; i++ {
			emptySegments[i] = "0000"
		}
		segments = append(segments[:emptySegmentIndex], append(emptySegments, segments[emptySegmentIndex+1:]...)...)
	}

	return segments
}

func segmentToBinary(segment string) string {
	if len(segment) < 4 {
		segment = fmt.Sprintf("%04s", segment)
	}
	segmentInt, _ := strconv.ParseInt(segment, 16, 16)
	segmentBinary := fmt.Sprintf("%016s", strconv.FormatInt(segmentInt, 2))
	return segmentBinary
}

func convertIPv6ToBinary(ipv6 string) string {
	segments := strings.Split(ipv6, ":")

	expandedSegments := expandIPv6Segments(segments)

	binarySegments := make([]string, len(expandedSegments))
	for i, segment := range expandedSegments {
		binarySegments[i] = segmentToBinary(segment)
	}

	return strings.Join(binarySegments, "")
}

func padBinary(binary string) string {
	paddedLength := (len(binary) + 127) / 128 * 128
	return binary + strings.Repeat("0", paddedLength-len(binary))
}

func splitBinaryIntoSegments(binary string, segmentSize int) []string {
	segments := make([]string, 0)
	for i := 0; i < len(binary); i += segmentSize {
		end := i + segmentSize
		if end > len(binary) {
			end = len(binary)
		}
		segments = append(segments, binary[i:end])
	}
	return segments
}

func convertBinarySegmentToHex(segment string) string {
	segmentInt, _ := strconv.ParseInt(segment, 2, 16)
	hexSegment := strconv.FormatInt(segmentInt, 16)
	if hexSegment == "0" {
		return ""
	}
	return hexSegment
}

func findLongestEmptySequence(segments []string) (int, int) {
	maxEmptyStart := -1
	maxEmptyCount := 0
	emptyStart := -1
	emptyCount := 0

	for i, segment := range segments {
		if segment == "" {
			if emptyCount == 0 {
				emptyStart = i
			}
			emptyCount++
			if emptyCount > maxEmptyCount {
				maxEmptyCount = emptyCount
				maxEmptyStart = emptyStart
			}
		} else {
			emptyCount = 0
		}
	}

	return maxEmptyStart, maxEmptyCount
}

func shortenIPv6Segments(segments []string, maxEmptyStart int, maxEmptyCount int) []string {
	if maxEmptyCount > 1 {
		segments = append(segments[:maxEmptyStart], segments[maxEmptyStart+maxEmptyCount:]...)
	}
	return segments
}

func convertBinaryToIPv6(binary string) string {
	paddedBinary := padBinary(binary)
	segments := splitBinaryIntoSegments(paddedBinary, 16)

	ipv6Segments := make([]string, len(segments))
	for i, segment := range segments {
		ipv6Segments[i] = convertBinarySegmentToHex(segment)
	}

	maxEmptyStart, maxEmptyCount := findLongestEmptySequence(ipv6Segments)
	shortenedSegments := shortenIPv6Segments(ipv6Segments, maxEmptyStart, maxEmptyCount)

	return strings.Join(shortenedSegments, ":")
}

func generateIPv6Commands(input string, ipv6Count int, interfaceName string) []string {
	s := strings.Split(input, "/")
	ipv6WithPrefix := s[0]
	prefixLength, _ := strconv.Atoi(s[1])

	fullBinary := convertIPv6ToBinary(ipv6WithPrefix)
	prefixBinary := fullBinary[:prefixLength]

	generatedIPv6Commands := make([]string, ipv6Count)
	for i := 0; i < ipv6Count; i++ {
		randomBinary := getRandomBinarySegment(128 - prefixLength)
		generatedBinary := prefixBinary + randomBinary
		generatedIPv6 := convertBinaryToIPv6(generatedBinary)

		generatedIPv6Commands[i] = generatedIPv6 + "/" + strconv.Itoa(prefixLength)
	}

	return generatedIPv6Commands
}

func generateShellCommands(ipv6Addresses []string, interfaceName string) string {
	shellCommands := ""

	for _, ipv6Address := range ipv6Addresses {
		ipv6Address = strings.TrimSpace(ipv6Address)
		if ipv6Address != "" {
			shellCommands += fmt.Sprintf("sudo ip addr add %s dev %s;", ipv6Address, interfaceName)
		}
	}

	return shellCommands
}

func main() {
	input := os.Args[1]
	ipv6Count, _ := strconv.Atoi(os.Args[2])
	interfaceName := os.Args[3]

	generatedIPv6Commands := generateIPv6Commands(input, ipv6Count, interfaceName)
	shellCommands := generateShellCommands(generatedIPv6Commands, interfaceName)

	fmt.Println(shellCommands)
	cmd := exec.Command("sh", "-c", shellCommands)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}
