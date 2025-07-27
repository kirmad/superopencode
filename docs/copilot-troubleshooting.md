# GitHub Copilot Language Server Troubleshooting Guide

## Quick Diagnostics

### Basic Health Check

```bash
# Check Copilot status
opencode --check-copilot

# Verify LSP server status
opencode --lsp-status

# Test authentication
opencode --test-auth

# Show current configuration
opencode --show-config

# Enable debug logging
opencode --debug --log-level debug
```

### System Information

```bash
# Check system requirements
node --version        # Should be >= 20.8.0
npm --version         # Should be >= 8.0.0
which copilot-language-server

# Check GitHub authentication
gh auth status

# Verify Copilot subscription
gh api user/copilot_billing
```

## Common Issues

### 1. Installation Problems

#### Issue: "Copilot language server not found"

**Symptoms**:
- Error: "Server executable not found"
- SuperOpenCode fails to start Copilot
- LSP client reports server startup failure

**Diagnosis**:
```bash
# Check if server is installed
which copilot-language-server

# Check npm global packages
npm list -g --depth=0 | grep copilot

# Check local installation
ls ./node_modules/.bin/copilot-language-server
```

**Solutions**:

1. **Install globally**:
   ```bash
   npm install -g @github/copilot-language-server
   ```

2. **Install locally**:
   ```bash
   npm install @github/copilot-language-server
   ```

3. **Fix npm permissions**:
   ```bash
   npm config set prefix ~/.npm-global
   export PATH=~/.npm-global/bin:$PATH
   ```

4. **Use explicit path**:
   ```json
   {
     "copilot": {
       "server_path": "/absolute/path/to/copilot-language-server"
     }
   }
   ```

#### Issue: "npm install fails with permission errors"

**Symptoms**:
- "EACCES: permission denied" during installation
- Installation fails on macOS/Linux

**Solutions**:

1. **Use npm prefix**:
   ```bash
   mkdir ~/.npm-global
   npm config set prefix '~/.npm-global'
   echo 'export PATH=~/.npm-global/bin:$PATH' >> ~/.bashrc
   source ~/.bashrc
   ```

2. **Use nvm (recommended)**:
   ```bash
   curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
   nvm install 20
   nvm use 20
   npm install -g @github/copilot-language-server
   ```

3. **Fix ownership (Linux/macOS)**:
   ```bash
   sudo chown -R $(whoami) $(npm config get prefix)/{lib/node_modules,bin,share}
   ```

#### Issue: "Node.js version too old"

**Symptoms**:
- Error: "Node.js version 20.8.0 or later required"
- Server fails to start

**Solutions**:

1. **Update Node.js**:
   ```bash
   # Using nvm
   nvm install 20
   nvm use 20
   
   # Using Homebrew (macOS)
   brew upgrade node
   
   # Using package manager (Linux)
   curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
   sudo apt-get install -y nodejs
   ```

2. **Specify Node.js path**:
   ```json
   {
     "copilot": {
       "node_path": "/path/to/node20/bin/node"
     }
   }
   ```

### 2. Authentication Issues

#### Issue: "Authentication failed"

**Symptoms**:
- "GitHub authentication required" error
- Copilot features not working
- "Invalid credentials" messages

**Diagnosis**:
```bash
# Check GitHub CLI authentication
gh auth status

# Test GitHub API access
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user

# Check Copilot subscription
gh api user/copilot_billing
```

**Solutions**:

1. **Re-authenticate with GitHub CLI**:
   ```bash
   gh auth logout
   gh auth login --scopes copilot
   ```

2. **Generate new personal access token**:
   - Visit https://github.com/settings/tokens
   - Create token with `copilot` scope
   - Add to configuration:
   ```json
   {
     "copilot": {
       "auth_token": "ghp_your_new_token"
     }
   }
   ```

3. **Use environment variable**:
   ```bash
   export GITHUB_TOKEN=ghp_your_token
   export OPENCODE_COPILOT_AUTH_TOKEN=ghp_your_token
   ```

