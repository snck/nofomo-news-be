package llm

import "testing"

func TestCleanJSONResponse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain JSON unchanged",
			input: `{"headline":"test"}`,
			want:  `{"headline":"test"}`,
		},
		{
			name:  "strips json fenced block",
			input: "```json\n{\"headline\":\"test\"}\n```",
			want:  `{"headline":"test"}`,
		},
		{
			name:  "strips plain fenced block",
			input: "```\n{\"headline\":\"test\"}\n```",
			want:  `{"headline":"test"}`,
		},
		{
			name:  "trims surrounding whitespace",
			input: "  {\"headline\":\"test\"}  ",
			want:  `{"headline":"test"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanJSONResponse(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
