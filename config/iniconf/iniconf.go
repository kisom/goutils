package iniconf

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
)

// ConfigMap is shorthand for the type used as a config struct.
type ConfigMap map[string]map[string]string

var (
	configSection    = regexp.MustCompile(`^\s*\[\s*(\w+)\s*\]\s*$`)
	quotedConfigLine = regexp.MustCompile(`^\s*(\w+)\s*=\s*["'](.*)["']\s*$`)
	configLine       = regexp.MustCompile(`^\s*(\w+)\s*=\s*(.*)\s*$`)
	commentLine      = regexp.MustCompile(`^#.*$`)
	blankLine        = regexp.MustCompile(`^\s*$`)
)

// DefaultSection is the label for the default ini file section.
var DefaultSection = "default"

// ParseFile attempts to load the named config file.
func ParseFile(fileName string) (cfg ConfigMap, err error) {
	var file *os.File
	file, err = os.Open(fileName)
	if err != nil {
		return
	}
	defer file.Close()
	return ParseReader(file)
}

// ParseReader reads a configuration from an io.Reader.
func ParseReader(r io.Reader) (cfg ConfigMap, err error) {
	cfg = ConfigMap{}
	buf := bufio.NewReader(r)

	var (
		line           string
		longLine       bool
		currentSection string
	)

	for {
		line, longLine, err = readConfigLine(buf, line, longLine)
		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			break
		}

		if line == "" {
			continue
		}

		currentSection, err = processConfigLine(cfg, line, currentSection)
		if err != nil {
			break
		}
	}
	return
}

// readConfigLine reads and assembles a complete configuration line, handling long lines.
func readConfigLine(buf *bufio.Reader, currentLine string, longLine bool) (line string, stillLong bool, err error) {
	lineBytes, isPrefix, err := buf.ReadLine()
	if err != nil {
		return "", false, err
	}

	if isPrefix {
		return currentLine + string(lineBytes), true, nil
	} else if longLine {
		return currentLine + string(lineBytes), false, nil
	}
	return string(lineBytes), false, nil
}

// processConfigLine processes a single line and updates the configuration map.
func processConfigLine(cfg ConfigMap, line string, currentSection string) (string, error) {
	if commentLine.MatchString(line) || blankLine.MatchString(line) {
		return currentSection, nil
	}

	if configSection.MatchString(line) {
		return handleSectionLine(cfg, line)
	}

	if configLine.MatchString(line) {
		return handleConfigLine(cfg, line, currentSection)
	}

	return currentSection, fmt.Errorf("invalid config file")
}

// handleSectionLine processes a section header line.
func handleSectionLine(cfg ConfigMap, line string) (string, error) {
	section := configSection.ReplaceAllString(line, "$1")
	if section == "" {
		return "", fmt.Errorf("invalid structure in file")
	}
	if !cfg.SectionInConfig(section) {
		cfg[section] = make(map[string]string, 0)
	}
	return section, nil
}

// handleConfigLine processes a key=value configuration line.
func handleConfigLine(cfg ConfigMap, line string, currentSection string) (string, error) {
	regex := configLine
	if quotedConfigLine.MatchString(line) {
		regex = quotedConfigLine
	}

	if currentSection == "" {
		currentSection = DefaultSection
		if !cfg.SectionInConfig(currentSection) {
			cfg[currentSection] = map[string]string{}
		}
	}

	key := regex.ReplaceAllString(line, "$1")
	val := regex.ReplaceAllString(line, "$2")
	if key != "" {
		cfg[currentSection][key] = val
	}

	return currentSection, nil
}

// SectionInConfig determines whether a section is in the configuration.
func (c ConfigMap) SectionInConfig(section string) bool {
	_, ok := c[section]
	return ok
}

// ListSections returns the list of sections in the config map.
func (c ConfigMap) ListSections() (sections []string) {
	for section := range c {
		sections = append(sections, section)
	}
	return
}

// WriteFile writes out the configuration to a file.
func (c ConfigMap) WriteFile(filename string) (err error) {
	file, err := os.Create(filename)
	if err != nil {
		return
	}
	defer file.Close()

	for _, section := range c.ListSections() {
		sName := fmt.Sprintf("[ %s ]\n", section)
		_, err = file.Write([]byte(sName))
		if err != nil {
			return
		}

		for k, v := range c[section] {
			line := fmt.Sprintf("%s = %s\n", k, v)
			_, err = file.Write([]byte(line))
			if err != nil {
				return
			}
		}
		_, err = file.Write([]byte{0x0a})
		if err != nil {
			return
		}
	}
	return
}

// AddSection creates a new section in the config map.
func (c ConfigMap) AddSection(section string) {
	if nil != c[section] {
		c[section] = map[string]string{}
	}
}

// AddKeyVal adds a key value pair to a config map.
func (c ConfigMap) AddKeyVal(section, key, val string) {
	if section == "" {
		section = DefaultSection
	}

	if nil == c[section] {
		c.AddSection(section)
	}

	c[section][key] = val
}

// GetValue retrieves the value from a key map.
func (c ConfigMap) GetValue(section, key string) (val string, present bool) {
	if c == nil {
		return
	}

	if section == "" {
		section = DefaultSection
	}

	_, ok := c[section]
	if !ok {
		return
	}

	val, present = c[section][key]
	return
}

// GetValueDefault retrieves the value from a key map if present,
// otherwise the default value.
func (c ConfigMap) GetValueDefault(section, key, value string) (val string) {
	kval, ok := c.GetValue(section, key)
	if !ok {
		return value
	}
	return kval
}

// SectionKeys returns the sections in the config map.
func (c ConfigMap) SectionKeys(section string) (keys []string, present bool) {
	if c == nil {
		return nil, false
	}

	if section == "" {
		section = DefaultSection
	}

	cm := c
	s, ok := cm[section]
	if !ok {
		return nil, false
	}

	keys = make([]string, 0, len(s))
	for key := range s {
		keys = append(keys, key)
	}

	return keys, true
}
