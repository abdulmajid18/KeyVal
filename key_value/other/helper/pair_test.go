package helper

import "testing"

func TestValidate(t *testing.T) {
	key1 := "1234567890123456789012345678901"
	value := "jffju1rig13rg3r2yg3g3r1t78r3t37t3r273r27f73r2g73r2ff3rf3r27" +
		"f3rasnjjnasdnjadsjndsanjsdanjdsanjdsjasndsjdjnsdajnsdajnsjndsajn" +
		"jasnadsaasdkhadhyy727t22effawfawhhdahsahgasdhgdsahgsadhgsadhghdghads" +
		"asdjajsddsjajsdajsadjndsjn"

	pair := NewPair(key1, value)

	err := pair.Validate()

	if err != nil {
		t.Errorf("Should throw an error")
	}

	err = NewPair(key1, "ss").Validate()
	if err != nil {
		t.Errorf("Shoudl throw error as key is longer than 30")
	}

	NewPair("smallKEY", value).Validate()
	if err != nil {
		t.Errorf("Shoudl throw error as value is longer than 90")
	}

}