4. **Clear stored credentials**:
   ```bash
   # macOS: Clear keychain
   security delete-generic-password -s "GitHub Copilot"
   
   # Linux: Clear stored credentials
   rm -rf ~/.config/gh/hosts.yml
   
   # Windows: Use Credential Manager
   ```

#### Issue: "No Copilot subscription"

**Symptoms**:
- "Copilot subscription required" error
- Authentication succeeds but features don't work

**Solutions**:

1. **Verify subscription**:
   ```bash
   gh api user/copilot_billing
   ```

2. **Purchase/enable Copilot**:
   - Visit https://github.com/settings/copilot
   - Enable GitHub Copilot for your account

3. **Check organization settings** (if using organization account):
   - Contact organization admin
   - Verify Copilot is enabled for organization members

### 3. Performance Issues

#### Issue: "Completions are slow"

**Symptoms**:
- Completions take >5 seconds
- Typing becomes laggy
- High CPU usage

**Diagnosis**:
```bash
# Monitor resource usage
top -p $(pgrep copilot-language-server)

# Check completion performance
opencode --benchmark-completions

# Enable performance debugging
opencode --debug --performance-monitoring
```

**Solutions**:

1. **Adjust performance settings**:
   ```json
   {
     "copilot": {
       "performance": {
         "max_completion_time": 2000,
         "debounce_delay": 300,
         "max_parallel_requests": 3
       }
     }
   }
   ```

2. **Enable caching**:
   ```json
   {
     "copilot": {
       "cache_enabled": true,
       "cache_size": 200
     }
   }
   ```

3. **Use native binary**:
   ```json
   {
     "copilot": {
       "use_native_binary": true
     }
   }
   ```

4. **Reduce trigger sensitivity**:
   ```json
   {
     "copilot": {
       "completion_trigger_length": 3
     }
   }
   ```

#### Issue: "High memory usage"

**Symptoms**:
- Copilot process uses >500MB RAM
- System becomes unresponsive
- Out of memory errors

**Solutions**:

1. **Limit cache size**:
   ```json
   {
     "copilot": {
       "cache_size": 50,
       "performance": {
         "max_parallel_requests": 2
       }
     }
   }
   ```

2. **Restart server periodically**:
   ```bash
   # Add to cron for automatic restart
   0 */6 * * * killall copilot-language-server
   ```

3. **Monitor and debug**:
   ```bash
   # Check memory usage
   ps aux | grep copilot-language-server
   
   # Enable memory debugging
   export NODE_OPTIONS="--max-old-space-size=512"
   ```

### 4. Connectivity Issues

#### Issue: "Network connection failed"

**Symptoms**:
- "Unable to connect to GitHub" errors
- Completions not working
- Chat features unavailable

**Diagnosis**:
```bash
# Test GitHub connectivity
curl -I https://api.github.com

# Check DNS resolution
nslookup api.github.com

# Test with proxy (if applicable)
curl -x $HTTP_PROXY https://api.github.com
```

**Solutions**:

1. **Configure proxy**:
   ```json
   {
     "copilot": {
       "environment": {
         "HTTP_PROXY": "http://proxy.company.com:8080",
         "HTTPS_PROXY": "http://proxy.company.com:8080"
       }
     }
   }
   ```

2. **Add to allowed domains**:
   ```json
   {
     "copilot": {
       "security": {
         "allowed_domains": [
           "api.github.com",
           "copilot.github.com",
           "copilot-proxy.githubusercontent.com"
         ]
       }
     }
   }
   ```

3. **Check firewall settings**:
   - Ensure ports 443 and 80 are open
   - Allow `copilot-language-server` through firewall

### 5. Configuration Issues

#### Issue: "Invalid configuration"

**Symptoms**:
- "Configuration validation failed" error
- SuperOpenCode won't start
- Unexpected behavior

**Solutions**:

1. **Validate configuration**:
   ```bash
   opencode --validate-config
   ```

