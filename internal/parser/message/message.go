package message

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	// https://regex101.com/r/qI1XzO/1
	regexBase = regexp.MustCompile(`^\$(?P<who>[a-zA-Z]+?)(_(?P<type>[a-zA-Z_]+?))?_?(?P<variant>\d+?)?(?P<parameters>:#.+)?;$`)

	// https://regex101.com/r/P26ztJ/1
	regexParameters = regexp.MustCompile(`:#(?P<key>[a-zA-Z0-9]+)=(?P<value>[$_a-zA-Z0-9 .+\-]+)`)
)

type DecodedMessage struct {
	What       string
	SubWhat    string
	Variant    int
	Parameters map[string]string
}

// OriginalIdentifier returns "What" or "What_SubWhat" reflecting original format without variant integer suffix.
func (m *DecodedMessage) OriginalIdentifier() string {
	if m.SubWhat == "" {
		return m.What
	}
	return fmt.Sprintf("%s_%s", m.What, m.SubWhat)
}

// DecodeMessage parses elite-style messages into more usable structure, for examples see message_test.go
func DecodeMessage(msg string) DecodedMessage {
	match := regexBase.FindAllStringSubmatch(msg, -1)

	if len(match) != 1 {
		return DecodedMessage{}
	}

	mVariant, _ := strconv.Atoi(match[0][4])
	mParameters := match[0][5]

	var parameters = make(map[string]string)

	match2 := regexParameters.FindAllStringSubmatch(mParameters, -1)
	for _, value2 := range match2 {
		parameters[value2[1]] = value2[2]
	}

	return DecodedMessage{
		What:       match[0][1],
		SubWhat:    match[0][3],
		Variant:    mVariant,
		Parameters: parameters,
	}
}
