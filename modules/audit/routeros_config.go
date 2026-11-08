package audit

import "strings"

// RouterOSObject is one /queue type or /queue tree entry, normalized so the
// generated script and a RouterOS export compare in the same shape. The
// script names the path on every command; an export prints a /queue type
// header once and then bare `add` lines. Both parse to this type.
type RouterOSObject struct {
	Section string
	Name    string
	Props   map[string]string
}

type RouterOSConfig struct {
	Objects []RouterOSObject
}

// Get returns every object with this section and name; more than one means
// the device stacked duplicates, which is a failure, not a detail.
func (c RouterOSConfig) Get(section, name string) []RouterOSObject {
	var out []RouterOSObject
	for _, o := range c.Objects {
		if o.Section == section && o.Name == name {
			out = append(out, o)
		}
	}
	return out
}

// ParseRouterOSConfig reads both dialects and keeps only queue type and
// queue tree objects; removals, prints and comments are ignored.
func ParseRouterOSConfig(text string) RouterOSConfig {
	var out RouterOSConfig
	section := ""
	for _, line := range logicalLines(text) {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var args []string
		if strings.HasPrefix(line, "/") {
			fields := splitProps(line)
			verb := -1
			for i, f := range fields {
				if isCommandVerb(f) {
					verb = i
					break
				}
			}
			if verb < 0 {
				section = strings.TrimPrefix(strings.Join(fields, " "), "/")
				continue
			}
			section = strings.TrimPrefix(strings.Join(fields[:verb], " "), "/")
			if fields[verb] != "add" {
				continue
			}
			args = fields[verb+1:]
		} else {
			fields := splitProps(line)
			if len(fields) == 0 || fields[0] != "add" {
				continue
			}
			args = fields[1:]
		}
		if section != "queue type" && section != "queue tree" {
			continue
		}
		props := map[string]string{}
		for _, a := range args {
			k, v, ok := strings.Cut(a, "=")
			if !ok {
				continue
			}
			props[k] = strings.Trim(v, `"`)
		}
		name := props["name"]
		if name == "" {
			continue
		}
		delete(props, "name")
		out.Objects = append(out.Objects, RouterOSObject{Section: section, Name: name, Props: props})
	}
	return out
}

// logicalLines joins continuation lines (trailing backslash), as RouterOS
// prints and accepts them, then trims each line.
func logicalLines(text string) []string {
	var out []string
	var cur strings.Builder
	for _, raw := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasSuffix(line, `\`) {
			cur.WriteString(strings.TrimSuffix(line, `\`))
			cur.WriteByte(' ')
			continue
		}
		cur.WriteString(line)
		out = append(out, strings.TrimSpace(cur.String()))
		cur.Reset()
	}
	if s := strings.TrimSpace(cur.String()); s != "" {
		out = append(out, s)
	}
	return out
}

func isCommandVerb(s string) bool {
	switch s {
	case "add", "set", "remove", "print", "export", "comment", "unset", "enable", "disable":
		return true
	}
	return false
}

// splitProps splits on spaces but keeps quoted values intact and keeps the
// key=value pair together when the value contains spaces.
func splitProps(line string) []string {
	var out []string
	var cur strings.Builder
	inQuote := false
	for _, r := range line {
		switch {
		case r == '"':
			inQuote = !inQuote
			cur.WriteRune(r)
		case r == ' ' && !inQuote:
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}