2. **Reset to defaults**:
   ```bash
   # Backup current config
   cp ~/.config/opencode/config.json ~/.config/opencode/config.json.backup
   
   # Remove Copilot config
   opencode --reset-copilot-config
   ```

3. **Check JSON syntax**:
   ```bash
   # Validate JSON
   cat ~/.config/opencode/config.json | jq .
   ```

4. **Use minimal configuration**:
   ```json
   {
     "copilot": {
       "enable_copilot": true
     }
   }
   ```

#### Issue: "gopls conflicts with Copilot"

**Symptoms**:
- Duplicate completions
- Conflicting diagnostics
- Performance issues

**Solutions**:

1. **Use replacement mode**:
   ```json
   {
     "copilot": {
       "replace_gopls": true
     },
     "lsp": {
       "disable_gopls": true
     }
   }
   ```

2. **Configure hybrid mode carefully**:
   ```json
   {
     "lsp": {
       "enable_hybrid_mode": true
     },
     "copilot": {
       "replace_gopls": false
     }
   }
   ```

3. **Prioritize Copilot completions**:
   ```json
   {
     "copilot": {
       "performance": {
         "max_parallel_requests": 5
       }
     }
   }
   ```

### 6. Feature-Specific Issues

#### Issue: "Chat not working"

**Symptoms**:
- Chat commands fail
- "Chat feature unavailable" error
- No response from chat

**Solutions**:

1. **Enable chat explicitly**:
   ```json
   {
     "copilot": {
       "chat_enabled": true
     }
   }
   ```

2. **Check server capabilities**:
   ```bash
   opencode --show-capabilities
   ```

3. **Test chat manually**:
   ```bash
   # In SuperOpenCode
   :copilot chat "Hello"
   ```

#### Issue: "No code suggestions"

**Symptoms**:
- No AI-powered completions
- Only basic LSP completions
- Empty completion list

**Solutions**:

1. **Enable completions**:
   ```json
   {
     "copilot": {
       "completion_enabled": true
     }
   }
   ```

2. **Check trigger length**:
   ```json
   {
     "copilot": {
       "completion_trigger_length": 1
     }
   }
   ```

3. **Verify file type support**:
   - Ensure file has proper extension (.go)
   - Check language ID detection

## Platform-Specific Issues

### macOS Issues

#### Issue: "Gatekeeper blocking execution"

**Solutions**:
```bash
# Allow execution
sudo xattr -rd com.apple.quarantine /path/to/copilot-language-server

# Or allow in Security & Privacy settings
```

#### Issue: "Rosetta 2 required" (Apple Silicon)

**Solutions**:
```bash
# Install Rosetta 2
sudo softwareupdate --install-rosetta

# Or use native ARM binary
```

### Linux Issues

#### Issue: "Missing dependencies"

**Solutions**:
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install curl gnupg

# CentOS/RHEL
sudo yum install curl gnupg
```

#### Issue: "SELinux blocking execution"

**Solutions**:
```bash
# Check SELinux status
sestatus

# Temporarily disable
sudo setenforce 0

# Or create policy for Copilot
```

### Windows Issues

#### Issue: "PowerShell execution policy"

**Solutions**:
```powershell
# Check current policy
Get-ExecutionPolicy

# Set policy for current user
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

#### Issue: "Windows Defender blocking"

**Solutions**:
- Add exception for `copilot-language-server.exe`
- Add exception for SuperOpenCode directory
- Temporarily disable real-time protection for installation

## Advanced Debugging

### Enable Debug Logging

```json
{
  "copilot": {
    "log_level": "debug",
    "environment": {
      "COPILOT_LOG_LEVEL": "debug",
      "NODE_OPTIONS": "--trace-warnings"
    }
  }
}
```

### Capture Network Traffic

```bash
# Monitor network requests
sudo tcpdump -i any host api.github.com

# Using mitmproxy for HTTPS
mitmproxy --mode transparent --showhost
```

