# GitHub Copilot Language Server Configuration Reference

## Configuration File Locations

SuperOpenCode looks for configuration in the following locations (in order of precedence):

1. `./opencode.json` (project-specific)
2. `~/.config/opencode/config.json` (user-specific)
3. `~/.opencode.json` (legacy location)
4. Environment variables with `OPENCODE_` prefix

## Complete Configuration Schema

```json
{
  "copilot": {
    "enable_copilot": false,
    "server_path": "auto",
    "node_path": "node",
    "use_native_binary": true,
    "replace_gopls": false,
    "auth_token": "",
    "chat_enabled": true,
    "completion_enabled": true,
    "auto_install": true,
    "server_args": [],
    "environment": {},
    "timeout": 30000,
    "retry_attempts": 3,
    "fallback_to_gopls": true,
    "log_level": "info",
    "cache_enabled": true,
    "cache_size": 100,
    "completion_trigger_length": 1,
    "chat_model": "default",
    "agent_config": {
      "coding_agent": true,
      "debugging_agent": true,
      "documentation_agent": true
    },
    "performance": {
      "max_completion_time": 5000,
      "debounce_delay": 200,
      "max_parallel_requests": 5
    },
    "security": {
      "disable_telemetry": false,
      "private_mode": false,
      "allowed_domains": []
    }
  },
  "lsp": {
    "disable_gopls": false,
    "gopls_path": "gopls",
    "enable_hybrid_mode": false
  }
}
```

## Core Configuration Options

### `enable_copilot`
- **Type**: `boolean`
- **Default**: `false`
- **Description**: Master switch to enable GitHub Copilot language server
- **Example**: 
  ```json
  {
    "copilot": {
      "enable_copilot": true
    }
  }
  ```

### `server_path`
- **Type**: `string`
- **Default**: `"auto"`
- **Options**: 
  - `"auto"`: Automatically detect installation path
  - `"global"`: Use globally installed package
  - `"/path/to/server"`: Explicit path to language server
- **Description**: Path to the Copilot language server executable
- **Examples**:
  ```json
  {
    "copilot": {
      "server_path": "auto"
    }
  }
  ```
  ```json
  {
    "copilot": {
      "server_path": "/usr/local/bin/copilot-language-server"
    }
  }
  ```
  ```json
  {
    "copilot": {
      "server_path": "./node_modules/.bin/copilot-language-server"
    }
  }
  ```

### `node_path`
- **Type**: `string`
- **Default**: `"node"`
- **Description**: Path to Node.js executable (when not using native binary)
- **Example**:
  ```json
  {
    "copilot": {
      "node_path": "/usr/local/bin/node"
    }
  }
  ```

### `use_native_binary`
- **Type**: `boolean`
- **Default**: `true`
- **Description**: Use platform-specific native binary instead of Node.js version
- **Note**: Native binaries are faster but Node.js version provides more debugging info
- **Example**:
  ```json
  {
    "copilot": {
      "use_native_binary": false
    }
  }
  ```

### `replace_gopls`
- **Type**: `boolean`
- **Default**: `false`
- **Description**: Replace gopls entirely instead of running both servers
- **Example**:
  ```json
  {
    "copilot": {
      "replace_gopls": true
    }
  }
  ```

## Authentication Configuration

### `auth_token`
- **Type**: `string`
- **Default**: `""`
- **Description**: GitHub personal access token with Copilot scope
- **Security**: Store in environment variable instead of config file
- **Example**:
  ```json
  {
    "copilot": {
      "auth_token": ""  // Leave empty for interactive auth
    }
  }
  ```

### Environment Variable Authentication
```bash
export GITHUB_TOKEN=ghp_your_token_here
export OPENCODE_COPILOT_AUTH_TOKEN=ghp_your_token_here
```

## Feature Configuration

### `chat_enabled`
- **Type**: `boolean`
- **Default**: `true`
- **Description**: Enable Copilot chat functionality
- **Example**:
  ```json
  {
    "copilot": {
      "chat_enabled": true
    }
  }
  ```

### `completion_enabled`
- **Type**: `boolean`
- **Default**: `true`
- **Description**: Enable AI-powered code completions
- **Example**:
  ```json
  {
    "copilot": {
      "completion_enabled": true
    }
  }
  ```

