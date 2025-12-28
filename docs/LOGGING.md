# Logging System

## Purpose

The ztiaws toolkit implements a centralized logging system to provide consistent, configurable logging across all scripts. This ensures uniform output formatting, optional file persistence, and maintainable logging behavior.

## Design Principles

- **Centralized**: All logging functions are defined once in `src/00_utils.sh`
- **Configurable**: Scripts can enable/disable file logging based on needs
- **Consistent**: Uniform message formatting and color coding across tools
- **Non-intrusive**: Console-only mode for interactive tools, file logging for automation

## Logging Levels

- `log_info` - Informational messages (green)
- `log_warn` - Warning messages (yellow)
- `log_error` - Error messages (red)
- `log_debug` - Debug messages (cyan)

## Script Behavior

### authaws

- **Purpose**: Authentication workflows require audit trails
- **Logging**: File logging enabled by default
- **Location**: `$HOME/logs/authaws-YYYY-MM-DD.log`

### ssm

- **Purpose**: Interactive session management prioritizes clean console output
- **Logging**: Console-only by default
- **Override**: `ENABLE_SSM_LOGGING=true` enables file logging when needed

## Configuration

### Environment Variables

- `LOG_DIR` - Override default log directory (`$HOME/logs`)
- `ENABLE_SSM_LOGGING` - Enable file logging for SSM operations

### File Format

- Timestamped entries: `YYYY-MM-DD HH:MM:SS - [LEVEL] message`
- Execution boundaries marked with session headers
- Daily rotation by filename
