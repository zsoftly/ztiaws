# Instance State Validation

## Overview

`ztictl` includes comprehensive instance state validation to prevent operations on instances in invalid states. This ensures safe operations and provides clear, actionable feedback when operations cannot proceed.

The validation system was introduced to address [Issue #133](https://github.com/zsoftly/ztiaws/issues/133) where users could attempt operations on instances in inappropriate states (e.g., trying to reboot a terminated instance).

## Architecture

### Centralized Validation

All validation logic is centralized in `cmd/ztictl/ssm_validation.go` following the DRY (Don't Repeat Yourself) principle. This ensures:

- **Single source of truth** for validation rules
- **Consistent error messages** across all commands
- **Easy maintenance** when adding new states or operations
- **Comprehensive testing** in one location

### Core Components

#### 1. `InstanceValidationRequirements` Struct

Defines validation criteria for an operation:

```go
type InstanceValidationRequirements struct {
    AllowedStates    []string  // EC2 states acceptable for operation
    RequireSSMOnline bool      // Whether SSM agent must be online
    Operation        string    // Operation name for error messages
}
```

#### 2. `ValidateInstanceState()` Function

Main entry point for validation:

```go
func ValidateInstanceState(
    ctx context.Context,
    ssmManager *ssm.Manager,
    instanceID string,
    region string,
    requirements InstanceValidationRequirements
) error
```

**Process:**

1. Fetches instance details from AWS
2. Validates EC2 instance state
3. Validates SSM agent status (if required)
4. Returns descriptive error if validation fails

#### 3. Helper Functions

- **`validateEC2State()`** - Checks if instance is in allowed EC2 state
- **`validateSSMStatus()`** - Checks if SSM agent is online
- **`displayStateSuggestion()`** - Provides helpful suggestions based on current state
- **`getInstanceStateColor()`** - Returns colored state representation
- **`getSSMStatusColor()`** - Returns colored SSM status representation

## Operation-Specific Requirements

### SSM Operations (Connect, Exec, Transfer)

**Requirements:**

- Instance State: `running`
- SSM Agent: `Online`

**Example:**

```go
ValidateInstanceState(ctx, ssmManager, instanceID, region, InstanceValidationRequirements{
    AllowedStates:    []string{"running"},
    RequireSSMOnline: true,
    Operation:        "connect",
})
```

### Power Operations - Start

**Requirements:**

- Instance State: `stopped`
- SSM Agent: Not required

**Example:**

```go
ValidateInstanceState(ctx, ssmManager, instanceID, region, InstanceValidationRequirements{
    AllowedStates:    []string{"stopped"},
    RequireSSMOnline: false,
    Operation:        "start",
})
```

### Power Operations - Stop/Reboot

**Requirements:**

- Instance State: `running`
- SSM Agent: Not required

**Example:**

```go
ValidateInstanceState(ctx, ssmManager, instanceID, region, InstanceValidationRequirements{
    AllowedStates:    []string{"running"},
    RequireSSMOnline: false,
    Operation:        "stop",  // or "reboot"
})
```

## EC2 Instance States

| State           | Symbol | Color  | Description               | Can Start | Can Stop | Can Connect |
| --------------- | ------ | ------ | ------------------------- | --------- | -------- | ----------- |
| `running`       | ●      | Green  | Instance is operational   | ❌        | ✅       | ✅          |
| `stopped`       | ○      | Red    | Instance is stopped       | ✅        | ❌       | ❌          |
| `stopping`      | ◑      | Yellow | Instance is stopping      | ❌        | ❌       | ❌          |
| `pending`       | ◐      | Yellow | Instance is starting      | ❌        | ❌       | ❌          |
| `shutting-down` | ◑      | Yellow | Instance is shutting down | ❌        | ❌       | ❌          |
| `terminated`    | ✗      | Red    | Instance is deleted       | ❌        | ❌       | ❌          |

## SSM Agent States

| Status           | Symbol | Color  | Description                        | Can Connect |
| ---------------- | ------ | ------ | ---------------------------------- | ----------- |
| `Online`         | ✓      | Green  | Agent is connected and ready       | ✅          |
| `ConnectionLost` | ⚠      | Yellow | Agent was online but disconnected  | ⚠️ Allowed  |
| `No Agent`       | ✗      | Red    | Agent not installed or not running | ❌          |

**Note:** Operations are allowed with `ConnectionLost` status but may fail. Users receive a warning.

## Error Message Format

When validation fails, users receive structured feedback:

```
✗ Cannot <operation> - Instance is not in required state

Instance Details:
  Instance ID: i-1234567890abcdef0
  Name:        web-server-prod
  State:       ● running
  Required:    [stopped]

💡 Tip: Instance is already running. Use 'reboot' to restart it:
   ztictl ssm reboot i-1234567890abcdef0 --region ca-central-1
```

### Components:

1. **Error Header** - Clear statement of why operation cannot proceed
2. **Instance Details** - Current instance information with colored state
3. **Helpful Suggestion** - Context-aware tip with example command

## Developer Guide

### Adding Validation to New Commands

1. Import the validation types (already in `main` package)
2. Call `ValidateInstanceState()` before performing the operation:

```go
func myNewCommand(regionCode, instanceIdentifier string) error {
    region := resolveRegion(regionCode)
    ctx := context.Background()
    ssmManager := ssm.NewManager(logger)

    // Get instance ID (may use fuzzy finder)
    instanceID, err := ssmManager.GetInstanceService().SelectInstanceWithFallback(
        ctx,
        instanceIdentifier,
        region,
        nil,
    )
    if err != nil {
        return fmt.Errorf("instance selection failed: %w", err)
    }

    // Validate instance state
    if err := ValidateInstanceState(ctx, ssmManager, instanceID, region, InstanceValidationRequirements{
        AllowedStates:    []string{"running"},
        RequireSSMOnline: true,
        Operation:        "my-operation",
    }); err != nil {
        return err
    }

    // Proceed with operation...
    return performMyOperation(instanceID, region)
}
```

### Extending State Suggestions

To add new state-specific suggestions, modify `displayStateSuggestion()` in `ssm_validation.go`:

```go
func displayStateSuggestion(state, region, instanceID string, requirements InstanceValidationRequirements) {
    switch state {
    case "my-new-state":
        colors.PrintData("💡 Tip: Do this to proceed:\n")
        fmt.Printf("   ztictl ssm <command> %s --region %s\n", instanceID, region)
    // ... existing cases
    }
}
```

### Testing Validation

Tests for validation are in `cmd/ztictl/ssm_power_test.go` and related test files. Key test scenarios:

- **Happy path** - Validation passes for correct states
- **Invalid states** - Validation fails with proper error messages
- **SSM agent offline** - Appropriate warnings/errors for agent status
- **Edge cases** - Terminated instances, stopping instances, etc.

## Benefits of Centralized Validation

### Before (Duplicated Code)

- `validateInstanceForSSM()` in `ssm_connect.go` (~75 lines)
- `validateInstanceForSSM()` in `ssm_exec.go` (~75 lines, duplicate)
- `validateInstanceForPowerOperation()` in `ssm_power.go` (~85 lines)
- **Total: ~235 lines across 3 files**

### After (Centralized)

- `ssm_validation.go` - Single file with all validation logic (~203 lines)
- Called from all commands with simple, configurable requirements
- **Savings: ~32 lines + improved maintainability**

### Maintainability Improvements

1. **Single Update Point** - Add new state? Update one place.
2. **Consistent Behavior** - All commands validate the same way.
3. **Testable** - One comprehensive test suite covers all scenarios.
4. **Discoverable** - New developers find validation logic in one file.
5. **Extensible** - Easy to add new operations or validation rules.

## Best Practices

### 1. Always Validate Before Operations

**Do:**

```go
// Validate first
if err := ValidateInstanceState(...); err != nil {
    return err
}
// Then operate
result, err := performOperation(...)
```

**Don't:**

```go
// Operate without validation - may fail with cryptic AWS errors
result, err := performOperation(...)
```

### 2. Use Descriptive Operation Names

**Do:**

```go
Operation: "execute commands"  // Clear and user-friendly
```

**Don't:**

```go
Operation: "exec"  // Too terse for error messages
```

### 3. Specify Exact Required States

**Do:**

```go
AllowedStates: []string{"stopped"}  // Explicit
```

**Don't:**

```go
AllowedStates: []string{"stopped", "stopping"}  // Ambiguous
```

## Related Documentation

- [Fuzzy Finder Features](FUZZY_FINDER_FEATURES.md) - Interactive instance selection
- [Commands Reference](../../docs/COMMANDS.md) - All available commands
- [Troubleshooting](../../docs/TROUBLESHOOTING.md) - Common issues and solutions

## Changelog

- **v2.1.0** - Initial validation system introduced ([Issue #133](https://github.com/zsoftly/ztiaws/issues/133))
- **v2.1.1** - Centralized validation logic following DRY principles
- **v2.1.2** - Added comprehensive state suggestions and colored output
