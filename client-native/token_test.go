package main

import "testing"

func TestGenerateToken(t *testing.T) {
	const want = "Z3PQGRM08FQBbbqCq8D2MQcZ91PPHH8ccDnFXBGEfc/s08hbSu8P4D8BCfViRIKrNMUDB65ePdeZ5m3WyccLjkXArhEZV1pC+YM/w5+ynDDzjdL+23TH79q3taU9WB6iY3GzwbAHhint8YBVyIQn5B8Ccm09odcohfVVPPBZShfFLDUFnMvAcY+uZ4Cla6j5uNM5SDUjqhhcAql9X9vJxSRYsv54sW9qa0OdSaJasxd83QWu3RNFyaWcvb9m7KYpEUrT7wmT/fN8g4tTg4lEJCqn+vKQ/OWVCv0Dly8tcvI4manqObBJts4yoYfiUt7RSTgclLT5FP5+pQsMXcs/"
	got, err := generateToken(1)
	if err != nil {
		t.Fatalf("generateToken: %v", err)
	}
	if got != want {
		t.Fatalf("token mismatch\n got: %s\nwant: %s", got, want)
	}
}

func TestMD5Password(t *testing.T) {
	if got, want := md5Password("abc"), "900150983cd24fb0d6963f7d28e17f72"; got != want {
		t.Fatalf("md5Password(abc) = %s, want %s", got, want)
	}
}
