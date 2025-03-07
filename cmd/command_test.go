package cmd_test

import (
	"bytes"
	"errors"
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/arisu-archive/assets-dumper/cmd"
)

var _ = Describe("Command", func() {
	var (
		testCmd    *cobra.Command
		logBuffer  *bytes.Buffer
		testLogger *slog.Logger
		origLogger *slog.Logger
	)

	BeforeEach(func() {
		testCmd = &cobra.Command{
			Use: "test",
		}

		// Set up log capture
		logBuffer = &bytes.Buffer{}
		testLogger = slog.New(slog.NewJSONHandler(logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		origLogger = slog.Default()
		slog.SetDefault(testLogger)
	})

	AfterEach(func() {
		// Restore original logger
		slog.SetDefault(origLogger)
	})

	Describe("RunE", func() {
		Context("when the wrapped function succeeds", func() {
			It("should execute successfully and log completion", func() {
				successFn := func(*cobra.Command, []string) error {
					return nil
				}

				wrappedFn := cmd.RunE("test-command", successFn)
				err := wrappedFn(testCmd, []string{})

				Expect(err).NotTo(HaveOccurred())

				logOutput := logBuffer.String()
				Expect(logOutput).To(ContainSubstring(`"msg":"test-command completed after`))
			})
		})

		Context("when the wrapped function fails", func() {
			It("should return an error with timing information", func() {
				expectedErr := errors.New("test error")
				failFn := func(*cobra.Command, []string) error {
					return expectedErr
				}

				wrappedFn := cmd.RunE("test-command", failFn)
				err := wrappedFn(testCmd, []string{})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("test-command failed after"))
				Expect(err.Error()).To(ContainSubstring("test error"))
			})

			It("should handle nil command gracefully", func() {
				failFn := func(*cobra.Command, []string) error {
					return errors.New("test error")
				}

				wrappedFn := cmd.RunE("test-command", failFn)
				err := wrappedFn(nil, []string{})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("test-command failed after"))
			})
		})

		Context("when checking execution time", func() {
			It("should record the correct duration", func() {
				startTime := time.Now()
				sleepDuration := 100 * time.Millisecond

				slowFn := func(*cobra.Command, []string) error {
					time.Sleep(sleepDuration)
					return nil
				}

				wrappedFn := cmd.RunE("slow-command", slowFn)
				err := wrappedFn(testCmd, []string{})

				executionTime := time.Since(startTime)
				Expect(err).NotTo(HaveOccurred())
				Expect(executionTime).To(BeNumerically(">=", sleepDuration))

				logOutput := logBuffer.String()
				Expect(logOutput).To(ContainSubstring(`"msg":"slow-command completed after`))
			})

			It("should handle very fast executions", func() {
				fastFn := func(*cobra.Command, []string) error {
					return nil
				}

				wrappedFn := cmd.RunE("fast-command", fastFn)
				err := wrappedFn(testCmd, []string{})

				Expect(err).NotTo(HaveOccurred())
				Expect(logBuffer.String()).To(ContainSubstring(`"msg":"fast-command completed after`))
			})
		})

		Context("when handling panics", func() {
			It("should recover from panics and return an error", func() {
				panicFn := func(*cobra.Command, []string) error {
					panic("unexpected panic")
				}

				wrappedFn := cmd.RunE("panic-command", panicFn)
				err := wrappedFn(testCmd, []string{})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("panic-command failed after"))
				Expect(err.Error()).To(ContainSubstring("unexpected panic"))
			})
		})

		Context("when handling command arguments", func() {
			It("should pass arguments correctly", func() {
				var capturedArgs []string
				argCheckFn := func(_ *cobra.Command, args []string) error {
					capturedArgs = args
					return nil
				}

				expectedArgs := []string{"arg1", "arg2"}
				wrappedFn := cmd.RunE("arg-command", argCheckFn)
				err := wrappedFn(testCmd, expectedArgs)

				Expect(err).NotTo(HaveOccurred())
				Expect(capturedArgs).To(Equal(expectedArgs))

				logOutput := logBuffer.String()
				Expect(logOutput).To(ContainSubstring(`"msg":"arg-command completed after`))
			})
		})
	})
})
