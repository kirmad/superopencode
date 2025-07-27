# GitHub Copilot Language Server Installation Guide

## Prerequisites

### System Requirements
- **Node.js**: Version 20.8 or later
- **npm**: Version 8 or later (typically included with Node.js)
- **Operating System**: Linux, macOS, or Windows
- **GitHub Account**: With Copilot subscription

### Verify Prerequisites
```bash
# Check Node.js version
node --version  # Should be >= 20.8.0

# Check npm version
npm --version   # Should be >= 8.0.0

# Check GitHub Copilot access
gh auth status  # Optional: GitHub CLI for easier auth
```

## Installation Methods

### Method 1: Automatic Installation (Recommended)

The SuperOpenCode application will handle installation automatically when Copilot is enabled.

1. **Enable Copilot in Configuration**:
   ```json
   {
     "copilot": {
       "enable_copilot": true,
       "auto_install": true
     }
   }
   ```

2. **Start SuperOpenCode**:
   ```bash
   opencode
   ```

3. **Follow Authentication Prompts**: The application will guide you through GitHub authentication.

### Method 2: Manual Installation

#### Global Installation
```bash
# Install globally
npm install -g @github/copilot-language-server

# Verify installation
copilot-language-server --version
```

#### Local Project Installation
```bash
# Navigate to your project directory
cd /path/to/your/project

# Install locally
npm install @github/copilot-language-server

# Verify installation
./node_modules/.bin/copilot-language-server --version
```

#### Using Native Binary
```bash
# After npm installation, test native binary
./node_modules/@github/copilot-language-server/native/darwin-arm64/copilot-language-server --version
```

### Method 3: Development Installation

For development and testing:

```bash
# Clone and build from source (if needed)
git clone https://github.com/github/copilot-language-server.git
cd copilot-language-server
npm install
npm run build
```

## Configuration Setup

### Basic Configuration

Create or update your SuperOpenCode configuration file:

```json
{
  "copilot": {
    "enable_copilot": true,
    "server_path": "auto",
    "use_native_binary": true,
    "replace_gopls": false,
    "auth_token": "",
    "chat_enabled": true,
    "completion_enabled": true,
    "auto_install": true
  }
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enable_copilot` | boolean | `false` | Enable GitHub Copilot language server |
| `server_path` | string | `"auto"` | Path to language server (`"auto"`, `"global"`, or explicit path) |
| `node_path` | string | `"node"` | Path to Node.js executable |
| `use_native_binary` | boolean | `true` | Use native binary instead of Node.js version |
| `replace_gopls` | boolean | `false` | Replace gopls entirely (vs. hybrid mode) |
| `auth_token` | string | `""` | GitHub authentication token (leave empty for interactive auth) |
| `chat_enabled` | boolean | `true` | Enable Copilot chat features |
| `completion_enabled` | boolean | `true` | Enable AI-powered completions |
| `auto_install` | boolean | `true` | Automatically install language server if missing |

### Advanced Configuration

```json
{
  "copilot": {
    "enable_copilot": true,
    "server_path": "/custom/path/to/copilot-language-server",
    "server_args": ["--stdio"],
    "environment": {
      "COPILOT_NODE_PATH": "/usr/bin/node",
      "COPILOT_LOG_LEVEL": "info"
    },
    "timeout": 30000,
    "retry_attempts": 3,
    "fallback_to_gopls": true
  }
}
```

## Authentication Setup

### Method 1: Interactive Authentication (Recommended)

1. **Start SuperOpenCode**: The application will prompt for authentication
2. **Follow Browser Prompts**: Complete GitHub OAuth flow
3. **Verify Access**: Check that Copilot features are working

### Method 2: GitHub CLI Authentication

```bash
# Login with GitHub CLI
gh auth login

# Generate token for Copilot
gh auth token --scopes copilot
```

### Method 3: Personal Access Token

1. **Generate Token**:
   - Go to https://github.com/settings/tokens
   - Click "Generate new token"
   - Select `copilot` scope
   - Copy the generated token