### `completion_trigger_length`
- **Type**: `integer`
- **Default**: `1`
- **Range**: `1-10`
- **Description**: Minimum characters typed before triggering completions
- **Example**:
  ```json
  {
    "copilot": {
      "completion_trigger_length": 3
    }
  }
  ```

## Installation and Startup Configuration

### `auto_install`
- **Type**: `boolean`
- **Default**: `true`
- **Description**: Automatically install Copilot language server if missing
- **Example**:
  ```json
  {
    "copilot": {
      "auto_install": false
    }
  }
  ```

### `server_args`
- **Type**: `array of strings`
- **Default**: `[]`
- **Description**: Additional command-line arguments for the language server
- **Example**:
  ```json
  {
    "copilot": {
      "server_args": ["--stdio", "--verbose"]
    }
  }
  ```

### `environment`
- **Type**: `object`
- **Default**: `{}`
- **Description**: Environment variables to set for the language server process
- **Example**:
  ```json
  {
    "copilot": {
      "environment": {
        "COPILOT_LOG_LEVEL": "debug",
        "COPILOT_NODE_PATH": "/usr/bin/node"
      }
    }
  }
  ```

### `timeout`
- **Type**: `integer`
- **Default**: `30000`
- **Unit**: milliseconds
- **Description**: Timeout for language server startup
- **Example**:
  ```json
  {
    "copilot": {
      "timeout": 60000
    }
  }
  ```

### `retry_attempts`
- **Type**: `integer`
- **Default**: `3`
- **Range**: `1-10`
- **Description**: Number of retry attempts if server fails to start
- **Example**:
  ```json
  {
    "copilot": {
      "retry_attempts": 5
    }
  }
  ```

### `fallback_to_gopls`
- **Type**: `boolean`
- **Default**: `true`
- **Description**: Fall back to gopls if Copilot server fails
- **Example**:
  ```json
  {
    "copilot": {
      "fallback_to_gopls": false
    }
  }
  ```

## Performance Configuration

### `performance`
- **Type**: `object`
- **Description**: Performance tuning options

#### `max_completion_time`
- **Type**: `integer`
- **Default**: `5000`
- **Unit**: milliseconds
- **Description**: Maximum time to wait for completions

#### `debounce_delay`
- **Type**: `integer`
- **Default**: `200`
- **Unit**: milliseconds  
- **Description**: Delay before sending completion requests while typing

#### `max_parallel_requests`
- **Type**: `integer`
- **Default**: `5`
- **Range**: `1-20`
- **Description**: Maximum number of parallel requests to Copilot

**Example**:
```json
{
  "copilot": {
    "performance": {
      "max_completion_time": 3000,
      "debounce_delay": 300,
      "max_parallel_requests": 3
    }
  }
}
```

## Agent Configuration

### `agent_config`
- **Type**: `object`
- **Description**: Configuration for Copilot agents

#### `coding_agent`
- **Type**: `boolean`
- **Default**: `true`
- **Description**: Enable coding assistance agent

#### `debugging_agent`
- **Type**: `boolean`
- **Default**: `true`
- **Description**: Enable debugging assistance agent

#### `documentation_agent`
- **Type**: `boolean`
- **Default**: `true`
- **Description**: Enable documentation generation agent

**Example**:
```json
{
  "copilot": {
    "agent_config": {
      "coding_agent": true,
      "debugging_agent": false,
      "documentation_agent": true
    }
  }
}
```

## Caching Configuration

### `cache_enabled`
- **Type**: `boolean`
- **Default**: `true`
- **Description**: Enable response caching for better performance

### `cache_size`
- **Type**: `integer`
- **Default**: `100`
- **Range**: `10-1000`
- **Description**: Maximum number of cached responses

**Example**:
```json
{
  "copilot": {
    "cache_enabled": true,
    "cache_size": 200
  }
}
```

## Security Configuration

### `security`
- **Type**: `object`
- **Description**: Security and privacy options

#### `disable_telemetry`
- **Type**: `boolean`
- **Default**: `false`
- **Description**: Disable telemetry data collection

#### `private_mode`
- **Type**: `boolean`
- **Default**: `false`
- **Description**: Enable private mode (no data sent to GitHub)

#### `allowed_domains`
- **Type**: `array of strings`
- **Default**: `[]`
- **Description**: Restrict network access to specific domains

**Example**:
```json
{
  "copilot": {
    "security": {
      "disable_telemetry": true,
      "private_mode": false,
      "allowed_domains": ["api.github.com", "copilot.github.com"]
    }
  }
}
```

