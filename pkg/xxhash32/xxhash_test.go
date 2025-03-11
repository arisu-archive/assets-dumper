package xxhash32_test

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/arisu-archive/assets-dumper/pkg/xxhash32"
)

type TestVector struct {
	Input        string `json:"input"`
	Seed         string `json:"seed"`
	ExpectedHash string `json:"expected_hash"`
}

var _ = Describe("XXHash32", func() {
	Context("when testing with generated test vectors", func() {
		It("should match xxhsum-verified values", func() {
			// Get all JSON test files
			testFiles, globErr := filepath.Glob("../../mocks/testdata/xxhash32/xxhash32_*.json")
			Expect(globErr).NotTo(HaveOccurred(), "Failed to find test vector files")
			Expect(testFiles).NotTo(BeEmpty(), "No test vector files found")

			for _, filename := range testFiles {
				// Use filename as test description to better identify failures
				By("test vector file: " + filename)
				// Load the test vector
				data, readErr := os.ReadFile(filename)
				Expect(readErr).NotTo(HaveOccurred(), "Failed to read test vector file: %s", filename)

				var testVector TestVector
				unmarshalErr := json.Unmarshal(data, &testVector)
				Expect(unmarshalErr).NotTo(HaveOccurred(), "Failed to parse test vector JSON: %s", filename)

				// Parse the seed value
				seedVal, parseErr := strconv.ParseUint(testVector.Seed, 0, 32)
				Expect(parseErr).NotTo(HaveOccurred(), "Invalid seed value in test vector: %s", filename)
				seed := uint32(seedVal)

				// Parse the expected hash
				expectedHash, parseErr := strconv.ParseUint(testVector.ExpectedHash, 16, 32)
				Expect(parseErr).NotTo(HaveOccurred(), "Invalid hash in test vector: %s", filename)

				// Calculate the hash with our implementation
				h := xxhash32.NewWithSeed(seed)
				h.Write([]byte(testVector.Input))
				result := h.Sum32()

				Expect(result).To(Equal(uint32(expectedHash)),
					"Failed for input: %s with seed: %s in file: %s",
					testVector.Input, testVector.Seed, filename)
			}
		})
	})

	Context("when writing incrementally", func() {
		It("should produce the same hash as single write", func() {
			input := "abcdefghijklmnopqrstuvwxyz"
			expected := uint32(0x63A14D5F)

			// Single write
			h1 := xxhash32.New()
			h1.Write([]byte(input))

			// Incremental writes
			h2 := xxhash32.New()
			h2.Write([]byte(input[:10]))
			h2.Write([]byte(input[10:20]))
			h2.Write([]byte(input[20:]))

			Expect(h2.Sum32()).To(Equal(h1.Sum32()))
			Expect(h2.Sum32()).To(Equal(expected))
		})
	})

	Context("when reset is called", func() {
		It("should reset the hash state", func() {
			h := xxhash32.New()
			h.Write([]byte("abcdefghijklmnopqrstuvwxyz"))
			h.Reset()
			h.Write([]byte("abcdefghijklmnopqrstuvwxyz"))
			Expect(h.Sum32()).To(Equal(uint32(0x63A14D5F)))
		})
	})

	Context("when Sum is called", func() {
		It("should append the correct hash to the input", func() {
			prefix := []byte{0xDE, 0xAD, 0xBE, 0xEF}
			expectedHash := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x63, 0xA1, 0x4D, 0x5F}
			h := xxhash32.New()
			h.Write([]byte("abcdefghijklmnopqrstuvwxyz"))
			sum := h.Sum(prefix)
			// Compare the byte arrays
			Expect(hex.EncodeToString(sum)).To(Equal(hex.EncodeToString(expectedHash)))
		})
	})
})
