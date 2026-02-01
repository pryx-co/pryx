package discord

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestMarshalJSON_Success(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
		want string
	}{
		{
			name: "simple string",
			data: "hello",
			want: `"hello"`,
		},
		{
			name: "integer",
			data: 42,
			want: `42`,
		},
		{
			name: "struct",
			data: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{Name: "test", Age: 25},
			want: `{"name":"test","age":25}`,
		},
		{
			name: "map",
			data: map[string]string{"key": "value"},
			want: `{"key":"value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := marshalJSON(tt.data)
			if err != nil {
				t.Fatalf("marshalJSON() error = %v", err)
			}

			if string(got) != tt.want {
				t.Errorf("marshalJSON() = %s, want %s", string(got), tt.want)
			}
		})
	}
}

func TestMarshalJSON_Error(t *testing.T) {
	// Test with data that can't be marshaled
	data := make(chan int) // channels can't be marshaled

	_, err := marshalJSON(data)
	if err == nil {
		t.Error("marshalJSON() should return error for unmarsalable data")
	}

	if !errors.Is(err, errors.New("failed to marshal JSON")) {
		// Error should contain the failure reason
		if err.Error() == "" {
			t.Error("Error message should not be empty")
		}
	}
}

func TestMarshalJSON_IdentifyData(t *testing.T) {
	identify := IdentifyData{
		Token: "test-token",
		Properties: Properties{
			OS:      "linux",
			Browser: "Pryx",
			Device:  "Pryx",
		},
		Intents: 12345,
	}

	data, err := marshalJSON(identify)
	if err != nil {
		t.Fatalf("marshalJSON() error = %v", err)
	}

	// Verify it can be unmarshaled
	var result IdentifyData
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result.Token != identify.Token {
		t.Errorf("Token mismatch: got %s, want %s", result.Token, identify.Token)
	}

	if result.Properties.OS != identify.Properties.OS {
		t.Errorf("OS mismatch: got %s, want %s", result.Properties.OS, identify.Properties.OS)
	}
}

func TestMarshalJSON_ResumeData(t *testing.T) {
	resume := ResumeData{
		Token:     "test-token",
		SessionID: "test-session",
		Seq:       42,
	}

	data, err := marshalJSON(resume)
	if err != nil {
		t.Fatalf("marshalJSON() error = %v", err)
	}

	// Verify it can be unmarshaled
	var result ResumeData
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result.Token != resume.Token {
		t.Errorf("Token mismatch: got %s, want %s", result.Token, resume.Token)
	}

	if result.SessionID != resume.SessionID {
		t.Errorf("SessionID mismatch: got %s, want %s", result.SessionID, resume.SessionID)
	}

	if result.Seq != resume.Seq {
		t.Errorf("Seq mismatch: got %d, want %d", result.Seq, resume.Seq)
	}
}

func TestMarshalJSON_Nil(t *testing.T) {
	data, err := marshalJSON(nil)
	if err != nil {
		t.Fatalf("marshalJSON(nil) error = %v", err)
	}

	if string(data) != "null" {
		t.Errorf("marshalJSON(nil) = %s, want null", string(data))
	}
}

func TestMarshalJSON_EmptyStruct(t *testing.T) {
	empty := struct{}{}
	data, err := marshalJSON(empty)
	if err != nil {
		t.Fatalf("marshalJSON() error = %v", err)
	}

	if string(data) != "{}" {
		t.Errorf("marshalJSON(empty struct) = %s, want {}", string(data))
	}
}