2. **Configure Token**:
   ```json
   {
     "copilot": {
       "auth_token": "ghp_your_token_here"
     }
   }
   ```

### Method 4: Environment Variable

```bash
export GITHUB_TOKEN=ghp_your_token_here
opencode
```

## Verification and Testing

### Basic Functionality Test

1. **Start SuperOpenCode**:
   ```bash
   opencode
   ```

2. **Check Server Status**:
   ```bash
   # In SuperOpenCode, check LSP status
   :lsp status
   ```

3. **Test Completions**:
   - Open a Go file
   - Start typing a function
   - Verify AI suggestions appear

### Diagnostic Commands

```bash
# Check installation paths
opencode --check-copilot

# Test language server directly
node ./node_modules/@github/copilot-language-server/dist/language-server.js --version

# Test native binary
./node_modules/@github/copilot-language-server/native/darwin-arm64/copilot-language-server --version
```

### Troubleshooting Installation

#### Common Issues

1. **Node.js Version Too Old**:
   ```bash
   # Update Node.js using nvm
   nvm install 20
   nvm use 20
   ```

2. **Permission Issues**:
   ```bash
   # Fix npm permissions
   npm config set prefix ~/.npm-global
   export PATH=~/.npm-global/bin:$PATH
   ```

3. **Missing GitHub Token**:
   - Ensure you have GitHub Copilot subscription
   - Check token has `copilot` scope
   - Verify token is not expired

4. **Path Detection Issues**:
   ```json
   {
     "copilot": {
       "server_path": "/explicit/path/to/copilot-language-server"
     }
   }
   ```

## Platform-Specific Instructions

### macOS

```bash
# Install Node.js via Homebrew
brew install node

# Install Copilot language server
npm install -g @github/copilot-language-server

# Verify native binary (Apple Silicon)
./node_modules/@github/copilot-language-server/native/darwin-arm64/copilot-language-server --version

# Verify native binary (Intel)
./node_modules/@github/copilot-language-server/native/darwin-x64/copilot-language-server --version
```

### Linux

```bash
# Install Node.js via package manager
sudo apt update
sudo apt install nodejs npm

# Or via NodeSource
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Install Copilot language server
npm install -g @github/copilot-language-server
```

### Windows

```powershell
# Install Node.js via Chocolatey
choco install nodejs

# Or download from nodejs.org

# Install Copilot language server
npm install -g @github/copilot-language-server

# Test native binary
.\node_modules\@github\copilot-language-server\native\win32-x64\copilot-language-server.exe --version
```

## Migration from gopls

### Gradual Migration (Recommended)

1. **Install Copilot alongside gopls**:
   ```json
   {
     "copilot": {
       "enable_copilot": true,
       "replace_gopls": false
     }
   }
   ```

2. **Test Copilot features** while keeping gopls active

3. **Complete migration** when ready:
   ```json
   {
     "copilot": {
       "replace_gopls": true
     }
   }
   ```

### Complete Replacement

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

## Uninstallation

### Remove Copilot Language Server

```bash
# Global uninstall
npm uninstall -g @github/copilot-language-server

# Local uninstall
npm uninstall @github/copilot-language-server
```

### Revert Configuration

```json
{
  "copilot": {
    "enable_copilot": false
  }
}
```

### Clean Credentials

```bash
# Remove stored tokens (varies by OS)
# macOS: Check Keychain Access
# Linux: Check ~/.config/gh/
# Windows: Check Windows Credential Manager
```

## Next Steps

After successful installation:

1. **Read Configuration Reference**: See `copilot-configuration-reference.md`
2. **Follow Migration Guide**: See `copilot-migration-guide.md`
3. **Explore API Features**: See `copilot-api-reference.md`
4. **Troubleshoot Issues**: See `copilot-troubleshooting.md`

## Support

- **GitHub Copilot Issues**: https://github.com/github/copilot-language-server/issues
- **SuperOpenCode Issues**: https://github.com/kirmad/superopencode/issues
- **Documentation**: See other guides in `docs/` folder