### Memory and CPU Profiling

```bash
# Node.js memory profiling
node --inspect ./node_modules/@github/copilot-language-server/dist/language-server.js

# CPU profiling
perf record -g ./copilot-language-server
perf report
```

### LSP Communication Debugging

```json
{
  "copilot": {
    "environment": {
      "COPILOT_TRACE": "verbose"
    }
  },
  "lsp": {
    "trace_level": "verbose"
  }
}
```

## Log Analysis

### Common Log Patterns

```bash
# Authentication errors
grep -i "auth" ~/.config/opencode/logs/copilot.log

# Connection errors
grep -i "connection\|network\|timeout" ~/.config/opencode/logs/copilot.log

# Performance issues
grep -i "slow\|timeout\|performance" ~/.config/opencode/logs/copilot.log

# Server startup issues
grep -i "start\|init\|spawn" ~/.config/opencode/logs/copilot.log
```

### Log Locations

```bash
# SuperOpenCode logs
~/.config/opencode/logs/

# Copilot server logs
~/.config/opencode/logs/copilot.log

# System logs (Linux)
journalctl -u opencode

# System logs (macOS)
log show --predicate 'process == "opencode"'
```

## Recovery Procedures

### Emergency Fallback

```bash
# Quickly disable Copilot
export OPENCODE_COPILOT_ENABLE_COPILOT=false
opencode

# Or use command line
opencode --disable-copilot
```

### Complete Reset

```bash
# Stop all processes
killall opencode copilot-language-server

# Remove configuration
mv ~/.config/opencode/config.json ~/.config/opencode/config.json.backup

# Remove cache
rm -rf ~/.cache/opencode/

# Remove logs
rm -rf ~/.config/opencode/logs/

# Restart with defaults
opencode
```

### Reinstall Copilot

```bash
# Uninstall
npm uninstall -g @github/copilot-language-server

# Clear npm cache
npm cache clean --force

# Reinstall
npm install -g @github/copilot-language-server

# Verify
copilot-language-server --version
```

## Getting Help

### Before Asking for Help

1. **Gather information**:
   ```bash
   opencode --version
   opencode --check-copilot
   opencode --system-info
   node --version
   npm --version
   ```

2. **Reproduce the issue**:
   - Minimal configuration
   - Step-by-step reproduction
   - Expected vs. actual behavior

3. **Check logs**:
   - Enable debug logging
   - Capture relevant log excerpts
   - Include timestamps

### Support Channels

1. **GitHub Issues**: https://github.com/kirmad/superopencode/issues
2. **GitHub Copilot Support**: https://support.github.com/
3. **Community Forums**: Check project README for links

### Issue Report Template

```markdown
## Environment
- OS: [e.g., macOS 14.0, Ubuntu 22.04]
- Node.js version: [e.g., 20.8.0]
- SuperOpenCode version: [e.g., 1.0.0]
- Copilot server version: [e.g., 1.0.0]

## Configuration
```json
{
  "copilot": {
    // Include relevant config
  }
}
```

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What you expected to happen

## Actual Behavior
What actually happened

## Logs
```
Include relevant log excerpts
```

## Additional Context
Any other relevant information
```

## Prevention Tips

1. **Keep software updated**:
   ```bash
   npm update -g @github/copilot-language-server
   npm update -g npm
   ```

2. **Regular configuration validation**:
   ```bash
   opencode --validate-config
   ```

3. **Monitor performance**:
   ```bash
   opencode --performance-report
   ```

4. **Backup configurations**:
   ```bash
   cp ~/.config/opencode/config.json ~/.config/opencode/config.json.$(date +%Y%m%d)
   ```

5. **Use minimal configurations** for testing:
   ```json
   {
     "copilot": {
       "enable_copilot": true,
       "fallback_to_gopls": true
     }
   }
   ```

Remember: Most issues can be resolved by ensuring proper installation, authentication, and configuration. When in doubt, start with minimal settings and gradually enable features while testing each step.