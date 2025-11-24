// Package platform provides an abstraction layer over OS-level operations
// to enable better testing and alternative implementations without changing code.
//
// # Overview
//
// The platform package wraps common OS operations behind interfaces, making it easy to:
//   - Write testable code that depends on OS operations
//   - Mock OS behavior in unit tests
//   - Swap implementations without changing business logic
//   - Create alternative implementations for different environments
//
// # Quick Start
//
// Using the real OS platform in production:
//
//	p := platform.NewOSPlatform()
//	value := p.Env().Getenv("MY_VAR")
//	data, err := p.File().ReadFile("/path/to/file")
//
// Using mock platform in tests:
//
//	mock := platform.NewMockPlatform()
//	mock.Env().Setenv("MY_VAR", "test_value")
//	mock.File().WriteFile("/test/file", []byte("content"), 0644)
//
// # Dependency Injection Pattern
//
// The recommended pattern is to inject the Platform interface into your components:
//
//	type MyService struct {
//	    platform platform.Platform
//	}
//
//	func NewMyService(p platform.Platform) *MyService {
//	    return &MyService{platform: p}
//	}
//
//	func (s *MyService) DoWork() error {
//	    config := s.platform.Env().Getenv("CONFIG")
//	    data, err := s.platform.File().ReadFile(config)
//	    // ... use data
//	    return nil
//	}
//
// Then in production use NewOSPlatform() and in tests use NewMockPlatform():
//
//	// Production
//	service := NewMyService(platform.NewOSPlatform())
//
//	// Testing
//	func TestMyService(t *testing.T) {
//	    mock := platform.NewMockPlatform()
//	    mock.Env().Setenv("CONFIG", "/test/config.json")
//	    mock.File().WriteFile("/test/config.json", []byte(`{}`), 0644)
//
//	    service := NewMyService(mock)
//	    err := service.DoWork()
//	    // ... assertions
//	}
//
// # Subsystems
//
// The platform provides access to five main subsystems:
//
//   - Signal: OS signal handling (SIGINT, SIGTERM, etc.)
//   - Env: Environment variable operations
//   - File: File system operations
//   - Process: Process-related operations (PID, hostname, etc.)
//   - Clock: Time operations for testing time-dependent code
//
// Each subsystem can be accessed through the Platform interface or injected independently.
//
// # Testing Time-Dependent Code
//
// The MockClock allows you to control time in tests:
//
//	mock := platform.NewMockPlatform()
//	clock := mock.Clock().(*platform.MockClock)
//
//	// Set specific time
//	clock.Set(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
//
//	// Run code that depends on time
//	result := MyTimeDependentFunction(mock)
//
//	// Advance time
//	clock.Advance(24 * time.Hour)
//
//	// Verify behavior after time passes
//	result2 := MyTimeDependentFunction(mock)
//
// # Signal Handling
//
// The SignalHandler allows you to test signal handling without sending actual OS signals:
//
//	mock := platform.NewMockPlatform()
//	handler := mock.Signal().(*platform.MockSignalHandler)
//
//	sigChan := make(chan os.Signal, 1)
//	handler.Notify(sigChan, os.Interrupt)
//
//	// Simulate signal
//	handler.SendSignal(os.Interrupt)
//
//	// Verify signal received
//	sig := <-sigChan
//
// # Best Practices
//
//   - Always use dependency injection to pass Platform instances
//   - Use NewOSPlatform() in production code
//   - Use NewMockPlatform() in all unit tests
//   - Avoid direct calls to os, signal, time packages in business logic
//   - If you only need one subsystem, inject that specific handler instead of the entire Platform
//
// For more examples and detailed documentation, see the README.md file.
package platform
