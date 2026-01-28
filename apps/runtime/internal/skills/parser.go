package skills

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	ErrMissingFrontmatter = errors.New("missing yaml frontmatter")
	ErrInvalidFrontmatter = errors.New("invalid yaml frontmatter")
)

func parseFrontmatterOnly(r io.Reader) (Frontmatter, error) {
	br := bufio.NewReader(r)
	firstLine, err := br.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return Frontmatter{}, err
	}
	if strings.TrimSpace(firstLine) != "---" {
		return Frontmatter{}, ErrMissingFrontmatter
	}

	var yamlBuf bytes.Buffer
	for {
		line, readErr := br.ReadString('\n')
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return Frontmatter{}, readErr
		}
		if strings.TrimSpace(line) == "---" {
			break
		}
		yamlBuf.WriteString(line)
		if errors.Is(readErr, io.EOF) {
			return Frontmatter{}, ErrInvalidFrontmatter
		}
	}

	fm := Frontmatter{}
	if err := yaml.Unmarshal(yamlBuf.Bytes(), &fm); err != nil {
		return Frontmatter{}, err
	}
	fm.Name = strings.TrimSpace(fm.Name)
	fm.Description = strings.TrimSpace(fm.Description)
	if fm.Name == "" {
		return Frontmatter{}, ErrInvalidFrontmatter
	}
	return fm, nil
}

func parseSkillFile(data []byte) (Frontmatter, string, error) {
	data = bytes.TrimPrefix(data, []byte("\ufeff"))
	s := string(data)
	if !strings.HasPrefix(s, "---") {
		return Frontmatter{}, "", ErrMissingFrontmatter
	}

	lines := strings.Split(s, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return Frontmatter{}, "", ErrMissingFrontmatter
	}

	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return Frontmatter{}, "", ErrInvalidFrontmatter
	}

	yamlPart := strings.Join(lines[1:end], "\n")
	body := strings.Join(lines[end+1:], "\n")

	fm := Frontmatter{}
	if err := yaml.Unmarshal([]byte(yamlPart), &fm); err != nil {
		return Frontmatter{}, "", err
	}
	fm.Name = strings.TrimSpace(fm.Name)
	fm.Description = strings.TrimSpace(fm.Description)
	if fm.Name == "" {
		return Frontmatter{}, "", ErrInvalidFrontmatter
	}
	return fm, strings.TrimLeft(body, "\r\n"), nil
}
