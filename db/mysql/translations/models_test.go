package translations

import (
	"testing"
)

func TestTranslations_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected map[string]string
		wantErr  bool
	}{
		{
			name:     "Valid JSON",
			input:    []byte(`{"en-US": "Hello", "tr-TR": "Merhaba"}`),
			expected: map[string]string{"en-US": "Hello", "tr-TR": "Merhaba"},
			wantErr:  false,
		},
		{
			name:     "Plain string",
			input:    []byte("Germany"),
			expected: map[string]string{"en-US": "Germany"},
			wantErr:  false,
		},
		{
			name:     "Empty bytes",
			input:    []byte(""),
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "Nil value",
			input:    nil,
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "Invalid type",
			input:    "string instead of bytes",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Translations{}
			err := tr.Scan(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Translations.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.expected == nil {
					if tr.data != nil {
						t.Errorf("Expected nil data, got %v", tr.data)
					}
				} else {
					if tr.data == nil {
						t.Errorf("Expected data %v, got nil", tt.expected)
						return
					}
					for key, expectedValue := range tt.expected {
						if actualValue := tr.Get(key); actualValue != expectedValue {
							t.Errorf("Expected %s = %s, got %s", key, expectedValue, actualValue)
						}
					}
				}
			}
		})
	}
}

func TestTranslations_ScanAndGet(t *testing.T) {
	// Test scanning a plain string and getting it back
	tr := &Translations{}
	err := tr.Scan([]byte("Germany"))
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should be stored under DEFAULT_LANG_CODE
	value := tr.Get(DEFAULT_LANG_CODE)
	if value != "Germany" {
		t.Errorf("Expected 'Germany', got '%s'", value)
	}

	// Test scanning JSON and getting specific languages
	tr2 := &Translations{}
	err = tr2.Scan([]byte(`{"en-US": "Hello", "tr-TR": "Merhaba"}`))
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if tr2.Get("en-US") != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", tr2.Get("en-US"))
	}
	if tr2.Get("tr-TR") != "Merhaba" {
		t.Errorf("Expected 'Merhaba', got '%s'", tr2.Get("tr-TR"))
	}
}
