package cryptograph

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {

	It("Compare password: Two Matching passwords", func() {
		passwordOneRaw := "PasswordOne123"
		passwordTwoRaw := "PasswordOne123"

		salt := GenerateRandomSalt(32)

		hashedPasswordOne, encodedSalt := HashPassword(&passwordOneRaw, &salt)

		Expect(ComparePassword(&passwordTwoRaw, &hashedPasswordOne, &encodedSalt)).To((BeTrue()), "Passwords should match")
	})

	It("Compare password: Two Non-Matching passwords", func() {
		passwordOneRaw := "PasswordOne123"
		passwordTwoRaw := "PasswordOne12"

		salt := GenerateRandomSalt(32)

		hashedPasswordOne, encodedSalt := HashPassword(&passwordOneRaw, &salt)

		Expect(ComparePassword(&passwordTwoRaw, &hashedPasswordOne, &encodedSalt)).To((BeFalse()), "Passwords should not match")
	})

})
