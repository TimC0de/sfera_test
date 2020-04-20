package types

import (
	"bytes"
	"encoding/hex"
	"math"
	"strconv"
	"strings"
)

func enchToI(encH string) string {
	i, err := strconv.ParseInt(encH, 16, 0)
	if err != nil {
		return ""
	}
	return strconv.Itoa(int(i))
}

func htoi(h []byte) string {
	return enchToI(hex.EncodeToString(h))
}

func htos(h []byte) string {
	return string(h)
}

func htof(h []byte) string {
	var signMult float64 = 1
	signPart := strconv.FormatInt(int64(h[0]), 16)
	if signPart[0] > '7' {
		signMult = -1
		signB := signPart[0]
		signB -= '8'
		signPart = strings.Replace(signPart, string([]byte{signPart[0]}), string([]byte{signB}), 1)

		h0int, err := strconv.ParseInt(signPart, 16, 8)
		if err != nil {
			return ""
		}
		h[0] = byte(h0int)
	}

	encH := hex.EncodeToString(h)
	println(encH)
	exponentDec, err := strconv.Atoi(enchToI(encH[:3]))
	if err != nil {
		return ""
	}
	print(exponentDec)

	mantisDec, err := strconv.Atoi(enchToI(encH[3:]))
	if err != nil {
		return ""
	}
	println(" ", mantisDec)

	return strconv.FormatFloat(signMult*(1+float64(mantisDec)/math.Pow(2, 52))*math.Pow(2, float64(exponentDec-1023)), 'f', 2, 64)
}

var parsingMethods map[int]func([]byte) string = map[int]func([]byte) string{
	1: htoi,
	2: htos,
	3: htof,
}

/*
 * API Return types
 */

type Video struct {
	Name    string `json:"name,omitempty"`
	Width   string `json:"width,omitempty"`
	Height  string `json:"height,omitempty"`
	BitRate string `json:"bitRate,omitempty"`
}

type Audio struct {
	Name    string `json:"name,omitempty"`
	BitRate string `json:"bitRate,omitempty"`
}

type Media struct {
	Name      string `json:"name,omitempty"`
	Extension string `json:"extension,omitempty"`
	Duration  string `json:"duration,omitempty"`
	Video     Video  `json:"video,omitempty"`
	Audio     Audio  `json:"audio,omitempty"`
}

/*
 * Parsing process types
 */

type Container struct {
	Name       string   `yaml:"name"`
	Extensions []string `yaml:"extensions"`
}

type Setting struct {
	Name          string `yaml:"name"`
	Offset        int    `yaml:"offset"`
	Length        int    `yaml:"length"`
	WrapperName   string `yaml:"wrapperName"`
	ParsingMethod int    `yaml:"parsingType"`
}

func (s *Setting) Parse(content []byte) string {
	res := ""

	index := 1
	if strings.Contains(s.WrapperName, "#") {
		separatorIndex := strings.Index(s.WrapperName, "#")
		offset, err := strconv.Atoi(s.WrapperName[separatorIndex+1:])
		if err != nil {
			return res
		}
		index += offset
		s.WrapperName = s.WrapperName[:separatorIndex]
	}

	if bytes.Contains(content, []byte(s.WrapperName)) {
		separatedContent := bytes.Split(content, []byte(s.WrapperName))
		if len(separatedContent) <= index {
			return res
		}

		settingContent := separatedContent[index]
		res = parsingMethods[s.ParsingMethod](settingContent[s.Offset : s.Offset+s.Length])
	}
	return res
}

type Config struct {
	Name     string    `yaml:"name"`
	Settings []Setting `yaml:"settings,flow"`
	content  []byte
}

func (c *Config) WithContent(content []byte) *Config {
	c.content = content
	return c
}

func (c *Config) Parse() Media {
	result := Media{}
	settingValues := make(map[string]string)
	for _, setting := range c.Settings {
		settingValues[setting.Name] = setting.Parse(c.content)
	}
	result.Duration = settingValues["duration"]
	result.Video.Name = settingValues["videoCodecID"]
	result.Video.Width = settingValues["width"]
	result.Video.Height = settingValues["height"]
	result.Video.BitRate = settingValues["videoBitRate"]

	result.Audio.Name = settingValues["audioCodecID"]
	result.Audio.BitRate = settingValues["audioBitRate"]

	return result
}