## Logging Configuration

### `log_level`
- **Type**: `string`
- **Default**: `"info"`
- **Options**: `"debug"`, `"info"`, `"warn"`, `"error"`
- **Description**: Logging verbosity level
- **Example**:
  ```json
  {
    "copilot": {
      "log_level": "debug"
    }
  }
  ```

## LSP Integration Configuration

### `lsp.disable_gopls`
- **Type**: `boolean`
- **Default**: `false`
- **Description**: Completely disable gopls (use with `copilot.replace_gopls`)
- **Example**:
  ```json
  {
    "lsp": {
      "disable_gopls": true
    }
  }
  ```

### `lsp.gopls_path`
- **Type**: `string`
- **Default**: `"gopls"`
- **Description**: Path to gopls executable
- **Example**:
  ```json
  {
    "lsp": {
      "gopls_path": "/usr/local/bin/gopls"
    }
  }
  ```

### `lsp.enable_hybrid_mode`
- **Type**: `boolean`
- **Default**: `false`
- **Description**: Run both gopls and Copilot simultaneously
- **Example**:
  ```json
  {
    "lsp": {
      "enable_hybrid_mode": true
    }
  }
  ```

## Configuration Profiles

### Minimal Configuration
```json
{
  "copilot": {
    "enable_copilot": true
  }
}
```

### Development Configuration
```json
{
  "copilot": {
    "enable_copilot": true,
    "log_level": "debug",
    "use_native_binary": false,
    "fallback_to_gopls": true,
    "performance": {
      "max_completion_time": 10000
    }
  }
}
```

### Production Configuration
```json
{
  "copilot": {
    "enable_copilot": true,
    "replace_gopls": true,
    "use_native_binary": true,
    "cache_enabled": true,
    "performance": {
      "max_completion_time": 2000,
      "debounce_delay": 150
    },
    "security": {
      "disable_telemetry": true
    }
  },
  "lsp": {
    "disable_gopls": true
  }
}
```

### High-Performance Configuration
```json
{
  "copilot": {
    "enable_copilot": true,
    "use_native_binary": true,
    "cache_enabled": true,
    "cache_size": 500,
    "performance": {
      "max_completion_time": 1000,
      "debounce_delay": 100,
      "max_parallel_requests": 10
    }
  }
}
```

## Environment Variables

All configuration options can be overridden with environment variables using the `OPENCODE_` prefix:

```bash
# Basic options
export OPENCODE_COPILOT_ENABLE_COPILOT=true
export OPENCODE_COPILOT_SERVER_PATH=/path/to/server
export OPENCODE_COPILOT_AUTH_TOKEN=ghp_token

# Nested options use underscores
export OPENCODE_COPILOT_PERFORMANCE_MAX_COMPLETION_TIME=3000
export OPENCODE_COPILOT_SECURITY_DISABLE_TELEMETRY=true
export OPENCODE_LSP_DISABLE_GOPLS=true
```

## Configuration Validation

SuperOpenCode validates configuration on startup and provides helpful error messages:

```bash
# Check configuration
opencode --validate-config

# Show current configuration
opencode --show-config

# Test Copilot configuration
opencode --test-copilot
```

## Migration Configuration

### From gopls to Copilot (Gradual)
```json
{
  "copilot": {
    "enable_copilot": true,
    "replace_gopls": false,
    "fallback_to_gopls": true
  },
  "lsp": {
    "enable_hybrid_mode": true
  }
}
```

### From gopls to Copilot (Complete)
```json
{
  "copilot": {
    "enable_copilot": true,
    "replace_gopls": true
  },
  "lsp": {
    "disable_gopls": true
  }
}
```

## Troubleshooting Configuration

### Debug Configuration Issues
```json
{
  "copilot": {
    "log_level": "debug",
    "timeout": 60000,
    "retry_attempts": 1
  }
}
```

### Test Mode Configuration
```json
{
  "copilot": {
    "enable_copilot": true,
    "server_path": "/explicit/path/to/test/server",
    "fallback_to_gopls": true,
    "performance": {
      "max_completion_time": 30000
    }
  }
}
```

## See Also

- **Installation Guide**: `copilot-installation-guide.md`
- **Migration Guide**: `copilot-migration-guide.md`
- **API Reference**: `copilot-api-reference.md`
- **Troubleshooting**: `copilot-troubleshooting.md`