package manifest

import "strings"

func (m *Manifest) Generate(input map[string]string) (string, error) {
	output := make([]string, 0, len(input))

	for k, v := range input {
		output = append(output, k+"  "+v)
	}

	return strings.Join(output, "\n"), nil
}
