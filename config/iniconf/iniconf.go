package iniconf

import (
	"bufio"
	"errors"
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
func ParseFile(fileName string) (ConfigMap, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ParseReader(file)
}

// ParseReader reads a configuration from an io.Reader.
func ParseReader(r io.Reader) (ConfigMap, error) {
	cfg := ConfigMap{}
	buf := bufio.NewReader(r)

	var (
		line           string
		longLine       bool
		currentSection string
		err            error
	)

	for {
		line, longLine, err = readConfigLine(buf, line, longLine)
		if errors.Is(err, io.EOF) {
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

	return cfg, err
}

// readConfigLine reads and assembles a complete configuration line, handling long lines.
func readConfigLine(buf *bufio.Reader, currentLine string, longLine bool) (string, bool, error) {
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

	return currentSection, errors.New("invalid config file")
}

// handleSectionLine processes a section header line.
func handleSectionLine(cfg ConfigMap, line string) (string, error) {
	section := configSection.ReplaceAllString(line, "$1")
	if section == "" {
		return "", errors.New("invalid structure in file")
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
func (c ConfigMap) ListSections() []string {
	sections := make([]string, 0, len(c))
	for section := range c {
		sections = append(sections, section)
	}
	return sections
}

// WriteFile writes out the configuration to a file.
func (c ConfigMap) WriteFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, section := range c.ListSections() {
		sName := fmt.Sprintf("[ %s ]\n", section)
		if _, err = file.WriteString(sName); err != nil {
			return err
		}

		for k, v := range c[section] {
			line := fmt.Sprintf("%s = %s\n", k, v)
			if _, err = file.WriteString(line); err != nil {
				return err
			}
		}
		if _, err = file.Write([]byte{0x0a}); err != nil {
			return err
		}
	}
	return nil
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
func (c ConfigMap) GetValue(section, key string) (string, bool) {
	if c == nil {
		return "", false
	}

	if section == "" {
		section = DefaultSection
	}

	if _, ok := c[section]; !ok {
		return "", false
	}

	val, present := c[section][key]
	return val, present
}

// GetValueDefault retrieves the value from a key map if present,
// otherwise the default value.
func (c ConfigMap) GetValueDefault(section, key, value string) string {
	kval, ok := c.GetValue(section, key)
	if !ok {
		return value
	}
	return kval
}

// SectionKeys returns the sections in the config map.
func (c ConfigMap) SectionKeys(section string) ([]string, bool) {
	if c == nil {
		return nil, false
	}

	if section == "" {
		section = DefaultSection
	}

	s, ok := c[section]
	if !ok {
		return nil, false
	}

	keys := make([]string, 0, len(s))
	for key := range s {
		keys = append(keys, key)
	}

	return keys, true
}
